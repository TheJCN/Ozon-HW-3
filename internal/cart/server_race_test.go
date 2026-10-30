//go:build race

package cart

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	"OzonHW2/internal/cart/repository"
	"OzonHW2/internal/cart/service"
	"github.com/stretchr/testify/require"
)

func TestConcurrentAdds(t *testing.T) {
	t.Parallel()
	service := New(fakeProducts{get: normalProduct})
	handler := service.Handler()
	const workers = 50
	var group sync.WaitGroup
	for i := 0; i < workers; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			response := send(handler, http.MethodPost, "/user/1/cart/1", `{"count":1}`)
			if response.Code != http.StatusNoContent {
				t.Errorf("add returned %d", response.Code)
			}
		}()
	}
	group.Wait()
	response := send(handler, http.MethodGet, "/user/1/cart", "")
	require.Equal(t, http.StatusOK, response.Code)
	var cart Cart
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &cart))
	require.Equal(t, uint32(workers), cart.Items[0].Count)
}

func TestAtomicAdditions(t *testing.T) {
	t.Parallel()
	logic := service.New(fakeProducts{get: normalProduct}, repository.New())
	const workers = 50
	var group sync.WaitGroup
	for i := 0; i < workers; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if err := logic.Add(context.Background(), 1, 1, 1); err != nil {
				t.Errorf("add: %v", err)
			}
		}()
	}
	group.Wait()
	require.Equal(t, uint64(workers), logic.Additions())
}
