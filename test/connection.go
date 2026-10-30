//go:build api

package test

import (
	"net/http"
	"time"

	pb "OzonHW2/api/product/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func connectProduct() (pb.ProductServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return pb.NewProductServiceClient(conn), conn, nil
}

func httpClient() *http.Client {
	return &http.Client{Timeout: 5 * time.Second}
}
