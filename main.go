// package main

// import (
// 	"connection/routes"
// 	"context"
// 	"log"

// 	"github.com/aws/aws-lambda-go/events"
// 	"github.com/aws/aws-lambda-go/lambda"
// 	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
// 	"github.com/gin-gonic/gin"
// )

// var ginLambdaV2 *ginadapter.GinLambdaV2

// func init() {
// 	r := gin.Default()

// 	r.GET("/health", func(c *gin.Context) {
// 		c.JSON(200, gin.H{"msg": "SplitExpress API running ✅"})
// 	})

// 	log.Println(">> Registering Auth Routes")
// 	routes.AuthRoutes(r)

// 	log.Println(">> Registering User Routes")
// 	routes.UserRoutes(r)

// 	log.Println(">> Registering Trip Routes")
// 	routes.TripRoutes(r)

// 	// Fallback route
// 	r.NoRoute(func(c *gin.Context) {
// 		log.Printf(">> No route matched: [%s] %s", c.Request.Method, c.Request.URL.Path)
// 		c.JSON(404, gin.H{"error": "Route not found"})
// 	})

// 	// Convert to Lambda
// 	 ginLambdaV2 = ginadapter.NewV2(r)
// }

// func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
// 	log.Println("====== Incoming Request ======")
// 	log.Printf("Full Request: %+v", req)
// 	log.Println("================================")
//   return ginLambdaV2.ProxyWithContext(ctx, req)
// }

// func main() {
// 	lambda.Start(handler)
// }

package main

import (
	"context"
	"log"

	"connection/routes"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var mongoClient *mongo.Client
var ginLambdaV2 *ginadapter.GinLambdaV2

func init() {
    r := gin.Default()
    r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"msg": "running"}) })

    log.Println(">> Registering auth/user/trip routes")
    routes.AuthRoutes(r)
    routes.UserRoutes(r)
    routes.TripRoutes(r)

    r.NoRoute(func(c *gin.Context) {
        c.JSON(404, gin.H{"error": "Route not found"})
    })

    ginLambdaV2 = ginadapter.NewV2(r)
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
    return ginLambdaV2.ProxyWithContext(ctx, req)
}

func main() {
    lambda.Start(handler)
}
