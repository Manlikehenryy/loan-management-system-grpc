package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/configs"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/dto"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/grpcclient"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/helpers"
	userPb "github.com/manlikehenryy/loan-management-system-grpc/apiGateway/user"
)


func Register(c *gin.Context) {
	var registerDto dto.RegisterDto

	if err := c.ShouldBindJSON(&registerDto); err != nil {
		log.Println("Unable to parse body:", err)
		helpers.SendError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Initialize the gRPC client
    userServiceClient, cleanup, err := grpcclient.NewUserServiceClient(c)
    if err != nil {
        log.Fatalf("Failed to connect to UserService: %v", err)
        helpers.SendError(c, http.StatusInternalServerError, "Internal service error")
        return
    }
    defer cleanup()

    // Set up the context with authorization metadata
    ctx := grpcclient.NewAuthContext(context.Background(), configs.Env.TOKEN)

	registerReq := &userPb.RegisterUserRequest{
		Username: registerDto.Username,
		Password: registerDto.Password,
		FirstName: registerDto.FirstName,
		LastName: registerDto.LastName,
	}

	registerResp, err_ := userServiceClient.RegisterUser(ctx, registerReq)

	if registerResp == nil{
        helpers.SendError(c, http.StatusInternalServerError, "Unexpected service response")
        return
    }

	if err_ != nil || !registerResp.Status {
		helpers.SendError(c, int(registerResp.StatusCode), registerResp.Message)
		return
	}


	helpers.SendJSON(c, http.StatusCreated, gin.H{
		"message": registerResp.Message,
	})
}

// auth.controller.go
func Login(c *gin.Context) {
    var loginDto dto.LoginDto

    if err := c.ShouldBindJSON(&loginDto); err != nil {
        log.Println("Unable to parse body:", err)
        helpers.SendError(c, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Initialize the gRPC client
    userServiceClient, cleanup, err := grpcclient.NewUserServiceClient(c)
    if err != nil {
        log.Fatalf("Failed to connect to UserService: %v", err)
        helpers.SendError(c, http.StatusInternalServerError, "Internal service error")
        return
    }
    defer cleanup()

    // Set up the context with authorization metadata
    ctx := grpcclient.NewAuthContext(context.Background(), configs.Env.TOKEN)

    // Create the request and send it
    loginReq := &userPb.LoginUserRequest{
        Username: loginDto.Username,
        Password: loginDto.Password,
    }

    loginResp, err := userServiceClient.LoginUser(ctx, loginReq)

    // Check if loginResp is nil to avoid nil pointer dereference
    if loginResp == nil || err != nil {
        log.Println("Error in LoginUser call:", err)
        helpers.SendError(c, http.StatusInternalServerError, "Unexpected service response")
        return
    }

    if !loginResp.Status {
        helpers.SendError(c, int(loginResp.StatusCode), loginResp.Message)
        return
    }

    maxAge := int(time.Hour * 24 / time.Second)
    c.SetCookie("jwt", loginResp.Token, maxAge, "/", configs.Env.APP_URL, configs.Env.MODE == "production", true)

    helpers.SendJSON(c, http.StatusOK, gin.H{
        "message": loginResp.Message,
    })
}


func Logout(c *gin.Context) {

	maxAge := -1 * int(time.Hour*24/time.Second)

	c.SetCookie("jwt", "", maxAge, "/", configs.Env.APP_URL, configs.Env.MODE == "production", true)

	helpers.SendJSON(c, http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
