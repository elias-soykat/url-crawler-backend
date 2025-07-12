package service

import (
	"fmt"

	"github.com/sykell/url-crawler/internal/db"
	"gorm.io/gorm"
)

// UpdateURLStatus updates the status of a URL
func UpdateURLStatus(dbConn *gorm.DB, id uint, status db.URLStatus, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
		"error":  errorMsg,
	}
	return dbConn.Model(&db.URL{}).Where("id = ?", id).Updates(updates).Error
}

// GetURLByID retrieves a URL by ID
func GetURLByID(dbConn *gorm.DB, id uint) (*db.URL, error) {
	var url db.URL
	err := dbConn.First(&url, id).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}

// GetURLByIDAndUser retrieves a URL by ID for a specific user
func GetURLByIDAndUser(dbConn *gorm.DB, id uint, userID uint) (*db.URL, error) {
	var url db.URL
	err := dbConn.Where("id = ? AND user_id = ?", id, userID).First(&url).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}

// CreateURL creates a new URL with address for a specific user
func CreateURL(dbConn *gorm.DB, userID uint, address string) (*db.URL, error) {
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}
	if userID == 0 {
		return nil, fmt.Errorf("user ID cannot be zero")
	}

	url := db.URL{
		UserID:  userID,
		Address: address,
		Status:  db.StatusQueued,
	}

	err := dbConn.Create(&url).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}

// CreateURLLegacy creates a new URL without user association (for backward compatibility)
func CreateURLLegacy(dbConn *gorm.DB, address string) (*db.URL, error) {
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	url := db.URL{
		Address: address,
		Status:  db.StatusQueued,
	}

	err := dbConn.Create(&url).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}

// GetURLByAddress retrieves a URL by address for a specific user
func GetURLByAddress(dbConn *gorm.DB, userID uint, address string) (*db.URL, error) {
	var url db.URL
	err := dbConn.Where("user_id = ? AND address = ?", userID, address).First(&url).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}

// GetURLByAddressLegacy retrieves a URL by address without user filtering (for backward compatibility)
func GetURLByAddressLegacy(dbConn *gorm.DB, address string) (*db.URL, error) {
	var url db.URL
	err := dbConn.Where("address = ?", address).First(&url).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
} 