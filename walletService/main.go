package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/manlikehenryy/loan-management-system-grpc/walletService/configs"
	"github.com/manlikehenryy/loan-management-system-grpc/walletService/database"
	pb "github.com/manlikehenryy/loan-management-system-grpc/walletService/wallet"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Wallet struct
type Wallet struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	UserID  primitive.ObjectID `bson:"userId,omitempty"`
	Balance float32            `bson:"balance,omitempty"`
}

// WalletServiceServer struct to implement gRPC functions
type WalletServiceServer struct {
	pb.UnimplementedWalletServiceServer
}

// CreateWallet creates a wallet for a user
func (s *WalletServiceServer) CreateWallet(ctx context.Context, req *pb.CreateWalletRequest) (*pb.CreateWalletResponse, error) {
	walletsCollection := database.DB.Collection("wallets")

	userId, err := primitive.ObjectIDFromHex(req.GetUserId())
	if err != nil {
		return &pb.CreateWalletResponse{Message: "Invalid user ID", Status: false, StatusCode: http.StatusBadRequest}, nil
	}

	
	wallet := Wallet{UserID: userId, Balance: 0}

	_, err = walletsCollection.InsertOne(context.Background(), wallet)
	if err != nil {
		return &pb.CreateWalletResponse{Message: "Wallet creation failed", Status: false, StatusCode: http.StatusInternalServerError}, nil
	}
	return &pb.CreateWalletResponse{Message: "Wallet created successfully", Status: true, StatusCode: http.StatusOK}, nil
}

func (s *WalletServiceServer) CreditWallet(ctx context.Context, req *pb.CreditWalletRequest) (*pb.CreditWalletResponse, error) {
	walletsCollection := database.DB.Collection("wallets")

	userId, err := primitive.ObjectIDFromHex(req.GetUserId())
	if err != nil {
		return &pb.CreditWalletResponse{Message: "Invalid user ID", Status: false, StatusCode: http.StatusBadRequest}, nil
	}

	url, err := walletsCollection.UpdateOne(
		context.TODO(),
		bson.M{"userId": userId},
		bson.M{
			"$inc": bson.M{"Balance": req.GetAmount()},
		},
	)

	if err != nil {
		log.Println("Database error:", err)
		return &pb.CreditWalletResponse{Message: "Failed to credit wallet", Status: false, StatusCode: http.StatusInternalServerError}, nil
	}

	if url.MatchedCount == 0 {
		return &pb.CreditWalletResponse{Message: "Wallet not found", Status: false, StatusCode: http.StatusNotFound}, nil
	}
	return &pb.CreditWalletResponse{Message: "Wallet credited successfully", Status: true, StatusCode: http.StatusOK}, nil
}

func tokenInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return nil, status.Errorf(codes.Unauthenticated, "no metadata provided")
    }

    // Get the "authorization" metadata value
    authHeader := md["authorization"]
    if len(authHeader) == 0 {
        return nil, status.Errorf(codes.Unauthenticated, "no auth token provided")
    }

    // Extract the token (remove "Bearer " prefix)
    token := authHeader[0]
    if len(token) < 7 || token[:7] != "Bearer " {
        return nil, status.Errorf(codes.Unauthenticated, "invalid auth token format")
    }
    actualToken := token[7:]

  
    if actualToken != configs.Env.TOKEN {
        return nil, status.Errorf(codes.Unauthenticated, "invalid auth token")
    }

    

    // Continue processing the request
    return handler(ctx, req)
}

func main() {

	// Start gRPC server for Loan service
	database.Connect()

	// Retrieve the port from the config
	port := configs.Env.PORT
	if port == "" {
		port = "50053" // Set a default port if not specified
	}

	// Start gRPC server for Loan service
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(tokenInterceptor))
	// s := grpc.NewServer()
	pb.RegisterWalletServiceServer(s, &WalletServiceServer{}) // Register WalletServiceServer

	fmt.Println("Wallet Service running on port 50053...")
	log.Fatal(s.Serve(lis))
}
