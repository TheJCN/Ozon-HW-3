package cart

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	pb "OzonHW2/api/product/v1"
	"OzonHW2/internal/cart/model"
	"OzonHW2/internal/cart/repository"
	"OzonHW2/internal/cart/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeProducts struct {
	get func(context.Context, int64) (*pb.GetProductResponse, error)
}

func (f fakeProducts) GetProduct(ctx context.Context, request *pb.GetProductRequest, _ ...grpc.CallOption) (*pb.GetProductResponse, error) {
	return f.get(ctx, request.Sku)
}

func normalProduct(_ context.Context, sku int64) (*pb.GetProductResponse, error) {
	if sku == 99 {
		return nil, status.Error(codes.NotFound, "missing")
	}
	return &pb.GetProductResponse{Sku: sku, Name: "product", Price: int32(sku * 10)}, nil
}

func send(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	handler.ServeHTTP(response, request)
	return response
}

func TestCartRequests(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, method, path, body string
		want                     int
	}{
		{"empty cart", "GET", "/user/1/cart", "", 404},
		{"invalid user", "GET", "/user/0/cart", "", 400},
		{"invalid sku", "POST", "/user/1/cart/0", `{"count":1}`, 400},
		{"invalid count", "POST", "/user/1/cart/1", `{"count":0}`, 400},
		{"invalid json", "POST", "/user/1/cart/1", `{`, 400},
		{"missing product", "POST", "/user/1/cart/99", `{"count":1}`, 412},
		{"add", "POST", "/user/1/cart/1", `{"count":1}`, 204},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			service := New(fakeProducts{get: normalProduct})
			response := send(service.Handler(), tc.method, tc.path, tc.body)
			require.Equal(t, tc.want, response.Code)
		})
	}
}

func TestGetCartJSON(t *testing.T) {
	t.Parallel()
	service := New(fakeProducts{get: normalProduct})
	handler := service.Handler()
	require.Equal(t, 204, send(handler, "POST", "/user/1/cart/2", `{"count":1}`).Code)
	require.Equal(t, 204, send(handler, "POST", "/user/1/cart/1", `{"count":2}`).Code)
	response := send(handler, "GET", "/user/1/cart", "")
	require.Equal(t, 200, response.Code)
	var body struct {
		Items []struct {
			SKU   int64  `json:"sku"`
			Count uint32 `json:"count"`
			Price uint32 `json:"price"`
		} `json:"items"`
		TotalPrice uint32 `json:"total_price"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Len(t, body.Items, 2)
	require.Equal(t, int64(1), body.Items[0].SKU)
	require.Equal(t, uint32(2), body.Items[0].Count)
	require.Equal(t, uint32(10), body.Items[0].Price)
	require.Equal(t, int64(2), body.Items[1].SKU)
	require.Equal(t, uint32(20), body.Items[1].Price)
	require.Equal(t, uint32(40), body.TotalPrice)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &raw))
	require.IsType(t, float64(0), raw["total_price"])
	require.IsType(t, float64(0), raw["items"].([]any)[0].(map[string]any)["sku"])
	require.IsType(t, float64(0), raw["items"].([]any)[0].(map[string]any)["price"])
}

func TestGetCartMissingProduct(t *testing.T) {
	t.Parallel()
	var missing bool
	var mu sync.Mutex
	fake := fakeProducts{get: func(ctx context.Context, sku int64) (*pb.GetProductResponse, error) {
		mu.Lock()
		gone := missing
		mu.Unlock()
		if gone {
			return nil, status.Error(codes.NotFound, "missing")
		}
		return normalProduct(ctx, sku)
	}}
	service := New(fake)
	handler := service.Handler()
	require.Equal(t, 204, send(handler, "POST", "/user/1/cart/1", `{"count":1}`).Code)
	mu.Lock()
	missing = true
	mu.Unlock()
	response := send(handler, "GET", "/user/1/cart", "")
	require.Equal(t, 412, response.Code)
	require.NotContains(t, response.Body.String(), `"items"`)
}

func TestGetCartRequestsProductsInParallel(t *testing.T) {
	t.Parallel()
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	fake := fakeProducts{get: func(ctx context.Context, sku int64) (*pb.GetProductResponse, error) {
		if sku == 1 || sku == 2 {
			select {
			case started <- struct{}{}:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			select {
			case <-release:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return normalProduct(ctx, sku)
	}}
	carts := repository.New()
	require.True(t, carts.Add(1, 1, model.Item{SKU: 1}, 1))
	require.True(t, carts.Add(1, 2, model.Item{SKU: 2}, 1))
	logic := service.New(fake, carts)
	done := make(chan error, 1)
	go func() {
		_, err := logic.Get(context.Background(), 1)
		done <- err
	}()
	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			close(release)
			t.Fatal("product requests did not start in parallel")
		}
	}
	close(release)
	require.NoError(t, <-done)
}
