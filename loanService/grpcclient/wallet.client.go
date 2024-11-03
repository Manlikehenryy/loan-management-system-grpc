package grpcclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/manlikehenryy/loan-management-system-grpc/loanService/configs"
	walletPb "github.com/manlikehenryy/loan-management-system-grpc/loanService/wallet" // Import your generated proto package
)

// NewWalletServiceClient initializes a new WalletServiceClient with context
func NewWalletServiceClient(ctx context.Context) (walletPb.WalletServiceClient, func(), error) {

	// Establish the connection to the UserService
	conn, err := grpc.NewClient(configs.Env.WALLET_SERVICE_URL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	// Define a cleanup function to close the connection
	cleanup := func() {
		conn.Close()
	}

	return walletPb.NewWalletServiceClient(conn), cleanup, nil
}

