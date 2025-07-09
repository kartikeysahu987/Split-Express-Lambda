package routes

import (
	"connection/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/auth/signup", controllers.Signup())
	incomingRoutes.POST("/auth/login", controllers.Login())
	incomingRoutes.POST("/auth/getotp", controllers.GetOTP())
	incomingRoutes.POST("/auth/verifyotp", controllers.VerifyOTP())

}
