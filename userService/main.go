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
	"github.com/manlikehenryy/loan-management-system-grpc/userService/helpers"
	pb "github.com/manlikehenryy/loan-management-system-grpc/userService/user" // Import generated protobuf code
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
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

	user := User{ Role: "admin", Username: req.GetUsername(), FirstName: req.GetFirstName(), LastName: req.GetLastName()}
	user.SetPassword(req.GetPassword())

	var existingUser User
	err := usersCollection.FindOne(context.Background(), bson.M{"username": strings.TrimSpace(req.GetUsername())}).Decode(&existingUser)
	if err != mongo.ErrNoDocuments {
		if err != nil {
			log.Println("Database error:", err)

			return &pb.RegisterUserResponse{Message: "Failed to check username", Status: false, StatusCode: http.StatusInternalServerError}, err
		}

		return &pb.RegisterUserResponse{Message: "Username already exists", Status: false, StatusCode: http.StatusBadRequest}, err
	}

	_, error_ := usersCollection.InsertOne(context.Background(), user)
	if error_ != nil {
		log.Println("Database error:", error_)
		return &pb.RegisterUserResponse{Message: "Failed to create account", Status: false, StatusCode: http.StatusInternalServerError}, error_
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
			return &pb.LoginUserResponse{Token: "", Message: "Incorrect username or password", Status: false, StatusCode: http.StatusUnauthorized}, err
		}
		log.Println("Database error:", err)
		return &pb.LoginUserResponse{Token: "", Message: "Database error", Status: false, StatusCode: http.StatusInternalServerError}, err
	}

	if err := user.ComparePassword(req.GetPassword()); err != nil {
		return &pb.LoginUserResponse{Token: "", Message: "Incorrect username or password", Status: false, StatusCode: http.StatusUnauthorized}, err
	}

	token, err := helpers.GenerateJwt(user.ID.Hex())
	if err != nil {
		log.Println("Token generation error:", err)

		return &pb.LoginUserResponse{Token: token, Message: "Failed to generate token", Status: false, StatusCode: http.StatusInternalServerError}, err
	}

	return &pb.LoginUserResponse{Token: token, Message: "Logged in successfully", Status: true, StatusCode: http.StatusOK}, nil
}

func (s *UserServiceServer) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {

	userIdStr, err := helpers.ParseJwt(req.GetToken())
	if err != nil {

		return &pb.VerifyTokenResponse{Valid: false, Message: "Unauthorized: Invalid JWT token", StatusCode:  http.StatusUnauthorized}, nil
	}

	_, err_ := primitive.ObjectIDFromHex(userIdStr)
	if err_ != nil {

		return &pb.VerifyTokenResponse{Valid: false, Message: "Unauthorized: Invalid user ID", StatusCode:  http.StatusUnauthorized}, nil
	}

	return &pb.VerifyTokenResponse{Valid: true, Message: "token valid", UserId: userIdStr, StatusCode:  http.StatusOK}, nil
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
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &UserServiceServer{}) // Register UserServiceServer
	fmt.Printf("User Service running on port %s...", port)
	log.Fatal(s.Serve(lis))
}
