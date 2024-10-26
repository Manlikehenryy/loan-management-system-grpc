package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/manlikehenryy/loan-management-system-grpc/loanService/configs"
	"github.com/manlikehenryy/loan-management-system-grpc/loanService/database"
	pb "github.com/manlikehenryy/loan-management-system-grpc/loanService/loan"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
)

// Loan struct
type Loan struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	UserID         primitive.ObjectID `bson:"userId,omitempty"`
	Amount         float32            `bson:"amount,omitempty"`
	Status         string             `bson:"status,omitempty"` // "approved", "rejected"
	ApprovedBy     primitive.ObjectID `bson:"approvedBy,omitempty"`
	RejectedBy     primitive.ObjectID `bson:"rejectedBy,omitempty"`
	ApprovedAmount float32            `bson:"approvedAmount,omitempty"`
}

// LoanServiceServer struct to implement gRPC functions
type LoanServiceServer struct {
	pb.UnimplementedLoanServiceServer
}

func (s *LoanServiceServer) ApplyLoan(ctx context.Context, req *pb.ApplyLoanRequest) (*pb.ApplyLoanResponse, error) {
	loansCollection := database.DB.Collection("loans")

	userId, err := primitive.ObjectIDFromHex(req.GetUserId())
	if err != nil {
		return &pb.ApplyLoanResponse{Message: "Invalid user ID", Status: false, StatusCode: http.StatusBadRequest}, err
	}

	loan := Loan{UserID: userId, Amount: req.GetAmount(), Status: "pending"}

	result, err_ := loansCollection.InsertOne(context.Background(), loan)
	if err_ != nil {
		return &pb.ApplyLoanResponse{Message: "Loan application failed", Status: false, StatusCode: http.StatusInternalServerError}, err_
	}
	return &pb.ApplyLoanResponse{Message: "Loan application submitted", Status: true, StatusCode: http.StatusCreated, LoandId: result.InsertedID.(primitive.ObjectID).Hex()}, nil
}

func (s *LoanServiceServer) ApproveLoan(ctx context.Context, req *pb.ApproveLoanRequest) (*pb.ApproveLoanResponse, error) {
	loansCollection := database.DB.Collection("loans")

	// var existingLoan Loan
	// err := loansCollection.FindOne(context.Background(), bson.M{"_id": req.GetLoanId()}).Decode(&existingLoan)
	// if err != nil {
	// 	if err == mongo.ErrNoDocuments {
	// 		return &pb.ApproveLoanResponse{Message: "Loan not found", Status: false, StatusCode: http.StatusNotFound}, err
	// 	} else {
	// 		return &pb.ApproveLoanResponse{Message: "Failed to approve loan", Status: false, StatusCode: http.StatusInternalServerError}, err
	// 	}
	// }
	userId, err := primitive.ObjectIDFromHex(req.GetUserId())
	if err != nil {
		return &pb.ApproveLoanResponse{Message: "Invalid user ID", Status: false, StatusCode: http.StatusBadRequest}, err
	}

	update := bson.M{
		"$set": bson.M{
			"approvedBy":     userId,
			"approvedAmount": req.GetApprovedAmount(),
			"status":         "approved",
		},
	}

	loanId, err := primitive.ObjectIDFromHex(req.GetLoanId())
	if err != nil {
		return &pb.ApproveLoanResponse{Message: "Invalid loan ID", Status: false, StatusCode: http.StatusBadRequest}, err
	}


	result, err := loansCollection.UpdateOne(context.Background(), bson.M{"_id": loanId}, update)
	if err != nil {
		log.Println("Database error:", err)
		return &pb.ApproveLoanResponse{Message: "Failed to approve loan", Status: false, StatusCode: http.StatusInternalServerError}, err
	}

	if result.MatchedCount == 0 {

		return &pb.ApproveLoanResponse{Message: "Loan not found", Status: false, StatusCode: http.StatusNotFound}, err
	}

	return &pb.ApproveLoanResponse{Message: "Loan approved successfully", Status: true, StatusCode: http.StatusOK}, nil
}

func (s *LoanServiceServer) RejectLoan(ctx context.Context, req *pb.RejectLoanRequest) (*pb.RejectLoanResponse, error) {
	loansCollection := database.DB.Collection("loans")

	userId, err := primitive.ObjectIDFromHex(req.GetUserId())
	if err != nil {
		return &pb.RejectLoanResponse{Message: "Invalid user ID", Status: false, StatusCode: http.StatusBadRequest}, err
	}

	update := bson.M{
		"$set": bson.M{
			"rejectedBy": userId,
			"status":     "rejected",
		},
	}

	loanId, err := primitive.ObjectIDFromHex(req.GetLoanId())
	if err != nil {
		return &pb.RejectLoanResponse{Message: "Invalid loan ID", Status: false, StatusCode: http.StatusBadRequest}, err
	}


	result, err := loansCollection.UpdateOne(context.Background(), bson.M{"_id": loanId}, update)
	if err != nil {
		log.Println("Database error:", err)
		return &pb.RejectLoanResponse{Message: "Failed to reject loan", Status: false, StatusCode: http.StatusInternalServerError}, err
	}

	if result.MatchedCount == 0 {

		return &pb.RejectLoanResponse{Message: "Loan not found", Status: false, StatusCode: http.StatusNotFound}, err
	}

	return &pb.RejectLoanResponse{Message: "Loan rejected successfully", Status: true, StatusCode: http.StatusOK}, nil
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
	s := grpc.NewServer()
	pb.RegisterLoanServiceServer(s, &LoanServiceServer{})

	fmt.Printf("Loan Service running on port %s...", port)
	log.Fatal(s.Serve(lis))
}
