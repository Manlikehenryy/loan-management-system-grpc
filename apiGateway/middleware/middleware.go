package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/configs"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/grpcclient"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/helpers"
	userPb "github.com/manlikehenryy/loan-management-system-grpc/apiGateway/user"
)

func IsAuthenticated(c *gin.Context) {
	// Retrieve the JWT token from the cookie
	token, err := c.Cookie("jwt")
	if err != nil {
		helpers.SendError(c, http.StatusUnauthorized, "Unauthorized: No JWT token provided")
		c.Abort() // Abort the request pipeline if authentication fails
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

	verifyTokenReq := &userPb.VerifyTokenRequest{
		Token: token,
	}

	verifyTokenResp, err_ := userServiceClient.VerifyToken(ctx, verifyTokenReq)

	if verifyTokenResp == nil {
		helpers.SendError(c, http.StatusInternalServerError, "Unexpected service response")
		return
	}

	if err_ != nil || !verifyTokenResp.Valid {
		helpers.SendError(c, int(verifyTokenResp.StatusCode), verifyTokenResp.Message)
		return
	}

	// Store the user ID in the request context
	c.Set("userId", verifyTokenResp.UserId)

	// Continue to the next middleware or handler
	c.Next()
}
