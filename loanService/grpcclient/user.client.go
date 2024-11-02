package grpcclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	userPb "github.com/manlikehenryy/loan-management-system-grpc/loanService/user" // Import your generated proto package
)

// NewUserServiceClient initializes a new UserServiceClient with context
func NewUserServiceClient(ctx context.Context) (userPb.UserServiceClient, func(), error) {

	// Establish the connection to the UserService
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	// Define a cleanup function to close the connection
	cleanup := func() {
		conn.Close()
	}

	return userPb.NewUserServiceClient(conn), cleanup, nil
}

func NewAuthContext(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}
