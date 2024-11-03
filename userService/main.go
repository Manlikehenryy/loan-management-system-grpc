package main

import (
	"fmt"
	"log"
	"net"

	"github.com/manlikehenryy/loan-management-system-grpc/userService/configs"
	"github.com/manlikehenryy/loan-management-system-grpc/userService/database"
	"github.com/manlikehenryy/loan-management-system-grpc/userService/helpers"
	"github.com/manlikehenryy/loan-management-system-grpc/userService/service"
	pb "github.com/manlikehenryy/loan-management-system-grpc/userService/user" // Import generated protobuf code
	"google.golang.org/grpc"
)

func main() {
	helpers.Initialize()
	// Connect to the database
	database.Connect()

	// Retrieve the port from the config
	port := configs.Env.PORT
	if port == "" {
		port = "50051" // Set a default port if not specified
	}

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(service.TokenInterceptor))
	pb.RegisterUserServiceServer(s, service.NewUserServiceServer()) // Register UserServiceServer
	fmt.Printf("User Service running on port %s...", port)
	log.Fatal(s.Serve(lis))
}
