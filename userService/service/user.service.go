package service

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/manlikehenryy/loan-management-system-grpc/userService/configs"
	"github.com/manlikehenryy/loan-management-system-grpc/userService/database"
	"github.com/manlikehenryy/loan-management-system-grpc/userService/grpcclient"
	"github.com/manlikehenryy/loan-management-system-grpc/userService/helpers"
	pb "github.com/manlikehenryy/loan-management-system-grpc/userService/user"         // Import generated protobuf code
	walletPb "github.com/manlikehenryy/loan-management-system-grpc/userService/wallet" // Import generated protobuf code
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// User struct
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string             `bson:"username,omitempty"`
	Password  []byte             `bson:"password,omitempty"`
	FirstName string             `bson:"firstName,omitempty"`
	LastName  string             `bson:"lastName,omitempty"`
	Role      string             `bson:"role,omitempty"`
}

// UserServiceServer struct to implement gRPC functions
type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
}

func NewUserServiceServer() *UserServiceServer {
	return &UserServiceServer{}
}

func (user *User) SetPassword(password string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	user.Password = hashedPassword
}

func (user *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword(user.Password, []byte(password))
}

// Register user
func (s *UserServiceServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	usersCollection := database.GetCollection("users")

	if req.GetUsername() == "" || req.GetFirstName() == "" || req.GetLastName() == "" || req.GetPassword() == "" {
		return registerUserErrorResponse("Missing required field(s)", http.StatusBadRequest), nil
	}

	user := User{Role: "user", Username: req.GetUsername(), FirstName: req.GetFirstName(), LastName: req.GetLastName()}
	user.SetPassword(req.GetPassword())

	var existingUser User
	err := usersCollection.FindOne(context.Background(), bson.M{"username": strings.TrimSpace(req.GetUsername())}).Decode(&existingUser)
	if err != mongo.ErrNoDocuments {
		if err != nil {
			log.Println("Database error:", err)

			return registerUserErrorResponse("Failed to check username", http.StatusInternalServerError), nil
		}

		return registerUserErrorResponse("Username already exists", http.StatusBadRequest), nil
	}

	result, error_ := usersCollection.InsertOne(context.Background(), user)
	if error_ != nil {
		log.Println("Database error:", error_)
		return registerUserErrorResponse("Failed to create account", http.StatusInternalServerError), nil
	}

	// Initialize the gRPC client
	walletServiceClient, cleanup, err := grpcclient.NewWalletServiceClient(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to WalletService: %v", err)
		return registerUserErrorResponse("Internal service error", http.StatusInternalServerError), nil
	}
	defer cleanup()

	// Set up the context with authorization metadata
	c := grpcclient.NewAuthContext(context.Background(), configs.Env.TOKEN)

	createWalletReq := &walletPb.CreateWalletRequest{
		UserId: result.InsertedID.(primitive.ObjectID).Hex(),
	}
	createWalletResp, err := walletServiceClient.CreateWallet(c, createWalletReq)

	if createWalletResp == nil {
		log.Println("Error in CreateWallet call:", err)
		return registerUserErrorResponse("Unexpected service response", http.StatusInternalServerError), nil
	}

	if err != nil || !createWalletResp.Status {
		return registerUserErrorResponse(createWalletResp.Message, int(createWalletResp.StatusCode)), nil
	}

	return registerUserSuccessResponse("User registered successfully!", http.StatusOK), nil
}

// Login user and generate JWT
func (s *UserServiceServer) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	usersCollection := database.GetCollection("users")
	var user User
	err := usersCollection.FindOne(ctx, bson.M{"username": req.GetUsername()}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return loginUserErrorResponse("Incorrect username or password", http.StatusUnauthorized, ""), nil
		}
		log.Println("Database error:", err)
		return loginUserErrorResponse("Database error", http.StatusInternalServerError, ""), nil
	}

	if err := user.ComparePassword(req.GetPassword()); err != nil {
		return loginUserErrorResponse("Incorrect username or password", http.StatusUnauthorized, ""), nil
	}

	token, err := helpers.GenerateJwt(user.ID.Hex())
	if err != nil {
		log.Println("Token generation error:", err)

		return loginUserErrorResponse("Failed to generate token", http.StatusInternalServerError, token), nil
	}

	return loginUserSuccessResponse("Logged in successfully", http.StatusOK, token), nil
}

func (s *UserServiceServer) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {

	userIdStr, err := helpers.ParseJwt(req.GetToken())
	if err != nil {

		return verifyTokenErrorResponse("Unauthorized: Invalid JWT token", http.StatusUnauthorized), nil
	}

	_, err_ := primitive.ObjectIDFromHex(userIdStr)
	if err_ != nil {

		return verifyTokenErrorResponse("Unauthorized: Invalid user ID", http.StatusUnauthorized), nil
	}

	return verifyTokenSuccessResponse("token valid", http.StatusOK, userIdStr), nil
}

func (s *UserServiceServer) IsAdmin(ctx context.Context, req *pb.IsAdminRequest) (*pb.IsAdminResponse, error) {
	userId, err_ := primitive.ObjectIDFromHex(req.GetUserId())
	if err_ != nil {

		return isAdminErrorResponse("Unauthorized: Invalid user ID", http.StatusUnauthorized), nil
	}

	usersCollection := database.GetCollection("users")
	var user User
	err := usersCollection.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return isAdminErrorResponse("Unauthorized", http.StatusUnauthorized), nil
		}
		log.Println("Database error:", err)
		return isAdminErrorResponse("Database error", http.StatusInternalServerError), nil
	}

	if user.Role != "admin" {
		return isAdminErrorResponse("Unauthorized", http.StatusUnauthorized), nil
	}

	return isAdminSuccessResponse("Successful", http.StatusOK), nil
}

func registerUserSuccessResponse(message string, statusCode int) *pb.RegisterUserResponse {
	return &pb.RegisterUserResponse{Message: message, Status: true, StatusCode: int32(statusCode)}
}

func registerUserErrorResponse(message string, statusCode int) *pb.RegisterUserResponse {
	return &pb.RegisterUserResponse{Message: message, Status: false, StatusCode: int32(statusCode)}
}

func loginUserSuccessResponse(message string, statusCode int, token string) *pb.LoginUserResponse {
	return &pb.LoginUserResponse{Message: message, Status: true, StatusCode: int32(statusCode), Token: token}
}

func loginUserErrorResponse(message string, statusCode int, token string) *pb.LoginUserResponse {
	return &pb.LoginUserResponse{Message: message, Status: false, StatusCode: int32(statusCode), Token: token}
}

func verifyTokenSuccessResponse(message string, statusCode int, userId string) *pb.VerifyTokenResponse {
	return &pb.VerifyTokenResponse{Message: message, Valid: true, UserId: userId, StatusCode: int32(statusCode)}
}

func verifyTokenErrorResponse(message string, statusCode int) *pb.VerifyTokenResponse {
	return &pb.VerifyTokenResponse{Message: message, Valid: false, StatusCode: int32(statusCode)}
}

func isAdminSuccessResponse(message string, statusCode int) *pb.IsAdminResponse {
	return &pb.IsAdminResponse{Message: message, Valid: true, StatusCode: int32(statusCode)}
}

func isAdminErrorResponse(message string, statusCode int) *pb.IsAdminResponse {
	return &pb.IsAdminResponse{Message: message, Valid: false, StatusCode: int32(statusCode)}
}
