package db

import (
	"log"

	"gorm.io/gorm"
)

// runMigrations performs database migrations
func runMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(&User{}, &URL{}); err != nil {
		return err
	}
	
	// Handle existing URLs that don't have a user_id
	return migrateExistingURLs(db)
}

// migrateExistingURLs assigns existing URLs without user_id to the first admin user
func migrateExistingURLs(db *gorm.DB) error {
	// Check if there are any URLs without user_id
	var count int64
	if err := db.Model(&URL{}).Where("user_id = 0 OR user_id IS NULL").Count(&count).Error; err != nil {
		return err
	}
	
	if count == 0 {
		return nil // No URLs need migration
	}
	
	// Find the first admin user to assign orphaned URLs to
	var adminUser User
	if err := db.First(&adminUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// No users exist yet, that's fine
			return nil
		}
		return err
	}
	
	// Update orphaned URLs to belong to the admin user
	result := db.Model(&URL{}).Where("user_id = 0 OR user_id IS NULL").Update("user_id", adminUser.ID)
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected > 0 {
		log.Printf("Migrated %d orphaned URLs to user %d (%s)", result.RowsAffected, adminUser.ID, adminUser.Username)
	}
	
	return nil
} 