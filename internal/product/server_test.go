package product

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	pb "OzonHW2/api/product/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetProduct(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		sku  int64
		want codes.Code
	}{
		{"found", 1, codes.OK},
		{"missing", 2, codes.NotFound},
		{"invalid", 0, codes.InvalidArgument},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "products.json")
			require.NoError(t, os.WriteFile(path, []byte(`{"1":{"name":"one","price":10}}`), 0600))
			service, err := Load(path)
			require.NoError(t, err)
			product, err := service.GetProduct(context.Background(), &pb.GetProductRequest{Sku: tc.sku})
			require.Equal(t, tc.want, status.Code(err))
			if tc.want == codes.OK {
				require.Equal(t, int32(10), product.Price)
			}
		})
	}
}

func TestInvalidCatalog(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, data string }{
		{"invalid json", `broken`},
		{"invalid sku", `{"0":{"price":1}}`},
		{"invalid price", `{"1":{"price":2147483648}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "products.json")
			require.NoError(t, os.WriteFile(path, []byte(tc.data), 0600))
			_, err := Load(path)
			require.Error(t, err)
		})
	}
}
