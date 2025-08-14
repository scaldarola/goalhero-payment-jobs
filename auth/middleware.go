package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func FirebaseAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			fmt.Println("Missing or invalid Authorization header")
			// Return 401 Unauthorized if the header is missing or does not start with "Bearer "
			// This is a common practice for APIs that require token-based authentication
			// It ensures that the request is authenticated before proceeding to the next handler
			// This helps protect the API from unauthorized access and ensures that only requests with a valid token can access protected resources
			fmt.Println("Unauthorized access attempt")
			// Abort the request and return a 401 Unauthorized response with a JSON error message
			// This response indicates that the request is not authorized due to a missing or invalid token
			// It provides a clear error message to the client, allowing them to understand the issue
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := FirebaseAuthClient.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			fmt.Printf("Error verifying ID token: %v\n", err)
			// If the token is invalid or expired, return a 401 Unauthorized response
			// This indicates that the provided token is not valid or has expired, preventing access to protected resources
			// It is important to handle token verification errors to ensure that only valid tokens
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("uid", token.UID)
		c.Next()
	}
}
