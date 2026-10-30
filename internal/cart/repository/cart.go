package repository

import (
	"math"
	"sync"

	"OzonHW2/internal/cart/model"
)

type Cart struct {
	mu    sync.RWMutex
	carts map[int64]map[int64]model.Item
}

func New() *Cart {
	return &Cart{carts: make(map[int64]map[int64]model.Item)}
}

func (c *Cart) Add(userID, sku int64, item model.Item, count uint32) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	items := c.carts[userID]
	if items == nil {
		items = make(map[int64]model.Item)
		c.carts[userID] = items
	}
	current := items[sku]
	if count > math.MaxUint32-current.Count {
		return false
	}
	item.Count = current.Count + count
	items[sku] = item
	return true
}

func (c *Cart) Items(userID int64) []model.Item {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stored := c.carts[userID]
	items := make([]model.Item, 0, len(stored))
	for _, item := range stored {
		items = append(items, item)
	}
	return items
}
