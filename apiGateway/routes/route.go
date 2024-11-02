package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/controllers"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/middleware"
)

func Setup(app *gin.Engine) {


	app.POST("/api/register", controllers.Register)
	app.POST("/api/login", controllers.Login)
	app.GET("/api/logout", controllers.Logout)

	app.Use(middleware.IsAuthenticated)

	app.POST("/api/url", controllers.CreateUrl)
	app.GET("/api/url/:id", controllers.GetUrl)
	app.PUT("/api/url/:id", controllers.UpdateUrl)
	app.GET("/api/url", controllers.GetAllUrl)
	app.DELETE("/api/url/:id", controllers.DeleteUrl)
}