package helpers

import (
	"github.com/gin-gonic/gin"
)

func SendJSON(c *gin.Context, statusCode int, data interface{}) {
     c.JSON(statusCode, data)
}

func SendError(c *gin.Context, statusCode int, message string) {
     c.JSON(statusCode, gin.H{"error": message})
}
