package service

import (
	"fmt"

	"github.com/sykell/url-crawler/internal/db"
	"gorm.io/gorm"
)

// CreateUser creates a new user with proper validation
func CreateUser(dbConn *gorm.DB, username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password cannot be empty")
	}

	user := db.User{
		Username: username,
		Password: password,
	}

	return dbConn.Create(&user).Error
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(dbConn *gorm.DB, username string) (*db.User, error) {
	var user db.User
	err := dbConn.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
} 