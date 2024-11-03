package service

import (
	"context"
	"log"
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

func NewLoanServiceServer() *LoanServiceServer {
    return &LoanServiceServer{}
}


func (s *LoanServiceServer) ApplyLoan(ctx context.Context, req *pb.ApplyLoanRequest) (*pb.ApplyLoanResponse, error) {
	loansCollection := database.DB.Collection("loans")

	userId, err := primitive.ObjectIDFromHex(req.GetUserId())
	if err != nil {
		return applyLoanErrorResponse("Invalid user ID", http.StatusBadRequest), nil
	}

	loan := Loan{UserID: userId, Amount: req.GetAmount(), Status: "pending"}

	result, err_ := loansCollection.InsertOne(context.Background(), loan)
	if err_ != nil {
		return applyLoanErrorResponse("Loan application failed", http.StatusInternalServerError), nil
	}
	return applyLoanSuccessResponse("Loan application submitted",http.StatusCreated, result.InsertedID.(primitive.ObjectID).Hex()), nil
}

func (s *LoanServiceServer) ApproveLoan(ctx context.Context, req *pb.ApproveLoanRequest) (*pb.ApproveLoanResponse, error) {

	// Initialize the gRPC client
	userServiceClient, cleanup, err := grpcclient.NewUserServiceClient(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to UserService: %v", err)
		return approveLoanErrorResponse("Internal service error", http.StatusInternalServerError), nil
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
		return approveLoanErrorResponse("Unexpected service response", http.StatusInternalServerError), nil
	}

	if err_ != nil || !isAdminResp.Valid {

		return approveLoanErrorResponse(isAdminResp.Message, int(isAdminResp.StatusCode)), nil
	}

	loansCollection := database.DB.Collection("loans")

	loanId, err := primitive.ObjectIDFromHex(req.GetLoanId())
	if err != nil {
		return approveLoanErrorResponse("Invalid loan ID", http.StatusBadRequest), nil
	}

	var existingLoan Loan
	err = loansCollection.FindOne(context.Background(), bson.M{"_id": loanId}).Decode(&existingLoan)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return approveLoanErrorResponse("Loan not found", http.StatusNotFound), nil
		} else {
			return approveLoanErrorResponse("Failed to approve loan", http.StatusInternalServerError), nil
		}
	}

	userId, err := primitive.ObjectIDFromHex(req.GetUserId())
	if err != nil {
		return approveLoanErrorResponse("Invalid user ID", http.StatusBadRequest), nil
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
		return approveLoanErrorResponse("Failed to approve loan", http.StatusInternalServerError), nil
	}

	if result.MatchedCount == 0 {

		return approveLoanErrorResponse("Loan not found", http.StatusNotFound), nil
	}

	// Initialize the gRPC client
	walletServiceClient, cleanup, err := grpcclient.NewWalletServiceClient(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to WalletService: %v", err)
		return approveLoanErrorResponse("Internal service error", http.StatusInternalServerError), nil
	}
	defer cleanup()

	creditWalletReq := &walletPb.CreditWalletRequest{
		UserId: existingLoan.UserID.Hex(),
		Amount: req.ApprovedAmount,
	}
	creditWalletResp, err := walletServiceClient.CreditWallet(c, creditWalletReq)

	if creditWalletResp == nil {
		log.Println("Error in creditWallet call:", err)
		return approveLoanErrorResponse("Unexpected service response", http.StatusInternalServerError), nil
	}

	if err != nil || !creditWalletResp.Status {

		return approveLoanErrorResponse(creditWalletResp.Message, int(creditWalletResp.StatusCode)), nil
	}

	return approveLoanSuccessResponse("Loan approved successfully", http.StatusOK), nil
}

func (s *LoanServiceServer) RejectLoan(ctx context.Context, req *pb.RejectLoanRequest) (*pb.RejectLoanResponse, error) {
	// Initialize the gRPC client
	userServiceClient, cleanup, err := grpcclient.NewUserServiceClient(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to UserService: %v", err)
		return rejectLoanErrorResponse("Internal service error", http.StatusInternalServerError), nil
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
		return rejectLoanErrorResponse("Unexpected service response", http.StatusInternalServerError), nil
	}

	if err_ != nil || !isAdminResp.Valid {

		return rejectLoanErrorResponse(isAdminResp.Message, int(isAdminResp.StatusCode)), nil
	}

	loansCollection := database.DB.Collection("loans")

	userId, err := primitive.ObjectIDFromHex(req.GetUserId())
	if err != nil {
		return rejectLoanErrorResponse("Invalid user ID", http.StatusBadRequest), nil
	}

	update := bson.M{
		"$set": bson.M{
			"rejectedBy": userId,
			"status":     "rejected",
		},
	}

	loanId, err := primitive.ObjectIDFromHex(req.GetLoanId())
	if err != nil {
		return  rejectLoanErrorResponse("Invalid loan ID", http.StatusBadRequest), nil
	}

	result, err := loansCollection.UpdateOne(context.Background(), bson.M{"_id": loanId}, update)
	if err != nil {
		log.Println("Database error:", err)
		return rejectLoanErrorResponse("Failed to reject loan", http.StatusInternalServerError), nil
	}

	if result.MatchedCount == 0 {

		return rejectLoanErrorResponse("Loan not found", http.StatusNotFound), nil
	}

	return rejectLoanSuccessResponse("Loan rejected successfully", http.StatusOK), nil
}

// Success and Error Response functions for ApplyLoan
func applyLoanSuccessResponse(message string, statusCode int, loanId string) *pb.ApplyLoanResponse {
    return &pb.ApplyLoanResponse{Message: message, LoandId: loanId, Status: true, StatusCode: int32(statusCode)}
}

func applyLoanErrorResponse(message string, statusCode int) *pb.ApplyLoanResponse {
    return &pb.ApplyLoanResponse{Message: message, Status: false, StatusCode: int32(statusCode)}
}

// Success and Error Response functions for ApproveLoan
func approveLoanSuccessResponse(message string, statusCode int) *pb.ApproveLoanResponse {
    return &pb.ApproveLoanResponse{Message: message, Status: true, StatusCode: int32(statusCode)}
}

func approveLoanErrorResponse(message string, statusCode int) *pb.ApproveLoanResponse {
    return &pb.ApproveLoanResponse{Message: message, Status: false, StatusCode: int32(statusCode)}
}

// Success and Error Response functions for RejectLoan
func rejectLoanSuccessResponse(message string, statusCode int) *pb.RejectLoanResponse {
    return &pb.RejectLoanResponse{Message: message, Status: true, StatusCode: int32(statusCode)}
}

func rejectLoanErrorResponse(message string, statusCode int) *pb.RejectLoanResponse {
    return &pb.RejectLoanResponse{Message: message, Status: false, StatusCode: int32(statusCode)}
}