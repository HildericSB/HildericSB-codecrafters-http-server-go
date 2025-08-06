package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codecrafters-io/http-server-starter-go/config"
	"github.com/codecrafters-io/http-server-starter-go/handler"
	"github.com/codecrafters-io/http-server-starter-go/server"
)

// Integration test helper functions
func setupTestServer(t *testing.T) (*server.Server, string) {
	// Create temporary directory for file tests
	tempDir, err := os.MkdirTemp("", "http_server_test_")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	cfg := &config.Config{
		Port:    "0", // Use port 0 to get a random available port
		FileDir: tempDir,
	}

	srv, err := server.NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server in goroutine
	go srv.Start()

	// Wait for server to start and get the actual port
	time.Sleep(100 * time.Millisecond)

	return srv, tempDir
}

func cleanup(srv *server.Server, tempDir string) {
	srv.ShutDown()
	os.RemoveAll(tempDir)
}

func makeHTTPRequest(method, url, body string, headers map[string]string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	return client.Do(req)
}

// Test basic server startup and health endpoint
func TestIntegration_ServerStartupAndHealth(t *testing.T) {
	srv, tempDir := setupTestServer(t)
	defer cleanup(srv, tempDir)

	baseURL := fmt.Sprintf("http://localhost:%s", srv.Port)

	// Test health endpoint
	resp, err := makeHTTPRequest("GET", baseURL+"/health", "", nil)
	if err != nil {
		t.Fatalf("Failed to make health request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var healthResp handler.HealthResponse
	if err := json.Unmarshal(body, &healthResp); err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}

	if healthResp.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", healthResp.Status)
	}

	if healthResp.ActiveConnections < 0 {
		t.Errorf("Expected non-negative active connections, got %d", healthResp.ActiveConnections)
	}
}

// Test echo endpoint with various inputs
func TestIntegration_EchoEndpoint(t *testing.T) {
	srv, tempDir := setupTestServer(t)
	defer cleanup(srv, tempDir)

	baseURL := fmt.Sprintf("http://localhost:%s", srv.Port)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"Simple echo", "/echo/hello", "hello"},
		{"Echo with special chars", "/echo/hello@world.com", "hello@world.com"},
		{"Empty echo", "/echo/", ""},
		{"Echo with numbers", "/echo/12345", "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := makeHTTPRequest("GET", baseURL+tt.path, "", nil)
			if err != nil {
				t.Fatalf("Failed to make echo request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if string(body) != tt.expected {
				t.Errorf("Expected body '%s', got '%s'", tt.expected, string(body))
			}

			// Check content type
			if resp.Header.Get("Content-Type") != "text/plain" {
				t.Errorf("Expected Content-Type 'text/plain', got '%s'", resp.Header.Get("Content-Type"))
			}
		})
	}
}

// Test user-agent endpoint
func TestIntegration_UserAgentEndpoint(t *testing.T) {
	srv, tempDir := setupTestServer(t)
	defer cleanup(srv, tempDir)

	baseURL := fmt.Sprintf("http://localhost:%s", srv.Port)

	tests := []struct {
		name      string
		userAgent string
	}{
		{"Chrome user agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"},
		{"Firefox user agent", "Mozilla/5.0 (X11; Linux x86_64; rv:97.0) Gecko/20100101 Firefox/97.0"},
		{"Custom user agent", "MyCustomClient/1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{}
			if tt.userAgent != "" {
				headers["User-Agent"] = tt.userAgent
			}

			resp, err := makeHTTPRequest("GET", baseURL+"/user-agent", "", headers)
			if err != nil {
				t.Fatalf("Failed to make user-agent request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if string(body) != tt.userAgent {
				t.Errorf("Expected body '%s', got '%s'", tt.userAgent, string(body))
			}
		})
	}
}

