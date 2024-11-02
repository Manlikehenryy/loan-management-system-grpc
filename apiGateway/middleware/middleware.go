package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/helpers"
)

var ctx = context.Background()

func IsAuthenticated(c *gin.Context) {
	// Retrieve the JWT token from the cookie
	token, err := c.Cookie("jwt")
	if err != nil {
		helpers.SendError(c, http.StatusUnauthorized, "Unauthorized: No JWT token provided")
		c.Abort() // Abort the request pipeline if authentication fails
		return
	}

	// Parse the JWT token from the cookie
	// userId, err := helpers.ParseJwt(cookie)
	// if err != nil {
	// 	helpers.SendError(c, http.StatusUnauthorized, "Unauthorized: Invalid JWT token")
	// 	c.Abort() // Abort the request pipeline if token parsing fails
	// 	return
	// }

	// Store the user ID in the request context
	// c.Set("userId", userId)

	// Continue to the next middleware or handler
	c.Next()
}
