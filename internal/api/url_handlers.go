package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/sykell/url-crawler/internal/crawler"
	"github.com/sykell/url-crawler/internal/db"
)

// PostURLRequest represents the URL creation request
type PostURLRequest struct {
	Address string `json:"address" binding:"required,url"`
}

// URLResponse represents a URL response
type URLResponse struct {
	ID            uint      `json:"id"`
	Address       string    `json:"address"`
	Title         string    `json:"title"`
	HTMLVersion   string    `json:"html_version"`
	HeadingCounts string    `json:"heading_counts"`
	InternalLinks int       `json:"internal_links"`
	ExternalLinks int       `json:"external_links"`
	BrokenLinks   int       `json:"broken_links"`
	BrokenList    string    `json:"broken_list"`
	HasLoginForm  bool      `json:"has_login_form"`
	Status        string    `json:"status"`
	Error         string    `json:"error"`
	CreatedAt     string    `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
}

// URLDetailResponse represents a detailed URL response
type URLDetailResponse struct {
	URLResponse
	HeadingCounts map[string]int      `json:"heading_counts"`
	BrokenList    []map[string]string `json:"broken_list"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data  interface{} `json:"data"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Total int64       `json:"total"`
	Pages int         `json:"pages"`
}

// BulkRequest represents a bulk operation request
type BulkRequest struct {
	Action string `json:"action" binding:"required,oneof=rerun delete"`
	IDs    []uint `json:"ids" binding:"required,min=1,max=100"`
}

// PostURLHandler handles URL creation
func PostURLHandler(dbConn *gorm.DB, crawlerService *crawler.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PostURLRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("URL creation validation error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid URL format",
				"details": err.Error(),
			})
			return
		}

		// Sanitize and validate URL
		req.Address = strings.TrimSpace(req.Address)
		if req.Address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL cannot be empty"})
			return
		}

		// Check if URL already exists
		existingURL, err := db.GetURLByAddress(dbConn, req.Address)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "URL already exists", "id": existingURL.ID})
			return
		} else if err != gorm.ErrRecordNotFound {
			log.Printf("Database error checking existing URL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		

		// Create new URL
		url, err := db.CreateURL(dbConn, req.Address)
		if err != nil {
			log.Printf("Failed to create URL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL"})
			return
		}

		// Notify crawler service
		if err := crawlerService.NotifyNewURL(url.ID); err != nil {
			log.Printf("Failed to notify crawler service: %v", err)
			// Don't fail the request, just log the error
		}

		log.Printf("Created new URL: %s (ID: %d)", req.Address, url.ID)
		c.JSON(http.StatusCreated, url)
	}
}

// ListURLsHandler handles URL listing with pagination and search
func ListURLsHandler(dbConn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse pagination parameters
		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 {
			page = 1
		}

		pageSize, err := strconv.Atoi(c.DefaultQuery("size", "10"))
		if err != nil || pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		// Parse sort parameter
		sort := c.DefaultQuery("sort", "created_at desc")
		allowedSorts := map[string]bool{
			"created_at desc": true,
			"created_at asc":  true,
			"updated_at desc": true,
			"updated_at asc":  true,
			"status asc":      true,
			"status desc":     true,
		}
		if !allowedSorts[sort] {
			sort = "created_at desc"
		}

		// Parse search parameter
		search := strings.TrimSpace(c.Query("q"))
		status := strings.TrimSpace(c.Query("status"))

		// Build query
		query := dbConn.Model(&db.URL{})
		
		if search != "" {
			query = query.Where("address LIKE ? OR title LIKE ?", "%"+search+"%", "%"+search+"%")
		}
		
		if status != "" {
			query = query.Where("status = ?", status)
		}

		// Get total count
		var total int64
		if err := query.Count(&total).Error; err != nil {
			log.Printf("Failed to count URLs: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Calculate pagination
		offset := (page - 1) * pageSize
		pages := int((total + int64(pageSize) - 1) / int64(pageSize))

		// Get URLs
		var urls []db.URL
		if err := query.Order(sort).Limit(pageSize).Offset(offset).Find(&urls).Error; err != nil {
			log.Printf("Failed to fetch URLs: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		response := PaginatedResponse{
			Data:  urls,
			Page:  page,
			Size:  pageSize,
			Total: total,
			Pages: pages,
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetURLHandler handles retrieving a single URL
func GetURLHandler(dbConn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL ID"})
			return
		}

		url, err := db.GetURLByID(dbConn, uint(id))
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
				return
			}
			log.Printf("Failed to fetch URL %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Parse JSON fields for detailed response
		var headingCounts map[string]int
		var brokenList []map[string]string

		if url.HeadingCounts != "" {
			if err := json.Unmarshal([]byte(url.HeadingCounts), &headingCounts); err != nil {
				log.Printf("Failed to parse heading counts for URL %d: %v", id, err)
			}
		}

		if url.BrokenList != "" {
			if err := json.Unmarshal([]byte(url.BrokenList), &brokenList); err != nil {
				log.Printf("Failed to parse broken list for URL %d: %v", id, err)
			}
		}

		detail := URLDetailResponse{
			URLResponse: URLResponse{
				ID:            url.ID,
				Address:       url.Address,
				Title:         url.Title,
				HTMLVersion:   url.HTMLVersion,
				HeadingCounts: url.HeadingCounts,
				InternalLinks: url.InternalLinks,
				ExternalLinks: url.ExternalLinks,
				BrokenLinks:   url.BrokenLinks,
				BrokenList:    url.BrokenList,
				HasLoginForm:  url.HasLoginForm,
				Status:        string(url.Status),
				Error:         url.Error,
				CreatedAt:     url.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:     url.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			},
			HeadingCounts: headingCounts,
			BrokenList:    brokenList,
		}

		c.JSON(http.StatusOK, detail)
	}
}

// BulkHandler handles bulk operations on URLs
func BulkHandler(dbConn *gorm.DB, crawlerService *crawler.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BulkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Bulk operation validation error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid bulk request",
				"details": err.Error(),
			})
			return
		}

		// Validate IDs
		if len(req.IDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No IDs provided"})
			return
		}

		var affected int64
		var err error

		switch req.Action {
		case "rerun":
			// Reset URLs to queued status
			result := dbConn.Model(&db.URL{}).Where("id IN ?", req.IDs).Updates(map[string]interface{}{
				"status": db.StatusQueued,
				"error":  "",
			})
			affected = result.RowsAffected
			err = result.Error

			if err == nil && affected > 0 {
				// Notify crawler service for each URL
				for _, id := range req.IDs {
					if notifyErr := crawlerService.NotifyNewURL(id); notifyErr != nil {
						log.Printf("Failed to notify crawler for URL %d: %v", id, notifyErr)
					}
				}
			}

		case "delete":
			// Delete URLs
			result := dbConn.Delete(&db.URL{}, req.IDs)
			affected = result.RowsAffected
			err = result.Error

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
			return
		}

		if err != nil {
			log.Printf("Bulk operation failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform bulk operation"})
			return
		}

		log.Printf("Bulk %s operation completed: %d URLs affected", req.Action, affected)
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"action":   req.Action,
			"affected": affected,
		})
	}
}