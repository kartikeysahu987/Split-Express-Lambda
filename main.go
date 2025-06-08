package main

import (
	"connection/routes"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func main(){
	port := os.Getenv("PORT")

	if port==""{
		port="8000"
		fmt.Println("Nothing in port in .env so taken 8000")
		// return 
	}
	router:=gin.New()
	router.Use(gin.Logger())
	
	routes.AuthRoutes(router)
	routes.UserRoutes(router)



	router.Run(":"+port)	



}