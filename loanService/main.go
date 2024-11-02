package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/manlikehenryy/loan-management-system-grpc/loanService/configs"
	"github.com/manlikehenryy/loan-management-system-grpc/loanService/database"
	"github.com/manlikehenryy/loan-management-system-grpc/loanService/grpcclient"
	pb "github.com/manlikehenryy/loan-management-system-grpc/loanService/loan"
	userPb "github.com/manlikehenryy/loan-management-system-grpc/loanService/user"
	walletPb "github.com/manlikehenryy/loan-management-system-grpc/loanService/wallet"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Loan struct
type Loan struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	UserID           primitive.ObjectID `bson:"userId,omitempty"`
	Amount           float32            `bson:"amount,omitempty"`
	Status           string             `bson:"status,omitempty"` // "approved", "rejected"
	ApprovedBy       primitive.ObjectID `bson:"approvedBy,omitempty"`
	RejectedBy       primitive.ObjectID `bson:"rejectedBy,omitempty"`
	ApprovedAmount   float32            `bson:"approvedAmount,omitempty"`
	Tenure           int32              `bson:"tenure,omitempty"`
	MonthlyRepayment float32            `bson:"monthlyRepayment,omitempty"`
	EffectiveDate    string             `bson:"effectiveDate,omitempty"`
	ExpiryDate       string             `bson:"expiryDate,omitempty"`
	AmountPaid       float32            `bson:"amountPaid,omitempty"`
}

// LoanServiceServer struct to implement gRPC functions
type LoanServiceServer struct {
	pb.UnimplementedLoanServiceServer
}

func (s *LoanServiceServer) ApplyLoan(ctx context.Context, req *pb.ApplyLoanRequest) (*pb.ApplyLoanResponse, error) {
	loansCollection := database.DB.Collection("loans")

	userId, err := primitive.ObjectIDFromHex(req.GetUserId())
	if err != nil {
		return &pb.ApplyLoanResponse{Message: "Invalid user ID", Status: false, StatusCode: http.StatusBadRequest}, nil
	}

	loan := Loan{UserID: userId, Amount: req.GetAmount(), Status: "pending"}

	result, err_ := loansCollection.InsertOne(context.Background(), loan)
	if err_ != nil {
		return &pb.ApplyLoanResponse{Message: "Loan application failed", Status: false, StatusCode: http.StatusInternalServerError}, nil
	}
	return &pb.ApplyLoanResponse{Message: "Loan application submitted", Status: true, StatusCode: http.StatusCreated, LoandId: result.InsertedID.(primitive.ObjectID).Hex()}, nil
}

func (s *LoanServiceServer) ApproveLoan(ctx context.Context, req *pb.ApproveLoanRequest) (*pb.ApproveLoanResponse, error) {

	// Initialize the gRPC client
	userServiceClient, cleanup, err := grpcclient.NewUserServiceClient(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to UserService: %v", err)
		return &pb.ApproveLoanResponse{Message: "Internal service error", Status: false, StatusCode: http.StatusInternalServerError}, nil
	}
	defer cleanup()

	// Set up the context with authorization metadata
	c := grpcclient.NewAuthContext(context.Background(), configs.Env.TOKEN)

	isAdminReq := &userPb.IsAdminRequest{
		UserId: req.UserId,
	}
	isAdminResp, err_ := userServiceClient.IsAdmin(c, isAdminReq)

	if isAdminResp == nil {
		log.Println("Error in IsAdmin call:", err)
		return &pb.ApproveLoanResponse{Message: "Unexpected service response", Status: false, StatusCode: http.StatusInternalServerError}, nil
	}

	if err_ != nil || !isAdminResp.Valid {

		return &pb.ApproveLoanResponse{Message: isAdminResp.Message, Status: false, StatusCode: isAdminResp.StatusCode}, nil
	}

	loansCollection := database.DB.Collection("loans")

	loanId, err := primitive.ObjectIDFromHex(req.GetLoanId())
	if err != nil {
		return &pb.ApproveLoanResponse{Message: "Invalid loan ID", Status: false, StatusCode: http.StatusBadRequest}, nil
	}

	var existingLoan Loan
	err = loansCollection.FindOne(context.Background(), bson.M{"_id": loanId}).Decode(&existingLoan)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &pb.ApproveLoanResponse{Message: "Loan not found", Status: false, StatusCode: http.StatusNotFound}, nil
		} else {
			return &pb.ApproveLoanResponse{Message: "Failed to approve loan", Status: false, StatusCode: http.StatusInternalServerError}, nil
		}
	}

	userId, err := primitive.ObjectIDFromHex(req.GetUserId())
	if err != nil {
		return &pb.ApproveLoanResponse{Message: "Invalid user ID", Status: false, StatusCode: http.StatusBadRequest}, nil
	}

	update := bson.M{
		"$set": bson.M{
			"approvedBy":       userId,
			"approvedAmount":   req.GetApprovedAmount(),
			"status":           "approved",
			"tenure":           req.GetTenure(),
			"monthlyRepayment": req.GetMonthlyRepayment(),
			"effectiveDate":    req.GetEffectiveDate(),
			"expiryDate":       req.GetExpiryDate(),
			"amountPaid":       0,
		},
	}

	

	result, err := loansCollection.UpdateOne(context.Background(), bson.M{"_id": loanId}, update)
	if err != nil {
		log.Println("Database error:", err)
		return &pb.ApproveLoanResponse{Message: "Failed to approve loan", Status: false, StatusCode: http.StatusInternalServerError}, nil
	}

	if result.MatchedCount == 0 {

		return &pb.ApproveLoanResponse{Message: "Loan not found", Status: false, StatusCode: http.StatusNotFound}, nil
	}

	// Initialize the gRPC client
	walletServiceClient, cleanup, err := grpcclient.NewWalletServiceClient(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to WalletService: %v", err)
		return &pb.ApproveLoanResponse{Message: "Internal service error", Status: false, StatusCode: http.StatusInternalServerError}, nil
	}
	defer cleanup()

	creditWalletReq := &walletPb.CreditWalletRequest{
		UserId: existingLoan.UserID.Hex(),
		Amount: req.ApprovedAmount,
	}
	creditWalletResp, err := walletServiceClient.CreditWallet(c, creditWalletReq)

	if creditWalletResp == nil {
		log.Println("Error in creditWallet call:", err)
		return &pb.ApproveLoanResponse{Message: "Unexpected service response", Status: false, StatusCode: http.StatusInternalServerError}, nil
	}

	if err != nil || !creditWalletResp.Status {

		return &pb.ApproveLoanResponse{Message: creditWalletResp.Message, Status: creditWalletResp.Status, StatusCode: creditWalletResp.StatusCode}, nil
	}

	return &pb.ApproveLoanResponse{Message: "Loan approved successfully", Status: true, StatusCode: http.StatusOK}, nil
}

