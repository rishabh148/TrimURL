package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

// URLService handles URL shortening operations
type URLService struct {
	urls   map[string]*ShortURL
	mutex  sync.RWMutex
	logger *Logger
}

// NewURLService creates a new URL service
func NewURLService(logger *Logger) *URLService {
	return &URLService{
		urls:   make(map[string]*ShortURL),
		logger: logger,
	}
}

// CreateShortURL creates a new shortened URL
func (s *URLService) CreateShortURL(req CreateShortURLRequest) (*CreateShortURLResponse, error) {
	s.logger.Log(BackendStack, InfoLevel, ServicePackage, "Creating short URL")

	// Validate URL
	if err := s.validateURL(req.URL); err != nil {
		s.logger.Log(BackendStack, ErrorLevel, DomainPackage, fmt.Sprintf("Invalid URL: %v", err))
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	// Set default validity to 30 minutes
	validity := req.Validity
	if validity <= 0 {
		validity = 30
	}

	s.logger.Log(BackendStack, DebugLevel, ServicePackage, fmt.Sprintf("URL validity set to %d minutes", validity))

	// Generate or validate shortcode
	shortCode := req.ShortCode
	if shortCode == "" {
		shortCode = s.generateShortCode()
		s.logger.Log(BackendStack, DebugLevel, ServicePackage, fmt.Sprintf("Generated shortcode: %s", shortCode))
	} else {
		if err := s.validateShortCode(shortCode); err != nil {
			s.logger.Log(BackendStack, ErrorLevel, DomainPackage, fmt.Sprintf("Invalid shortcode: %v", err))
			return nil, fmt.Errorf("invalid shortcode: %v", err)
		}

		// Check if shortcode already exists
		if s.shortCodeExists(shortCode) {
			s.logger.Log(BackendStack, ErrorLevel, DomainPackage, fmt.Sprintf("Shortcode collision: %s", shortCode))
			return nil, fmt.Errorf("shortcode already exists")
		}
	}

	// Create short URL entry
	now := time.Now()
	shortURL := &ShortURL{
		ShortCode:    shortCode,
		OriginalURL:  req.URL,
		CreatedAt:    now,
		ExpiresAt:    now.Add(time.Duration(validity) * time.Minute),
		ClickCount:   0,
		ClickHistory: []Click{},
	}

	// Store the short URL
	s.mutex.Lock()
	s.urls[shortCode] = shortURL
	s.mutex.Unlock()

	s.logger.Log(BackendStack, InfoLevel, ServicePackage, fmt.Sprintf("Short URL created: %s -> %s", shortCode, req.URL))

	return &CreateShortURLResponse{
		ShortLink: fmt.Sprintf("http://localhost:3000/%s", shortCode),
		Expiry:    shortURL.ExpiresAt.Format(time.RFC3339),
	}, nil
}

// GetOriginalURL retrieves the original URL for a short code
func (s *URLService) GetOriginalURL(shortCode string) (string, error) {
	s.logger.Log(BackendStack, InfoLevel, ServicePackage, fmt.Sprintf("Retrieving original URL for: %s", shortCode))

	s.mutex.RLock()
	shortURL, exists := s.urls[shortCode]
	s.mutex.RUnlock()

	if !exists {
		s.logger.Log(BackendStack, ErrorLevel, DomainPackage, fmt.Sprintf("Shortcode not found: %s", shortCode))
		return "", fmt.Errorf("shortcode not found")
	}

	// Check if expired
	if time.Now().After(shortURL.ExpiresAt) {
		s.logger.Log(BackendStack, WarnLevel, DomainPackage, fmt.Sprintf("Shortcode expired: %s", shortCode))
		return "", fmt.Errorf("shortcode expired")
	}

	return shortURL.OriginalURL, nil
}

// RecordClick records a click on a short URL
func (s *URLService) RecordClick(shortCode, source, location string) error {
	s.logger.Log(BackendStack, DebugLevel, ServicePackage, fmt.Sprintf("Recording click for: %s", shortCode))

	s.mutex.Lock()
	defer s.mutex.Unlock()

	shortURL, exists := s.urls[shortCode]
	if !exists {
		return fmt.Errorf("shortcode not found")
	}

	// Record the click
	click := Click{
		Timestamp: time.Now(),
		Source:    source,
		Location:  location,
	}

	shortURL.ClickCount++
	shortURL.ClickHistory = append(shortURL.ClickHistory, click)

	s.logger.Log(BackendStack, InfoLevel, ServicePackage, fmt.Sprintf("Click recorded for %s (total: %d)", shortCode, shortURL.ClickCount))

	return nil
}

// GetStats retrieves statistics for a short URL
func (s *URLService) GetStats(shortCode string) (*ShortURLStats, error) {
	s.logger.Log(BackendStack, InfoLevel, ServicePackage, fmt.Sprintf("Retrieving stats for: %s", shortCode))

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	shortURL, exists := s.urls[shortCode]
	if !exists {
		s.logger.Log(BackendStack, ErrorLevel, DomainPackage, fmt.Sprintf("Shortcode not found for stats: %s", shortCode))
		return nil, fmt.Errorf("shortcode not found")
	}

	return &ShortURLStats{
		TotalClicks: shortURL.ClickCount,
		CreatedAt:   shortURL.CreatedAt,
		ExpiresAt:   shortURL.ExpiresAt,
		Clicks:      shortURL.ClickHistory,
	}, nil
}

// validateURL validates if a URL is properly formatted
func (s *URLService) validateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Add protocol if missing
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	_, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}

	return nil
}

// validateShortCode validates if a shortcode is valid
func (s *URLService) validateShortCode(shortCode string) error {
	if len(shortCode) < 4 || len(shortCode) > 20 {
		return fmt.Errorf("shortcode must be 4-20 characters")
	}

	// Check if alphanumeric
	for _, char := range shortCode {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return fmt.Errorf("shortcode must be alphanumeric")
		}
	}

	return nil
}

// generateShortCode generates a unique shortcode
func (s *URLService) generateShortCode() string {
	for {
		bytes := make([]byte, 4)
		rand.Read(bytes)
		shortCode := hex.EncodeToString(bytes)[:8]

		if !s.shortCodeExists(shortCode) {
			return shortCode
		}
	}
}

// shortCodeExists checks if a shortcode already exists
func (s *URLService) shortCodeExists(shortCode string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, exists := s.urls[shortCode]
	return exists
}
