package cart

import (
	pb "OzonHW2/api/product/v1"
	"OzonHW2/internal/cart/handler"
	"OzonHW2/internal/cart/model"
	"OzonHW2/internal/cart/repository"
	"OzonHW2/internal/cart/service"
)

type Item = model.Item
type AddRequest = model.AddRequest
type Cart = model.Cart
type Server = handler.Server

func New(products pb.ProductServiceClient) *Server {
	carts := repository.New()
	logic := service.New(products, carts)
	return handler.New(logic)
}
