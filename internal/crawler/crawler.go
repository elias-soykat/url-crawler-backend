package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gorm.io/gorm"

	"github.com/sykell/url-crawler/internal/db"
)

// Service represents the crawler service
type Service struct {
	db       *gorm.DB
	queue    chan uint
	workers  int
	timeout  time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.RWMutex
	isRunning bool
}

// Config holds crawler configuration
type Config struct {
	Workers     int
	QueueSize   int
	Timeout     time.Duration
	MaxRetries  int
}

// DefaultConfig returns default crawler configuration
func DefaultConfig() *Config {
	return &Config{
		Workers:    5,
		QueueSize:  100,
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}
}

// NewService creates a new crawler service
func NewService(db *gorm.DB, config *Config) *Service {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	return &Service{
		db:      db,
		queue:   make(chan uint, config.QueueSize),
		workers: config.Workers,
		timeout: config.Timeout,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start starts the crawler service
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.isRunning {
		return fmt.Errorf("crawler service is already running")
	}

	s.isRunning = true
	
	// Start worker goroutines
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	log.Printf("Crawler service started with %d workers", s.workers)
	return nil
}

// Stop stops the crawler service gracefully
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.isRunning {
		return nil
	}

	s.isRunning = false
	s.cancel()
	close(s.queue)
	
	// Wait for all workers to finish
	s.wg.Wait()
	
	log.Println("Crawler service stopped")
	return nil
}

// NotifyNewURL adds a URL to the processing queue
func (s *Service) NotifyNewURL(id uint) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if !s.isRunning {
		return fmt.Errorf("crawler service is not running")
	}

	select {
	case s.queue <- id:
		return nil
	default:
		return fmt.Errorf("queue is full")
	}
}

// worker processes URLs from the queue
func (s *Service) worker(id int) {
	defer s.wg.Done()
	
	log.Printf("Worker %d started", id)
	
	for {
		select {
		case urlID, ok := <-s.queue:
			if !ok {
				log.Printf("Worker %d shutting down", id)
				return
			}
			s.processURL(urlID)
		case <-s.ctx.Done():
			log.Printf("Worker %d shutting down", id)
			return
		}
	}
}

// processURL processes a single URL
func (s *Service) processURL(id uint) {
	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	// Get URL from database
	url, err := db.GetURLByID(s.db, id)
	if err != nil {
		log.Printf("Failed to get URL %d: %v", id, err)
		return
	}

	// Check if URL is still queued
	if url.Status != db.StatusQueued {
		log.Printf("URL %d is not in queued status: %s", id, url.Status)
		return
	}

	// Update status to running
	if err := db.UpdateURLStatus(s.db, id, db.StatusRunning, ""); err != nil {
		log.Printf("Failed to update URL %d status to running: %v", id, err)
		return
	}

	// Crawl the URL
	result, err := s.crawlWithContext(ctx, url.Address)
	if err != nil {
		log.Printf("Failed to crawl URL %d (%s): %v", id, url.Address, err)
		if updateErr := db.UpdateURLStatus(s.db, id, db.StatusError, err.Error()); updateErr != nil {
			log.Printf("Failed to update URL %d error status: %v", id, updateErr)
		}
		return
	}

	// Update URL with results
	if err := s.updateURLWithResults(id, result); err != nil {
		log.Printf("Failed to update URL %d with results: %v", id, err)
		if updateErr := db.UpdateURLStatus(s.db, id, db.StatusError, err.Error()); updateErr != nil {
			log.Printf("Failed to update URL %d error status: %v", id, updateErr)
		}
		return
	}

	log.Printf("Successfully processed URL %d (%s)", id, url.Address)
}

