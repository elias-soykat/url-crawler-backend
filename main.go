package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sykell/url-crawler/internal/api"
	"github.com/sykell/url-crawler/internal/crawler"
	"github.com/sykell/url-crawler/internal/db"
	"github.com/sykell/url-crawler/internal/middleware"
)

// Config holds application configuration
type Config struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// NewConfig creates a new configuration from environment variables
func NewConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port:            port,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

func main() {
	// Initialize configuration
	config := NewConfig()

	// Initialize database
	log.Println("Initializing database...")
	dbConn, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully")

	// Initialize crawler service
	log.Println("Initializing crawler service...")
	crawlerService := crawler.NewService(dbConn, nil)
	if err := crawlerService.Start(); err != nil {
		log.Fatalf("Failed to start crawler service: %v", err)
	}
	log.Println("Crawler service started successfully")

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   "url-crawler",
		})
	})

	// Authentication endpoint
	r.POST("/auth/login", api.LoginHandler(dbConn))

	// Protected routes
	authorized := r.Group("/")
	authorized.Use(middleware.JWTRequired())
	{
		authorized.POST("/urls", api.PostURLHandler(dbConn, crawlerService))
		authorized.GET("/urls", api.ListURLsHandler(dbConn))
		authorized.GET("/urls/:id", api.GetURLHandler(dbConn))
		authorized.POST("/urls/bulk", api.BulkHandler(dbConn, crawlerService))
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      r,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create shutdown context
	ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	// Shutdown server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Stop crawler service gracefully
	if err := crawlerService.Stop(); err != nil {
		log.Printf("Failed to stop crawler service: %v", err)
	}

	log.Println("Server exited")
}