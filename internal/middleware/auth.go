package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// UserContext represents user information in the request context
type UserContext struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

// JWTRequired middleware validates JWT tokens and extracts user information
func JWTRequired() gin.HandlerFunc {
	secret := getJWTSecret()
	
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format. Expected 'Bearer <token>'",
			})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token cannot be empty",
			})
			return
		}

		// Validate token
		claims, err := validateToken(tokenStr, secret)
		if err != nil {
			log.Printf("JWT validation failed: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		// Extract user information
		userID, ok := claims["user_id"].(float64)
		if !ok {
			log.Printf("Invalid user_id in JWT claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		username, ok := claims["username"].(string)
		if !ok {
			log.Printf("Invalid username in JWT claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		// Set user context
		userCtx := UserContext{
			UserID:   uint(userID),
			Username: username,
		}
		c.Set("user", userCtx)

		c.Next()
	}
}

// GetUserFromContext extracts user information from the request context
func GetUserFromContext(c *gin.Context) (*UserContext, bool) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	user, ok := userInterface.(UserContext)
	if !ok {
		return nil, false
	}

	return &user, true
}

// validateToken validates a JWT token and returns the claims
func validateToken(tokenString, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

// getJWTSecret returns the JWT secret from environment variables
func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			log.Println("WARNING: JWT_SECRET not set, using default secret")
		}
	return secret
}

// OptionalAuth middleware validates JWT tokens if present but doesn't require them
func OptionalAuth() gin.HandlerFunc {
	secret := getJWTSecret()
	
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == "" {
			c.Next()
			return
		}

		// Try to validate token
		claims, err := validateToken(tokenStr, secret)
		if err != nil {
			// Log but don't fail the request
			log.Printf("Optional JWT validation failed: %v", err)
			c.Next()
			return
		}

		// Extract user information
		if userID, ok := claims["user_id"].(float64); ok {
			if username, ok := claims["username"].(string); ok {
				userCtx := UserContext{
					UserID:   uint(userID),
					Username: username,
				}
				c.Set("user", userCtx)
			}
		}

		c.Next()
	}
}

// CORS middleware handles Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}