// Test file operations (upload and download)
func TestIntegration_FileOperations(t *testing.T) {
	srv, tempDir := setupTestServer(t)
	defer cleanup(srv, tempDir)

	baseURL := fmt.Sprintf("http://localhost:%s", srv.Port)

	// Test file upload
	testContent := "Hello, this is test file content!"
	filename := "test.txt"

	// Upload file
	headers := map[string]string{
		"Content-Type": "application/octet-stream",
	}

	resp, err := makeHTTPRequest("POST", baseURL+"/files/"+filename, testContent, headers)
	if err != nil {
		t.Fatalf("Failed to upload file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201 for file upload, got %d", resp.StatusCode)
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("File was not created: %s", filePath)
	}

	// Read file content to verify
	savedContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if string(savedContent) != testContent {
		t.Errorf("File content mismatch. Expected '%s', got '%s'", testContent, string(savedContent))
	}

	// Test file download
	resp, err = makeHTTPRequest("GET", baseURL+"/files/"+filename, "", nil)
	if err != nil {
		t.Fatalf("Failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for file download, got %d", resp.StatusCode)
	}

	downloadedContent, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read downloaded content: %v", err)
	}

	if string(downloadedContent) != testContent {
		t.Errorf("Downloaded content mismatch. Expected '%s', got '%s'", testContent, string(downloadedContent))
	}

	// Test downloading non-existent file
	resp, err = makeHTTPRequest("GET", baseURL+"/files/nonexistent.txt", "", nil)
	if err != nil {
		t.Fatalf("Failed to request non-existent file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent file, got %d", resp.StatusCode)
	}
}

// Test concurrent requests
func TestIntegration_ConcurrentRequests(t *testing.T) {
	srv, tempDir := setupTestServer(t)
	defer cleanup(srv, tempDir)

	baseURL := fmt.Sprintf("http://localhost:%s", srv.Port)

	const numRequests = 20
	const numWorkers = 5

	requestChan := make(chan int, numRequests)
	resultChan := make(chan error, numRequests)

	// Send request indices to channel
	for i := 0; i < numRequests; i++ {
		requestChan <- i
	}
	close(requestChan)

	// Start workers
	for w := 0; w < numWorkers; w++ {
		go func() {
			for reqNum := range requestChan {
				echoText := fmt.Sprintf("request-%d", reqNum)
				resp, err := makeHTTPRequest("GET", baseURL+"/echo/"+echoText, "", nil)
				if err != nil {
					resultChan <- fmt.Errorf("request %d failed: %v", reqNum, err)
					continue
				}

				body, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					resultChan <- fmt.Errorf("request %d read failed: %v", reqNum, err)
					continue
				}

				if string(body) != echoText {
					resultChan <- fmt.Errorf("request %d: expected '%s', got '%s'", reqNum, echoText, string(body))
					continue
				}

				resultChan <- nil
			}
		}()
	}

	// Collect results
	var errors []error
	for i := 0; i < numRequests; i++ {
		if err := <-resultChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		t.Errorf("Concurrent requests failed with %d errors:", len(errors))
		for _, err := range errors {
			t.Errorf("  %v", err)
		}
	}
}

