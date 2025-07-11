package main

import (
	"flag"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/sykell/url-crawler/internal/db"
)

// SeedConfig holds seed configuration
type SeedConfig struct {
	Username string
	Password string
	Force    bool
}

// NewSeedConfig creates a new seed configuration
func NewSeedConfig() *SeedConfig {
	username := flag.String("username", "admin", "Admin username")
	password := flag.String("password", "adminpass", "Admin password")
	force := flag.Bool("force", false, "Force recreation of admin user")
	
	flag.Parse()

	return &SeedConfig{
		Username: *username,
		Password: *password,
		Force:    *force,
	}
}

func main() {
	config := NewSeedConfig()

	// Validate configuration
	if config.Username == "" {
		log.Fatal("Username cannot be empty")
	}
	if config.Password == "" {
		log.Fatal("Password cannot be empty")
	}
	if len(config.Password) < 6 {
		log.Fatal("Password must be at least 6 characters long")
	}

	log.Println("Starting database seeding...")

	// Initialize database connection
	dbConn, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Check if admin user already exists
	var existingUser db.User
	err = dbConn.Where("username = ?", config.Username).First(&existingUser).Error
	if err == nil {
		if !config.Force {
			log.Printf("Admin user '%s' already exists. Use -force flag to recreate.", config.Username)
			return
		}
		
		log.Printf("Recreating admin user '%s'...", config.Username)
		if err := dbConn.Delete(&existingUser).Error; err != nil {
			log.Fatalf("Failed to delete existing user: %v", err)
		}
	} else if err != gorm.ErrRecordNotFound {
		log.Fatalf("Database error checking existing user: %v", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(config.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create admin user
	adminUser := db.User{
		Username: config.Username,
		Password: string(hashedPassword),
	}

	if err := dbConn.Create(&adminUser).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Printf("Successfully created admin user: %s/%s", config.Username, config.Password)
	log.Printf("User ID: %d", adminUser.ID)
	log.Println("Database seeding completed successfully")
}