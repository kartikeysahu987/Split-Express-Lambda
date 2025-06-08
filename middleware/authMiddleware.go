package middleware

import (
	"connection/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token directly from the header
		clientToken := c.GetHeader("token")
		if clientToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is required in header"})
			c.Abort()
			return
		}

		// Validate the token
		claims, err := helpers.ValidateToken(clientToken)
		if err != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("email", claims.Email)
		c.Set("first_name", claims.First_name)
		c.Set("last_name", claims.Last_Name)
		c.Set("uid", claims.Uid)
		c.Set("user_type", claims.User_type)

		c.Next()
	}
}
