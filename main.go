package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Initialize logger
	logger := NewLogger("http://20.244.56.144/evaluation-service/logs")

	// Test connection
	if err := logger.Log(BackendStack, InfoLevel, ServicePackage, "URL Shortener service starting"); err != nil {
		fmt.Printf("Failed to connect to logging server: %v\n", err)
		fmt.Println("Continuing without logging...")
	} else {
		fmt.Println("Connected to logging server!")
	}

	// Initialize URL service
	urlService := NewURLService(logger)
	logger.Log(BackendStack, InfoLevel, ServicePackage, "URL service initialized")

	// Initialize handlers
	urlHandler := NewURLHandler(urlService, logger)
	logger.Log(BackendStack, InfoLevel, HandlerPackage, "URL handlers initialized")

	// Set up routes (order matters - specific routes first)
	http.Handle("/health", LoggingMiddleware(logger, BackendStack, RoutePackage)(http.HandlerFunc(urlHandler.HealthCheck)))
	http.Handle("/shorturls/", LoggingMiddleware(logger, BackendStack, RoutePackage)(http.HandlerFunc(urlHandler.GetStats)))
	http.Handle("/shorturls", LoggingMiddleware(logger, BackendStack, RoutePackage)(http.HandlerFunc(urlHandler.CreateShortURL)))
	http.Handle("/", LoggingMiddleware(logger, BackendStack, RoutePackage)(http.HandlerFunc(urlHandler.RedirectURL)))

	// Start server
	port := "3000"
	logger.Log(BackendStack, InfoLevel, ServicePackage, fmt.Sprintf("Starting server on port %s", port))

	fmt.Printf("URL Shortener Service starting on port %s\n", port)
	fmt.Printf("API Endpoints:\n")
	fmt.Printf("POST   http://localhost:%s/shorturls     - Create short URL\n", port)
	fmt.Printf("GET    http://localhost:%s/shorturls/:id - Get statistics\n", port)
	fmt.Printf("GET    http://localhost:%s/health        - Health check\n", port)
	fmt.Printf("GET    http://localhost:%s/:shortcode    - Redirect to original URL\n", port)
	fmt.Printf("\nAll operations are logged to the evaluation server\n")

	// Start server in background
	go func() {
		logger.Log(BackendStack, InfoLevel, ServicePackage, "HTTP server started")
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	fmt.Println("\nPress Ctrl+C to stop the server...")
	<-c

	logger.Log(BackendStack, InfoLevel, ServicePackage, "Server shutting down")
	fmt.Println("\nShutting down URL Shortener Service...")
}
