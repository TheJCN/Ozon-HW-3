package main

import (
	"log"
	"net"
	"net/http"
	"time"

	pb "OzonHW2/api/product/v1"
	"OzonHW2/internal/cart"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	credentials := insecure.NewCredentials()
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(credentials))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	productClient := pb.NewProductServiceClient(conn)
	service := cart.New(productClient)
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cart listening on :8080")
	if err := serve(listener, service.Handler()); err != nil {
		log.Fatal(err)
	}
}

func serve(listener net.Listener, handler http.Handler) error {
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return server.Serve(listener)
}
