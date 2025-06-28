package routes

import (
	"connection/controllers"
	// "connection/middleware"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	// incomingRoutes.Use(middleware.Authenticate())

	incomingRoutes.POST("users/signup", controllers.Signup())
	incomingRoutes.POST("users/login", controllers.Login())
	incomingRoutes.POST("users/getotp", controllers.GetOTP())
	incomingRoutes.POST("users/verifyotp", controllers.VerifyOTP())

}
