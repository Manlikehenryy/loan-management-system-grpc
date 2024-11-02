package main

import (
	"context"
	"fmt"
	"log"
	"net"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

func (user *User) SetPassword(password string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	user.Password = hashedPassword
}

func (user *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword(user.Password, []byte(password))
}

// Register user
func (s *UserServiceServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	usersCollection := database.DB.Collection("users")

	if req.GetUsername() == "" || req.GetFirstName() == "" || req.GetLastName() == "" || req.GetPassword() == ""  {
		return &pb.RegisterUserResponse{Message: "Missing required field(s)", Status: false, StatusCode: http.StatusBadRequest}, nil	
	}

	user := User{Role: "user", Username: req.GetUsername(), FirstName: req.GetFirstName(), LastName: req.GetLastName()}
	user.SetPassword(req.GetPassword())

	var existingUser User
	err := usersCollection.FindOne(context.Background(), bson.M{"username": strings.TrimSpace(req.GetUsername())}).Decode(&existingUser)
	if err != mongo.ErrNoDocuments {
		if err != nil {
			log.Println("Database error:", err)

			return &pb.RegisterUserResponse{Message: "Failed to check username", Status: false, StatusCode: http.StatusInternalServerError}, nil
		}

		return &pb.RegisterUserResponse{Message: "Username already exists", Status: false, StatusCode: http.StatusBadRequest}, nil
	}

	result, error_ := usersCollection.InsertOne(context.Background(), user)
	if error_ != nil {
		log.Println("Database error:", error_)
		return &pb.RegisterUserResponse{Message: "Failed to create account", Status: false, StatusCode: http.StatusInternalServerError}, nil
	}

	 // Initialize the gRPC client
	 walletServiceClient, cleanup, err := grpcclient.NewWalletServiceClient(ctx)
	 if err != nil {
		 log.Fatalf("Failed to connect to WalletService: %v", err)
		 return &pb.RegisterUserResponse{Message: "Internal service error", Status: false, StatusCode: http.StatusInternalServerError}, nil
	 }
	 defer cleanup()
 
	 // Set up the context with authorization metadata
	 c := grpcclient.NewAuthContext(context.Background(), configs.Env.TOKEN)
	
	createWalletReq := &walletPb.CreateWalletRequest{
        UserId:       result.InsertedID.(primitive.ObjectID).Hex(),
    }
    createWalletResp, err := walletServiceClient.CreateWallet(c, createWalletReq)

	if createWalletResp == nil{
        log.Println("Error in CreateWallet call:", err)
		return &pb.RegisterUserResponse{Message: "Unexpected service response", Status: false, StatusCode: http.StatusInternalServerError}, nil
    }

    if err != nil || !createWalletResp.Status{
		return &pb.RegisterUserResponse{Message: createWalletResp.Message, Status: createWalletResp.Status, StatusCode: createWalletResp.StatusCode}, nil
    }
    

	return &pb.RegisterUserResponse{Message: "User registered successfully!", Status: true, StatusCode: http.StatusOK}, nil
}

// Login user and generate JWT
func (s *UserServiceServer) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	usersCollection := database.DB.Collection("users")
	var user User
	err := usersCollection.FindOne(ctx, bson.M{"username": req.GetUsername()}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &pb.LoginUserResponse{Token: "", Message: "Incorrect username or password", Status: false, StatusCode: http.StatusUnauthorized}, nil
		}
		log.Println("Database error:", err)
		return &pb.LoginUserResponse{Token: "", Message: "Database error", Status: false, StatusCode: http.StatusInternalServerError}, nil
	}

	if err := user.ComparePassword(req.GetPassword()); err != nil {
		return &pb.LoginUserResponse{Token: "", Message: "Incorrect username or password", Status: false, StatusCode: http.StatusUnauthorized}, nil
	}

	token, err := helpers.GenerateJwt(user.ID.Hex())
	if err != nil {
		log.Println("Token generation error:", err)

		return &pb.LoginUserResponse{Token: token, Message: "Failed to generate token", Status: false, StatusCode: http.StatusInternalServerError}, nil
	}

	return &pb.LoginUserResponse{Token: token, Message: "Logged in successfully", Status: true, StatusCode: http.StatusOK}, nil
}

func (s *UserServiceServer) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {

	userIdStr, err := helpers.ParseJwt(req.GetToken())
	if err != nil {

		return &pb.VerifyTokenResponse{Valid: false, Message: "Unauthorized: Invalid JWT token", StatusCode: http.StatusUnauthorized}, nil
	}

	_, err_ := primitive.ObjectIDFromHex(userIdStr)
	if err_ != nil {

		return &pb.VerifyTokenResponse{Valid: false, Message: "Unauthorized: Invalid user ID", StatusCode: http.StatusUnauthorized}, nil
	}

	return &pb.VerifyTokenResponse{Valid: true, Message: "token valid", UserId: userIdStr, StatusCode: http.StatusOK}, nil
}

func (s *UserServiceServer) IsAdmin(ctx context.Context, req *pb.IsAdminRequest) (*pb.IsAdminResponse, error) {
	userId, err_ := primitive.ObjectIDFromHex(req.GetUserId())
	if err_ != nil {

		return &pb.IsAdminResponse{Valid: false, Message: "Unauthorized: Invalid user ID", StatusCode: http.StatusUnauthorized}, nil
	}

	usersCollection := database.DB.Collection("users")
	var user User
	err := usersCollection.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &pb.IsAdminResponse{Valid: false, Message: "Unauthorized", StatusCode: http.StatusUnauthorized}, nil
		}
		log.Println("Database error:", err)
		return &pb.IsAdminResponse{Valid: false, Message: "Database error", StatusCode: http.StatusInternalServerError}, nil
	}

	if user.Role != "admin" {
		return &pb.IsAdminResponse{Valid: false, Message: "Unauthorized", StatusCode: http.StatusUnauthorized}, nil
	}

	return &pb.IsAdminResponse{Valid: true, Message: "Successful", StatusCode: http.StatusOK}, nil
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
	helpers.Initialize()
	// Connect to the database
	database.Connect()

	// Retrieve the port from the config
	port := configs.Env.PORT
	if port == "" {
		port = "50051" // Set a default port if not specified
	}

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(tokenInterceptor))
	pb.RegisterUserServiceServer(s, &UserServiceServer{}) // Register UserServiceServer
	fmt.Printf("User Service running on port %s...", port)
	log.Fatal(s.Serve(lis))
}
