package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Stack string
type Level string
type Package string

const (
	BackendStack      Stack   = "backend"
	FrontendStack     Stack   = "frontend"
	DebugLevel        Level   = "debug"
	InfoLevel         Level   = "info"
	WarnLevel         Level   = "warn"
	ErrorLevel        Level   = "error"
	FatalLevel        Level   = "fatal"
	CachePackage      Package = "cache"
	ControllerPackage Package = "controller"
	CronJobPackage    Package = "cron_job"
	DbPackage         Package = "db"
	DomainPackage     Package = "domain"
	HandlerPackage    Package = "handler"
	RepositoryPackage Package = "repository"
	RoutePackage      Package = "route"
	ServicePackage    Package = "service"
)

type LogEntry struct {
	Stack   Stack   `json:"stack"`
	Level   Level   `json:"level"`
	Package Package `json:"package"`
	Message string  `json:"message"`
	Time    string  `json:"time"`
}

type Logger struct {
	serverURL string
	client    *http.Client
}

func NewLogger(serverURL string) *Logger {
	return &Logger{
		serverURL: serverURL,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (l *Logger) Log(stack Stack, level Level, pkg Package, message string) error {
	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	jsonData, _ := json.Marshal(LogEntry{
		Stack:   stack,
		Level:   level,
		Package: pkg,
		Message: message,
		Time:    time.Now().Format(time.RFC3339),
	})

	req, _ := http.NewRequest("POST", l.serverURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJNYXBDbGFpbXMiOnsiYXVkIjoiaHR0cDovLzIwLjI0NC41Ni4xNDQvZXZhbHVhdGlvbi1zZXJ2aWNlIiwiZW1haWwiOiIyMmNzMzA0OEByZ2lwdC5hYy5pbiIsImV4cCI6MTc1ODQ0OTMwOCwiaWF0IjoxNzU4NDQ4NDA4LCJpc3MiOiJBZmZvcmQgTWVkaWNhbCBUZWNobm9sb2dpZXMgUHJpdmF0ZSBMaW1pdGVkIiwianRpIjoiNzgyODA2NzYtZTliZC00NGIzLWIzNmQtMTg5NmMzNjNkM2EzIiwibG9jYWxlIjoiZW4tSU4iLCJuYW1lIjoicmlzaGFiaCB0cmlwYXRoaSIsInN1YiI6IjdlMjZmZDlkLWJjMDQtNDM5My04ZTIyLTFiNjJiYjJjY2RlNCJ9LCJlbWFpbCI6IjIyY3MzMDQ4QHJnaXB0LmFjLmluIiwibmFtZSI6InJpc2hhYmggdHJpcGF0aGkiLCJyb2xsTm8iOiIyMmNzMzA0OCIsImFjY2Vzc0NvZGUiOiJhcnpVY0ciLCJjbGllbnRJRCI6IjdlMjZmZDlkLWJjMDQtNDM5My04ZTIyLTFiNjJiYjJjY2RlNCIsImNsaWVudFNlY3JldCI6InpVcFd1WFlrYWpmdWdQTlEifQ.zSMCZUtLiq0TmOx61nRM9e_jT69aMIg2DgVkV0I-E-4")

	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func LoggingMiddleware(logger *Logger, stack Stack, pkg Package) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Log(stack, InfoLevel, pkg, fmt.Sprintf("%s %s", r.Method, r.URL.Path))
			next.ServeHTTP(w, r)
		})
	}
}
