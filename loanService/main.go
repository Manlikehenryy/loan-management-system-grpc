package main

import (
	"fmt"
	"log"
	"net"

	"github.com/manlikehenryy/loan-management-system-grpc/loanService/configs"
	"github.com/manlikehenryy/loan-management-system-grpc/loanService/database"
	pb "github.com/manlikehenryy/loan-management-system-grpc/loanService/loan"
	"github.com/manlikehenryy/loan-management-system-grpc/loanService/service"

	"google.golang.org/grpc"
)

func main() {

	// Start gRPC server for Loan service
	database.Connect()

	// Retrieve the port from the config
	port := configs.Env.PORT
	if port == "" {
		port = "50052" // Set a default port if not specified
	}

	// Start gRPC server for Loan service
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(service.TokenInterceptor))
	pb.RegisterLoanServiceServer(s, service.NewLoanServiceServer())

	fmt.Printf("Loan Service running on port %s...", port)
	log.Fatal(s.Serve(lis))
}
