package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// URLHandler handles HTTP requests for URL shortening
type URLHandler struct {
	urlService *URLService
	logger     *Logger
}

// NewURLHandler creates a new URL handler
func NewURLHandler(urlService *URLService, logger *Logger) *URLHandler {
	return &URLHandler{
		urlService: urlService,
		logger:     logger,
	}
}

// CreateShortURL handles POST /shorturls
func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	h.logger.Log(BackendStack, InfoLevel, HandlerPackage, "POST /shorturls - Creating short URL")

	if r.Method != "POST" {
		h.logger.Log(BackendStack, ErrorLevel, HandlerPackage, "Invalid method for /shorturls")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the raw body for debugging
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Log(BackendStack, ErrorLevel, HandlerPackage, fmt.Sprintf("Failed to read body: %v", err))
		h.sendErrorResponse(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	h.logger.Log(BackendStack, DebugLevel, HandlerPackage, fmt.Sprintf("Received body: %s", string(body)))

	var req CreateShortURLRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Log(BackendStack, ErrorLevel, HandlerPackage, fmt.Sprintf("Invalid JSON: %v", err))
		h.sendErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.URL == "" {
		h.logger.Log(BackendStack, ErrorLevel, HandlerPackage, "Missing URL field")
		h.sendErrorResponse(w, "URL is required", http.StatusBadRequest)
		return
	}

	h.logger.Log(BackendStack, DebugLevel, HandlerPackage, fmt.Sprintf("Processing URL: %s", req.URL))

	// Create short URL
	resp, err := h.urlService.CreateShortURL(req)
	if err != nil {
		h.logger.Log(BackendStack, ErrorLevel, HandlerPackage, fmt.Sprintf("Failed to create short URL: %v", err))
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Log(BackendStack, InfoLevel, HandlerPackage, fmt.Sprintf("Short URL created successfully: %s", resp.ShortLink))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// RedirectURL handles GET /:shortcode (redirect)
func (h *URLHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/")
	h.logger.Log(BackendStack, InfoLevel, HandlerPackage, fmt.Sprintf("GET /%s - Redirecting", shortCode))

	// Skip API endpoints
	if shortCode == "shorturls" || shortCode == "health" {
		return
	}

	// Get original URL
	originalURL, err := h.urlService.GetOriginalURL(shortCode)
	if err != nil {
		h.logger.Log(BackendStack, ErrorLevel, HandlerPackage, fmt.Sprintf("Redirect failed for %s: %v", shortCode, err))
		h.sendErrorResponse(w, "Short URL not found or expired", http.StatusNotFound)
		return
	}

	// Record click
	source := r.Header.Get("Referer")
	if source == "" {
		source = "direct"
	}
	location := "unknown" // In a real app, you'd use IP geolocation

	if err := h.urlService.RecordClick(shortCode, source, location); err != nil {
		h.logger.Log(BackendStack, WarnLevel, HandlerPackage, fmt.Sprintf("Failed to record click: %v", err))
	}

	h.logger.Log(BackendStack, InfoLevel, HandlerPackage, fmt.Sprintf("Redirecting %s -> %s", shortCode, originalURL))

	// Redirect to original URL
	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

// GetStats handles GET /shorturls/:shortcode
func (h *URLHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	shortCode := strings.TrimPrefix(path, "/shorturls/")

	h.logger.Log(BackendStack, InfoLevel, HandlerPackage, fmt.Sprintf("GET /shorturls/%s - Getting stats", shortCode))

	if shortCode == "" {
		h.logger.Log(BackendStack, ErrorLevel, HandlerPackage, "Missing shortcode in stats request")
		h.sendErrorResponse(w, "Shortcode is required", http.StatusBadRequest)
		return
	}

	// Get statistics
	stats, err := h.urlService.GetStats(shortCode)
	if err != nil {
		h.logger.Log(BackendStack, ErrorLevel, HandlerPackage, fmt.Sprintf("Failed to get stats for %s: %v", shortCode, err))
		h.sendErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	h.logger.Log(BackendStack, InfoLevel, HandlerPackage, fmt.Sprintf("Stats retrieved for %s: %d clicks", shortCode, stats.TotalClicks))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// HealthCheck handles GET /health
func (h *URLHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.logger.Log(BackendStack, DebugLevel, HandlerPackage, "GET /health - Health check")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"message": "URL Shortener service is running",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// sendErrorResponse sends a JSON error response
func (h *URLHandler) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	errorResp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResp)
}
