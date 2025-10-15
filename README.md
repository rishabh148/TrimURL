TrimURL - URL Shortener Service

A fast and lightweight URL shortening service built with Go that allows you to create short URLs with expiration times and track click statistics.

Features

- URL Shortening: Create short URLs from long URLs with customizable expiration times
- Custom Short Codes: Option to provide custom short codes (4-20 alphanumeric characters)
- Automatic Expiration: URLs expire after a specified time (default: 30 minutes)
- Click Tracking: Track clicks with source and location information
- Statistics: View detailed statistics for each short URL
- Health Monitoring: Built-in health check endpoint
- Comprehensive Logging: All operations are logged to an external logging server
- Thread-Safe: Uses mutex locks for concurrent access safety

API Endpoints

Create Short URL
POST /shorturls

Creates a new short URL from a long URL.

Request Body:
{
  "url": "https://example.com/very/long/url/path",
  "validity": 60,
  "shortcode": "custom123"
}

Response:
{
  "shortLink": "http://localhost:3000/abc12345",
  "expiry": "2024-01-20T15:30:00Z"
}

Get URL Statistics
GET /shorturls/{shortcode}

Retrieves statistics for a specific short URL.

Response:
{
  "totalClicks": 5,
  "createdAt": "2024-01-20T14:30:00Z",
  "expiresAt": "2024-01-20T15:30:00Z",
  "clicks": [
    {
      "timestamp": "2024-01-20T14:35:00Z",
      "source": "https://google.com",
      "location": "unknown"
    }
  ]
}

Redirect to Original URL
GET /{shortcode}

Redirects to the original URL and records the click.

Health Check
GET /health

Returns the service health status.

Response:
{
  "status": "healthy",
  "message": "URL Shortener service is running",
  "time": "2024-01-20T14:30:00Z"
}

Installation & Setup

Prerequisites
- Go 1.21 or higher
- Internet connection (for logging service)

Running the Service

1. Clone the repository:
   git clone https://github.com/rishabh148/TrimURL.git
   cd TrimURL

2. Install dependencies:
   go mod tidy

3. Run the service:
   go run .

The service will start on port 3000 by default.

Usage Examples

Using cURL

Create a short URL:
curl -X POST http://localhost:3000/shorturls \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.google.com/search?q=golang+tutorial",
    "validity": 60
  }'

Get statistics:
curl http://localhost:3000/shorturls/abc12345

Test health:
curl http://localhost:3000/health

Using the Short URL
Simply visit http://localhost:3000/{shortcode} in your browser to be redirected to the original URL.

Configuration

Environment Variables
- The service runs on port 3000 by default
- Logging is configured to send logs to http://20.244.56.144/evaluation-service/logs
- Default URL validity is 30 minutes

Customization
You can modify the following in main.go:
- Port number (line 39)
- Logging server URL (line 14)
- Default validity period (line 41 in url_service.go)

Project Structure

TrimURL/
├── main.go           Application entry point and server setup
├── handlers.go       HTTP request handlers
├── models.go         Data structures and request/response models
├── url_service.go    Business logic for URL operations
├── logger.go         Logging functionality and middleware
├── go.mod           Go module dependencies
└── README.md        This file

Technical Details

Data Storage
- Uses in-memory storage (map with mutex locks)
- Data is lost when the service restarts
- For production use, consider implementing database persistence

URL Validation
- Automatically adds https:// protocol if missing
- Validates URL format using Go's net/url package
- Custom short codes must be 4-20 alphanumeric characters

Security Features
- Thread-safe operations using sync.RWMutex
- Input validation for URLs and short codes
- Bearer token authentication for logging service

Logging
- All operations are logged to an external evaluation server
- Logs include stack, level, package, message, and timestamp
- Graceful degradation if logging service is unavailable

Error Handling

The service returns appropriate HTTP status codes and error messages:

- 400 Bad Request: Invalid input data
- 404 Not Found: Short URL not found or expired
- 405 Method Not Allowed: Wrong HTTP method
- 500 Internal Server Error: Server-side errors

Development

Adding New Features
1. Define data structures in models.go
2. Implement business logic in url_service.go
3. Add HTTP handlers in handlers.go
4. Update routes in main.go

Testing
You can test the service using the provided endpoints with tools like:
- cURL
- Postman
- Any HTTP client library

License

This project is part of an evaluation service and is intended for educational purposes.

Support

For issues or questions, please check the service logs or contact the development team.
