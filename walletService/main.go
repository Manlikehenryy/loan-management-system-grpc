package main

import (
    "fmt"
    "log"
    "net"

    "github.com/manlikehenryy/loan-management-system-grpc/walletService/configs"
    "github.com/manlikehenryy/loan-management-system-grpc/walletService/database"
    "github.com/manlikehenryy/loan-management-system-grpc/walletService/service"
    pb "github.com/manlikehenryy/loan-management-system-grpc/walletService/wallet"
    "google.golang.org/grpc"
)

func main() {
    database.Connect()

    port := configs.Env.PORT
    if port == "" {
        port = "50053"
    }

    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer(grpc.UnaryInterceptor(service.TokenInterceptor))
    pb.RegisterWalletServiceServer(grpcServer, service.NewWalletServiceServer())

    fmt.Printf("Wallet Service running on port %s...\n", port)
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
