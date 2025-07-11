package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type URLStatus string

const (
	StatusQueued  URLStatus = "queued"
	StatusRunning URLStatus = "running"
	StatusDone    URLStatus = "done"
	StatusError   URLStatus = "error"
)

// URL represents a web page to be crawled
type URL struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Address       string    `gorm:"unique;not null;size:768" json:"address"`
	Title         string    `json:"title"`
	HTMLVersion   string    `json:"html_version"`
	HeadingCounts string    `json:"heading_counts"` // JSON: {"h1":2,"h2":1...}
	InternalLinks int       `json:"internal_links"`
	ExternalLinks int       `json:"external_links"`
	BrokenLinks   int       `json:"broken_links"`
	BrokenList    string    `json:"broken_list"` // JSON: [{"url":"...","code":404}]
	HasLoginForm  bool      `json:"has_login_form"`
	Status        URLStatus `gorm:"default:'queued'" json:"status"`
	Error         string    `json:"error"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// User represents an authenticated user
type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null;size:100" json:"username"`
	Password  string    `gorm:"not null;size:255" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	MaxOpen  int
	MaxIdle  int
	Timeout  time.Duration
}

// NewConfig creates a new database configuration from environment variables
func NewConfig() *Config {
	return &Config{
		Host:     getEnvOrDefault("MYSQL_HOST", "localhost"),
		Port:     getEnvOrDefault("MYSQL_PORT", "3306"),
		User:     getEnvOrDefault("MYSQL_USER", "root"),
		Password: getEnvOrDefault("MYSQL_PASSWORD", ""),
		Database: getEnvOrDefault("MYSQL_DATABASE", "url_crawler"),
		MaxOpen:  25,
		MaxIdle:  5,
		Timeout:  30 * time.Second,
	}
}

// InitDB initializes the database connection with proper configuration
func InitDB() (*gorm.DB, error) {
	config := NewConfig()
	
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		config.User, config.Password, config.Host, config.Port, config.Database)

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxOpen)
	sqlDB.SetMaxIdleConns(config.MaxIdle)
	sqlDB.SetConnMaxLifetime(config.Timeout)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations performs database migrations
func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &URL{})
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CreateUser creates a new user with proper validation
func CreateUser(db *gorm.DB, username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password cannot be empty")
	}

	user := User{
		Username: username,
		Password: password,
	}

	return db.Create(&user).Error
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(db *gorm.DB, username string) (*User, error) {
	var user User
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateURLStatus updates the status of a URL
func UpdateURLStatus(db *gorm.DB, id uint, status URLStatus, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
		"error":  errorMsg,
	}
	return db.Model(&URL{}).Where("id = ?", id).Updates(updates).Error
}

// GetURLByID retrieves a URL by ID
func GetURLByID(db *gorm.DB, id uint) (*URL, error) {
	var url URL
	err := db.First(&url, id).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}

// CreateURL creates a new URL with address
func CreateURL(db *gorm.DB, address string) (*URL, error) {
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	url := URL{
		Address:  address,
		Status:      StatusQueued,
	}

	err := db.Create(&url).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}

// GetURLByAddress retrieves a URL by address
func GetURLByAddress(db *gorm.DB, address string) (*URL, error) {
	var url URL
	err := db.Where("address = ?", address).First(&url).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}