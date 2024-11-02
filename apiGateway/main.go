package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/configs"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/routes"
	userPb "github.com/manlikehenryy/loan-management-system-grpc/apiGateway/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var userServiceClient userPb.UserServiceClient

func init() {
	// walletConn, err := grpc.NewClient("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	log.Fatalf("Did not connect: %v", err)
	// }
	// defer walletConn.Close()

	// walletServiceClient = walletPb.NewWalletServiceClient(walletConn)

	userConn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer userConn.Close()

	userServiceClient = userPb.NewUserServiceClient(userConn)
}


func main() {

	// Retrieve the port from the config
	port := configs.Env.PORT
	if port == "" {
		port = "8080" // Set a default port if not specified
	}

	// Create a new Gin router
	app := gin.New()

	// Apply middleware
	app.Use(gin.Logger())
	app.Use(gin.Recovery())

	// Set up routes
	routes.Setup(app)

	// Start the server
	err := app.Run(":" + port)
	if err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
