package product

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	pb "OzonHW2/api/product/v1"
)

func BenchmarkGetProduct(b *testing.B) {
	path := filepath.Join(b.TempDir(), "products.json")
	if err := os.WriteFile(path, []byte(`{"1":{"name":"one","price":10}}`), 0600); err != nil {
		b.Fatal(err)
	}
	server, err := Load(path)
	if err != nil {
		b.Fatal(err)
	}
	request := &pb.GetProductRequest{Sku: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := server.GetProduct(context.Background(), request); err != nil {
			b.Fatal(err)
		}
	}
}
