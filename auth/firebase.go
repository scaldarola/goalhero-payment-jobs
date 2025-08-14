package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/api/option"
)

var firebaseAuth *auth.Client
var jwtSecretKey []byte

// InitFirebase initializes Firebase connection
func InitFirebase() {
	log.Println("🔥 Initializing Firebase...")
	
	// Check for Firebase credentials
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		credentialsPath = "./auth/firebase_credentials.json"
	}

	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		log.Printf("⚠️ Firebase credentials not found at %s, Firebase auth will be disabled", credentialsPath)
		return
	}

	// Initialize Firebase app
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase app: %v", err)
	}

	// Initialize Firebase Auth client
	firebaseAuth, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth client: %v", err)
	}

	// Initialize JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-default-secret-key" // Change this in production
	}
	jwtSecretKey = []byte(jwtSecret)

	log.Println("✅ Firebase initialized successfully")
}

// FirebaseAuthMiddleware validates Firebase tokens
func FirebaseAuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if firebaseAuth == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Authentication service unavailable",
			})
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format",
			})
			c.Abort()
			return
		}

		idToken := tokenParts[1]

		// Verify the Firebase token
		token, err := firebaseAuth.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("userID", token.UID)
		c.Set("userClaims", token.Claims)

		c.Next()
	})
}

// GenerateJWT generates a JWT token for internal service communication
func GenerateJWT(userID string, claims map[string]interface{}) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"claims": claims,
		"exp":    jwt.TimeFunc().Add(24 * 60 * 60 * 1000).Unix(), // 24 hours
	})

	return token.SignedString(jwtSecretKey)
}

// ValidateJWT validates a JWT token for internal service communication
func ValidateJWT(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}