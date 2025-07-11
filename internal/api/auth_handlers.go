package api

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/sykell/url-crawler/internal/db"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse represents the login response payload
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
}

// Config holds authentication configuration
type Config struct {
	JWTSecret     string
	TokenDuration time.Duration
}

// NewAuthConfig creates a new auth configuration
func NewAuthConfig() *Config {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("WARNING: JWT_SECRET not set, using default secret")
		secret = "changeme"
	}

	duration := 24 * time.Hour
	if durationStr := os.Getenv("JWT_DURATION"); durationStr != "" {
		if parsed, err := time.ParseDuration(durationStr); err == nil {
			duration = parsed
		}
	}

	return &Config{
		JWTSecret:     secret,
		TokenDuration: duration,
	}
}

// LoginHandler handles user authentication
func LoginHandler(dbConn *gorm.DB) gin.HandlerFunc {
	config := NewAuthConfig()
	
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Login validation error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		// Sanitize input
		req.Username = strings.TrimSpace(req.Username)
		if req.Username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username cannot be empty"})
			return
		}

		// Get user from database
		user, err := db.GetUserByUsername(dbConn, req.Username)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("Login attempt with non-existent username: %s", req.Username)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				return
			}
			log.Printf("Database error during login: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			log.Printf("Failed login attempt for user: %s", req.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Generate JWT token
		expiresAt := time.Now().Add(config.TokenDuration)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID,
			"username": user.Username,
			"exp":     expiresAt.Unix(),
			"iat":     time.Now().Unix(),
		})

		tokenStr, err := token.SignedString([]byte(config.JWTSecret))
		if err != nil {
			log.Printf("Failed to sign JWT token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		log.Printf("Successful login for user: %s", req.Username)
		c.JSON(http.StatusOK, LoginResponse{
			Token:     tokenStr,
			ExpiresAt: expiresAt,
			UserID:    user.ID,
			Username:  user.Username,
		})
	}
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string, secret string) (jwt.MapClaims, error) {
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