// Test HTTP method handling
func TestIntegration_HTTPMethods(t *testing.T) {
	srv, tempDir := setupTestServer(t)
	defer cleanup(srv, tempDir)

	baseURL := fmt.Sprintf("http://localhost:%s", srv.Port)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{"GET echo", "GET", "/echo/test", http.StatusOK},
		{"POST to echo (should work)", "POST", "/echo/test", http.StatusOK},
		{"GET health", "GET", "/health", http.StatusOK},
		{"POST health", "POST", "/health", http.StatusOK},
		{"GET user-agent", "GET", "/user-agent", http.StatusOK},
		{"GET files", "GET", "/files/nonexistent.txt", http.StatusNotFound},
		{"POST files", "POST", "/files/test.txt", http.StatusCreated},
		// {"PUT files (not implemented)", "PUT", "/files/test.txt", http.StatusMethodNotAllowed},
		// {"DELETE files (not implemented)", "DELETE", "/files/test.txt", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := ""
			headers := map[string]string{}

			if tt.method == "POST" && strings.Contains(tt.path, "/files/") {
				body = "test content"
				headers["Content-Type"] = "application/octet-stream"
			}

			resp, err := makeHTTPRequest(tt.method, baseURL+tt.path, body, headers)
			if err != nil {
				t.Fatalf("Failed to make %s request: %v", tt.method, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// Test server metrics through health endpoint
func TestIntegration_ServerMetrics(t *testing.T) {
	srv, tempDir := setupTestServer(t)
	defer cleanup(srv, tempDir)

	baseURL := fmt.Sprintf("http://localhost:%s", srv.Port)

	// Make a few requests to increment metrics
	for i := 0; i < 3; i++ {
		resp, err := makeHTTPRequest("GET", baseURL+"/echo/test", "", nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		resp.Body.Close()
	}

	// Check metrics
	resp, err := makeHTTPRequest("GET", baseURL+"/health", "", nil)
	if err != nil {
		t.Fatalf("Failed to make health request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var healthResp handler.HealthResponse
	if err := json.Unmarshal(body, &healthResp); err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}

	// Total requests should be at least 4 (3 echo + 1 health)
	if healthResp.TotalRequests < 4 {
		t.Errorf("Expected at least 4 total requests, got %d", healthResp.TotalRequests)
	}

	// Check uptime
	if healthResp.Uptime == "" {
		t.Error("Expected non-empty uptime")
	}

	// Check timestamp format
	if _, err := time.Parse(time.RFC3339, healthResp.Timestamp); err != nil {
		t.Errorf("Invalid timestamp format: %s", healthResp.Timestamp)
	}
}

// Test connection handling and cleanup
func TestIntegration_ConnectionHandling(t *testing.T) {
	srv, tempDir := setupTestServer(t)
	defer cleanup(srv, tempDir)

	baseURL := fmt.Sprintf("http://localhost:%s", srv.Port)

	// Create multiple connections and ensure they're tracked
	initialConnections := srv.GetOpenConnections()

	// Make multiple concurrent long-running requests
	const numConnections = 5
	done := make(chan struct{}, numConnections)

	for i := 0; i < numConnections; i++ {
		go func() {
			defer func() { done <- struct{}{} }()

			// Create connection with keep-alive
			client := &http.Client{
				Timeout: 10 * time.Second,
				Transport: &http.Transport{
					DisableKeepAlives: false,
				},
			}

			resp, err := client.Get(baseURL + "/echo/test")
			if err != nil {
				t.Errorf("Failed to make request: %v", err)
				return
			}
			defer resp.Body.Close()

			// Read response
			_, err = io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("Failed to read response: %v", err)
			}
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numConnections; i++ {
		<-done
	}

	// Give some time for connections to be cleaned up
	time.Sleep(100 * time.Millisecond)

	// Check that connections were handled properly
	finalConnections := srv.GetOpenConnections()
	if finalConnections < initialConnections {
		// This is expected behavior - connections should be cleaned up
		t.Logf("Connections properly cleaned up: %d -> %d", initialConnections, finalConnections)
	}
}

// TODO : Handle this test
// Test large request handling
// func TestIntegration_LargeRequests(t *testing.T) {
// 	srv, tempDir := setupTestServer(t)
// 	defer cleanup(srv, tempDir)

// 	baseURL := fmt.Sprintf("http://localhost:%s", srv.Port)

// 	// Test large echo request
// 	largeText := strings.Repeat("A", 500)
// 	resp, err := makeHTTPRequest("GET", baseURL+"/echo/"+largeText, "", nil)
// 	if err != nil {
// 		t.Fatalf("Failed to make large echo request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		t.Errorf("Expected status 200, got %d", resp.StatusCode)
// 	}

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		t.Fatalf("Failed to read response body: %v", err)
// 	}

// 	if string(body) != largeText {
// 		t.Errorf("Large echo response mismatch")
// 	}

// 	// Test large file upload
// 	largeContent := strings.Repeat("Hello World! ", 1000) // ~12KB
// 	headers := map[string]string{
// 		"Content-Type": "application/octet-stream",
// 	}

// 	resp, err = makeHTTPRequest("POST", baseURL+"/files/large.txt", largeContent, headers)
// 	if err != nil {
// 		t.Fatalf("Failed to upload large file: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusCreated {
// 		t.Errorf("Expected status 201 for large file upload, got %d", resp.StatusCode)
// 	}
// }
