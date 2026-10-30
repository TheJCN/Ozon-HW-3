package cart

import (
	"net/http"
	"testing"
)

func BenchmarkAddToCart(b *testing.B) {
	server := New(fakeProducts{get: normalProduct})
	handler := server.Handler()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response := send(handler, http.MethodPost, "/user/1/cart/1", `{"count":1}`)
		if response.Code != http.StatusNoContent {
			b.Fatalf("add returned %d", response.Code)
		}
	}
}

func BenchmarkGetCart(b *testing.B) {
	server := New(fakeProducts{get: normalProduct})
	handler := server.Handler()
	if response := send(handler, http.MethodPost, "/user/1/cart/1", `{"count":1}`); response.Code != http.StatusNoContent {
		b.Fatalf("setup returned %d", response.Code)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response := send(handler, http.MethodGet, "/user/1/cart", "")
		if response.Code != http.StatusOK {
			b.Fatalf("get returned %d", response.Code)
		}
	}
}
