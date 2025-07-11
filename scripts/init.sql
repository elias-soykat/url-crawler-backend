-- URL Crawler Database Initialization Script
-- This script sets up the initial database schema and configuration

-- Create database if it doesn't exist
CREATE DATABASE IF NOT EXISTS url_crawler CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Use the database
USE url_crawler;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create urls table
CREATE TABLE IF NOT EXISTS urls (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    title VARCHAR(500),
    html_version VARCHAR(10),
    heading_counts TEXT,
    internal_links INT DEFAULT 0,
    external_links INT DEFAULT 0,
    broken_links INT DEFAULT 0,
    broken_list TEXT,
    has_login_form BOOLEAN DEFAULT FALSE,
    status VARCHAR(191) DEFAULT 'queued',
    error VARCHAR(1000),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_urls_address (address(191)),
    INDEX idx_urls_status (status),
    INDEX idx_urls_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default admin user (password will be hashed by the application)
-- Note: This is just a placeholder, the actual user creation is handled by the seed script
INSERT IGNORE INTO users (username, password) VALUES 
('admin', '$2a$10$placeholder.hash.will.be.replaced.by.seed.script');

-- Create indexes for better performance
-- (Indexes are already defined in the table creation above)

-- Set up proper permissions for the application user
-- (This will be handled by the MySQL container environment variables)

-- Log the initialization
SELECT 'Database initialization completed successfully' as status; 