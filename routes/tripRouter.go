package routes

import (
	"connection/controllers"
	"connection/middleware"

	"github.com/gin-gonic/gin"
)

func TripRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())

	incomingRoutes.POST("/trip/create", controllers.CreateTrip())
	incomingRoutes.GET("/trip/getalltrip", controllers.GetAllTrip())
	incomingRoutes.GET("/trip/getallmytrip", controllers.GetAllMyTrip())
	incomingRoutes.POST("/trip/getmembers", controllers.GetAllNotFreeMemberOnInviteCode())
	incomingRoutes.POST("/trip/linkmember", controllers.LinkMember())
	incomingRoutes.POST("/trip/automaticlinkmember", controllers.AutomaticLinkMember())
	incomingRoutes.POST("/trip/pay", controllers.Pay())
	incomingRoutes.POST("/trip/settle", controllers.Settle())
	incomingRoutes.POST("/trip/getAllTransaction", controllers.GetAllTransaction())
	incomingRoutes.POST("/trip/getsettlements", controllers.GetSettlements())
	incomingRoutes.POST("/trip/getcausualnamebyuid", controllers.GetCasualNameByUID())
	incomingRoutes.POST("/trip/contactinfo", controllers.GetContactInfo())
	incomingRoutes.POST("/trip/deleteTrip", controllers.DeleteTrip())
	incomingRoutes.POST("/trip/deleteTransaction", controllers.DeleteTransaction())
}
