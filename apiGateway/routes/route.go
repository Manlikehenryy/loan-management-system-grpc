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

	app.POST("/api/loan/apply-loan", controllers.ApplyLoan)
	app.PUT("/api/loan/approve-loan", controllers.ApproveLoan)
	app.PUT("/api/loan/reject-loan", controllers.RejectLoan)
}