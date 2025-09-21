package main

import (
	"time"
)

// ShortURL represents a shortened URL entry
type ShortURL struct {
	ShortCode    string    `json:"shortcode"`
	OriginalURL  string    `json:"original_url"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	ClickCount   int       `json:"click_count"`
	ClickHistory []Click   `json:"click_history"`
}

// Click represents a click event on a short URL
type Click struct {
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	Location  string    `json:"location"`
}

// CreateShortURLRequest represents the request to create a short URL
type CreateShortURLRequest struct {
	URL       string `json:"url"`
	Validity  int    `json:"validity,omitempty"`
	ShortCode string `json:"shortcode,omitempty"`
}

// CreateShortURLResponse represents the response for creating a short URL
type CreateShortURLResponse struct {
	ShortLink string `json:"shortLink"`
	Expiry    string `json:"expiry"`
}

// ShortURLStats represents statistics for a short URL
type ShortURLStats struct {
	TotalClicks int       `json:"totalClicks"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Clicks      []Click   `json:"clicks"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