// crawlWithContext crawls a URL with context support
func (s *Service) crawlWithContext(ctx context.Context, address string) (*CrawlResult, error) {
	client := &http.Client{
		Timeout: s.timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "URL-Crawler/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return s.parseDocument(doc, address)
}

// parseDocument parses the HTML document and extracts information
func (s *Service) parseDocument(doc *goquery.Document, baseAddress string) (*CrawlResult, error) {
	baseURL, err := url.Parse(baseAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	result := &CrawlResult{
		Title:         strings.TrimSpace(doc.Find("title").Text()),
		HTMLVersion:   s.detectHTMLVersion(doc),
		HeadingCounts: s.countHeadings(doc),
		HasLoginForm:  s.detectLoginForm(doc),
	}

	// Analyze links
	internal, external, brokenLinks := s.analyzeLinks(doc, baseURL)
	result.InternalLinks = internal
	result.ExternalLinks = external
	result.BrokenList = brokenLinks

	return result, nil
}

// detectHTMLVersion detects the HTML version
func (s *Service) detectHTMLVersion(doc *goquery.Document) string {
	if doc.Find("!DOCTYPE").Length() > 0 {
		return "HTML5"
	}
	if strings.Contains(strings.ToLower(doc.Text()), "xhtml") {
		return "XHTML"
	}
	return "HTML5"
}

// countHeadings counts heading elements
func (s *Service) countHeadings(doc *goquery.Document) map[string]int {
	headings := make(map[string]int)
	for i := 1; i <= 6; i++ {
		tag := "h" + strconv.Itoa(i)
		headings[tag] = doc.Find(tag).Length()
	}
	return headings
}

// detectLoginForm detects if the page has a login form
func (s *Service) detectLoginForm(doc *goquery.Document) bool {
	return doc.Find("form input[type='password']").Length() > 0
}

// analyzeLinks analyzes all links on the page
func (s *Service) analyzeLinks(doc *goquery.Document, baseURL *url.URL) (internal, external int, brokenLinks []map[string]string) {
	linksChecked := make(map[string]bool)
	
	doc.Find("a[href]").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		if !exists || href == "" || linksChecked[href] {
			return
		}
		linksChecked[href] = true

		linkURL, err := url.Parse(href)
		if err != nil {
			return
		}

		// Resolve relative URLs
		if !linkURL.IsAbs() {
			linkURL = baseURL.ResolveReference(linkURL)
		}

		// Determine link type
		if linkURL.Host != "" && linkURL.Host != baseURL.Host {
			external++
		} else {
			internal++
		}

		// Check link accessibility for absolute URLs
		if linkURL.Scheme != "" {
			if code := s.checkLink(linkURL.String()); code >= 400 {
				brokenLinks = append(brokenLinks, map[string]string{
					"url":  linkURL.String(),
					"code": strconv.Itoa(code),
				})
			}
		}
	})

	return internal, external, brokenLinks
}

// checkLink checks if a link is accessible
func (s *Service) checkLink(link string) int {
	client := &http.Client{Timeout: 10 * time.Second}
	
	req, err := http.NewRequest("HEAD", link, nil)
	if err != nil {
		return 500
	}
	
	req.Header.Set("User-Agent", "URL-Crawler/1.0")
	
	resp, err := client.Do(req)
	if err != nil {
		return 500
	}
	defer resp.Body.Close()
	
	return resp.StatusCode
}

// updateURLWithResults updates the URL record with crawl results
func (s *Service) updateURLWithResults(id uint, result *CrawlResult) error {
	brokenListJSON, err := json.Marshal(result.BrokenList)
	if err != nil {
		return fmt.Errorf("failed to marshal broken list: %w", err)
	}

	headingsJSON, err := json.Marshal(result.HeadingCounts)
	if err != nil {
		return fmt.Errorf("failed to marshal heading counts: %w", err)
	}

	updates := map[string]interface{}{
		"title":          result.Title,
		"html_version":   result.HTMLVersion,
		"heading_counts": string(headingsJSON),
		"internal_links": result.InternalLinks,
		"external_links": result.ExternalLinks,
		"broken_links":   len(result.BrokenList),
		"broken_list":    string(brokenListJSON),
		"has_login_form": result.HasLoginForm,
		"status":         db.StatusDone,
		"error":          "",
	}

	return s.db.Model(&db.URL{}).Where("id = ?", id).Updates(updates).Error
}

// CrawlResult represents the result of crawling a URL
type CrawlResult struct {
	Title         string              `json:"title"`
	HTMLVersion   string              `json:"html_version"`
	HeadingCounts map[string]int      `json:"heading_counts"`
	InternalLinks int                 `json:"internal_links"`
	ExternalLinks int                 `json:"external_links"`
	BrokenList    []map[string]string `json:"broken_list"`
	HasLoginForm  bool                `json:"has_login_form"`
}