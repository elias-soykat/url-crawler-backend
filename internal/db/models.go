package db

import "time"

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
	UserID        uint      `gorm:"index" json:"user_id"`
	Address       string    `gorm:"not null;size:768" json:"address"`
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
	User          User      `gorm:"foreignKey:UserID" json:"-"`
}

// User represents an authenticated user
type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null;size:100" json:"username"`
	Password  string    `gorm:"not null;size:255" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
} 