func (s *LoanServiceServer) RejectLoan(ctx context.Context, req *pb.RejectLoanRequest) (*pb.RejectLoanResponse, error) {
	// Initialize the gRPC client
	userServiceClient, cleanup, err := grpcclient.NewUserServiceClient(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to UserService: %v", err)
		return &pb.RejectLoanResponse{Message: "Internal service error", Status: false, StatusCode: http.StatusInternalServerError}, nil
	}
	defer cleanup()

	// Set up the context with authorization metadata
	c := grpcclient.NewAuthContext(context.Background(), configs.Env.TOKEN)

	isAdminReq := &userPb.IsAdminRequest{
		UserId: req.UserId,
	}

	isAdminResp, err_ := userServiceClient.IsAdmin(c, isAdminReq)

	if isAdminResp == nil {
		log.Println("Error in IsAdmin call:", err)
		return &pb.RejectLoanResponse{Message: "Unexpected service response", Status: false, StatusCode: http.StatusInternalServerError}, nil
	}

	if err_ != nil || !isAdminResp.Valid {

		return &pb.RejectLoanResponse{Message: isAdminResp.Message, Status: false, StatusCode: isAdminResp.StatusCode}, nil
	}

	loansCollection := database.DB.Collection("loans")

	userId, err := primitive.ObjectIDFromHex(req.GetUserId())
	if err != nil {
		return &pb.RejectLoanResponse{Message: "Invalid user ID", Status: false, StatusCode: http.StatusBadRequest}, nil
	}

	update := bson.M{
		"$set": bson.M{
			"rejectedBy": userId,
			"status":     "rejected",
		},
	}

	loanId, err := primitive.ObjectIDFromHex(req.GetLoanId())
	if err != nil {
		return &pb.RejectLoanResponse{Message: "Invalid loan ID", Status: false, StatusCode: http.StatusBadRequest}, nil
	}

	result, err := loansCollection.UpdateOne(context.Background(), bson.M{"_id": loanId}, update)
	if err != nil {
		log.Println("Database error:", err)
		return &pb.RejectLoanResponse{Message: "Failed to reject loan", Status: false, StatusCode: http.StatusInternalServerError}, nil
	}

	if result.MatchedCount == 0 {

		return &pb.RejectLoanResponse{Message: "Loan not found", Status: false, StatusCode: http.StatusNotFound}, nil
	}

	return &pb.RejectLoanResponse{Message: "Loan rejected successfully", Status: true, StatusCode: http.StatusOK}, nil
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
		port = "50052" // Set a default port if not specified
	}

	// Start gRPC server for Loan service
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(tokenInterceptor))
	pb.RegisterLoanServiceServer(s, &LoanServiceServer{})

	fmt.Printf("Loan Service running on port %s...", port)
	log.Fatal(s.Serve(lis))
}
