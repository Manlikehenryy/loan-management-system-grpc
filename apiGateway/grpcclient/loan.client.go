package grpcclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/configs"
	loanPb "github.com/manlikehenryy/loan-management-system-grpc/apiGateway/loan" // Import your generated proto package
)

// NewLoanServiceClient initializes a new LoanServiceClient with context
func NewLoanServiceClient(ctx context.Context) (loanPb.LoanServiceClient, func(), error) {

	// Establish the connection to the LoanService
	conn, err := grpc.NewClient(configs.Env.LOAN_SERVICE_URL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	// Define a cleanup function to close the connection
	cleanup := func() {
		conn.Close()
	}

	return loanPb.NewLoanServiceClient(conn), cleanup, nil
}
