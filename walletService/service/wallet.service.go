package service

import (
    "context"
    "log"
    "net/http"

    "github.com/manlikehenryy/loan-management-system-grpc/walletService/database"
    pb "github.com/manlikehenryy/loan-management-system-grpc/walletService/wallet"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Wallet struct {
    ID      primitive.ObjectID `bson:"_id,omitempty"`
    UserID  primitive.ObjectID `bson:"userId,omitempty"`
    Balance float32            `bson:"balance,omitempty"`
}

type WalletServiceServer struct {
    pb.UnimplementedWalletServiceServer
}

func NewWalletServiceServer() *WalletServiceServer {
    return &WalletServiceServer{}
}

func (s *WalletServiceServer) CreateWallet(ctx context.Context, req *pb.CreateWalletRequest) (*pb.CreateWalletResponse, error) {
    walletsCollection := database.GetCollection("wallets")

    userID, err := primitive.ObjectIDFromHex(req.GetUserId())
    if err != nil {
        return createWalletErrorResponse("Invalid user ID", http.StatusBadRequest), nil
    }

    wallet := Wallet{UserID: userID, Balance: 0}
    _, err = walletsCollection.InsertOne(ctx, wallet)
    if err != nil {
        return createWalletErrorResponse("Wallet creation failed", http.StatusInternalServerError), nil
    }
    return createWalletSuccessResponse("Wallet created successfully", http.StatusOK), nil
}

func (s *WalletServiceServer) CreditWallet(ctx context.Context, req *pb.CreditWalletRequest) (*pb.CreditWalletResponse, error) {
    walletsCollection := database.GetCollection("wallets")

    userID, err := primitive.ObjectIDFromHex(req.GetUserId())
    if err != nil {
        return creditWalletErrorResponse("Invalid user ID", http.StatusBadRequest), nil
    }

    updateResult, err := walletsCollection.UpdateOne(
        ctx,
        bson.M{"userId": userID},
        bson.M{"$inc": bson.M{"balance": req.GetAmount()}},
    )
    if err != nil {
        log.Println("Database error:", err)
        return creditWalletErrorResponse("Failed to credit wallet", http.StatusInternalServerError), nil
    }
    if updateResult.MatchedCount == 0 {
        return creditWalletErrorResponse("Wallet not found", http.StatusNotFound), nil
    }

    return creditWalletSuccessResponse("Wallet credited successfully", http.StatusOK), nil
}

// Success and Error Response functions for CreateWallet
func createWalletSuccessResponse(message string, statusCode int) *pb.CreateWalletResponse {
    return &pb.CreateWalletResponse{Message: message, Status: true, StatusCode: int32(statusCode)}
}

func createWalletErrorResponse(message string, statusCode int) *pb.CreateWalletResponse {
    return &pb.CreateWalletResponse{Message: message, Status: false, StatusCode: int32(statusCode)}
}

// Success and Error Response functions for CreditWallet
func creditWalletSuccessResponse(message string, statusCode int) *pb.CreditWalletResponse {
    return &pb.CreditWalletResponse{Message: message, Status: true, StatusCode: int32(statusCode)}
}

func creditWalletErrorResponse(message string, statusCode int) *pb.CreditWalletResponse {
    return &pb.CreditWalletResponse{Message: message, Status: false, StatusCode: int32(statusCode)}
}
