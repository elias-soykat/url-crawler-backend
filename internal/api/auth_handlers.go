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

	"github.com/sykell/url-crawler/internal/service"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=6"`
}

// SignupRequest represents the signup request payload
type SignupRequest struct {
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

// SignupResponse represents the signup response payload
type SignupResponse struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// Config holds authentication configuration
type Config struct {
	JWTSecret     string
	TokenDuration time.Duration
}

// NewAuthConfig creates a new auth configuration
func NewAuthConfig() *Config {
	secret := os.Getenv("JWT_SECRET")
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
		user, err := service.GetUserByUsername(dbConn, req.Username)
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

// SignupHandler handles user registration
func SignupHandler(dbConn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SignupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Signup validation error: %v", err)
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

		// Check if user already exists
		existingUser, err := service.GetUserByUsername(dbConn, req.Username)
		if err == nil && existingUser != nil {
			log.Printf("Signup attempt with existing username: %s", req.Username)
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		} else if err != gorm.ErrRecordNotFound {
			log.Printf("Database error during signup user check: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash password during signup: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
			return
		}

		// Create new user
		if err := service.CreateUser(dbConn, req.Username, string(hashedPassword)); err != nil {
			log.Printf("Failed to create user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// Get the created user to return the ID
		newUser, err := service.GetUserByUsername(dbConn, req.Username)
		if err != nil {
			log.Printf("Failed to fetch created user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User created but failed to fetch details"})
			return
		}

		log.Printf("Successfully created user: %s (ID: %d)", newUser.Username, newUser.ID)
		c.JSON(http.StatusCreated, SignupResponse{
			UserID:   newUser.ID,
			Username: newUser.Username,
			Message:  "User created successfully",
		})
	}
}