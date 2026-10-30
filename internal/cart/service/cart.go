package service

import (
	"context"
	"errors"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	pb "OzonHW2/api/product/v1"
	"OzonHW2/internal/cart/model"
	"OzonHW2/internal/cart/repository"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Cart struct {
	products  pb.ProductServiceClient
	carts     *repository.Cart
	additions atomic.Uint64
}

var ErrEmptyCart = errors.New("cart is empty")

func New(products pb.ProductServiceClient, carts *repository.Cart) *Cart {
	return &Cart{products: products, carts: carts}
}

func (s *Cart) Add(ctx context.Context, userID, sku int64, count uint32) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	product, err := s.products.GetProduct(ctx, &pb.GetProductRequest{Sku: sku})
	if err != nil {
		return err
	}
	if product.Price < 0 {
		return status.Error(codes.Internal, "invalid product price")
	}
	item := model.Item{SKU: sku, Name: product.Name, Price: uint32(product.Price)}
	if !s.carts.Add(userID, sku, item, count) {
		return status.Error(codes.OutOfRange, "count overflow")
	}
	s.additions.Add(1)
	return nil
}

func (s *Cart) Additions() uint64 {
	return s.additions.Load()
}

func (s *Cart) Get(ctx context.Context, userID int64) (model.Cart, error) {
	items := s.carts.Items(userID)
	if len(items) == 0 {
		return model.Cart{}, ErrEmptyCart
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].SKU < items[j].SKU
	})

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	result := model.Cart{Items: make([]model.Item, len(items))}
	errs := make([]error, len(items))
	var group sync.WaitGroup
	for index, item := range items {
		group.Add(1)
		go func(index int, item model.Item) {
			defer group.Done()
			product, err := s.products.GetProduct(ctx, &pb.GetProductRequest{Sku: item.SKU})
			if err != nil {
				errs[index] = err
				return
			}
			if product.Price < 0 {
				errs[index] = status.Error(codes.Internal, "invalid product price")
				return
			}
			result.Items[index] = model.Item{
				SKU: item.SKU, Name: product.Name, Count: item.Count, Price: uint32(product.Price),
			}
		}(index, item)
	}
	group.Wait()
	for _, err := range errs {
		if err != nil {
			return model.Cart{}, err
		}
	}
	var total uint64
	for _, item := range result.Items {
		total += uint64(item.Price) * uint64(item.Count)
		if total > math.MaxUint32 {
			return model.Cart{}, status.Error(codes.Internal, "total price overflow")
		}
	}
	result.TotalPrice = uint32(total)
	return result, nil
}
