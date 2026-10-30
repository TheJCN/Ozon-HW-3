//go:build api

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	pb "OzonHW2/api/product/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLiveAPI(t *testing.T) {
	client, conn, err := connectProduct()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	product, err := client.GetProduct(ctx, &pb.GetProductRequest{Sku: 1076963})
	require.NoError(t, err)
	require.Equal(t, int64(1076963), product.Sku)
	_, err = client.GetProduct(ctx, &pb.GetProductRequest{Sku: 0})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	_, err = client.GetProduct(ctx, &pb.GetProductRequest{Sku: 999999999})
	require.Equal(t, codes.NotFound, status.Code(err))

	userID := time.Now().UnixNano()
	base := fmt.Sprintf("http://localhost:8080/user/%d/cart", userID)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/1076963", strings.NewReader(`{"count":2}`))
	require.NoError(t, err)
	response, err := httpClient().Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, response.StatusCode)
	require.NoError(t, response.Body.Close())
	request, err = http.NewRequestWithContext(ctx, http.MethodGet, base, nil)
	require.NoError(t, err)
	response, err = httpClient().Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	var cart struct {
		Items []struct {
			SKU   int64  `json:"sku"`
			Count uint32 `json:"count"`
		} `json:"items"`
		TotalPrice uint32 `json:"total_price"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&cart))
	require.NoError(t, response.Body.Close())
	require.Len(t, cart.Items, 1)
	require.Equal(t, int64(1076963), cart.Items[0].SKU)
	require.Equal(t, uint32(2), cart.Items[0].Count)
	require.Equal(t, uint32(product.Price)*2, cart.TotalPrice)

	request, err = http.NewRequestWithContext(ctx, http.MethodPost, base+"/999999999", strings.NewReader(`{"count":1}`))
	require.NoError(t, err)
	response, err = httpClient().Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusPreconditionFailed, response.StatusCode)
	require.NoError(t, response.Body.Close())

	request, err = http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/user/0/cart", nil)
	require.NoError(t, err)
	response, err = httpClient().Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.StatusCode)
	require.NoError(t, response.Body.Close())
}
