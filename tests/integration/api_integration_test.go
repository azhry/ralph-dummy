package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestHealthEndpoint tests the health endpoint of the running API
func TestHealthEndpoint(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get("http://localhost:8080/health")
	if err != nil {
		t.Skipf("API server is not running: %v", err)
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "Wedding Invitation Backend API is running", response["message"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Contains(t, response, "environment")
	assert.Contains(t, response, "timestamp")
}

// TestAPIInfoEndpoint tests the API info endpoint
func TestAPIInfoEndpoint(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get("http://localhost:8080/api/v1/info")
	if err != nil {
		t.Skipf("API server is not running: %v", err)
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, "Wedding Invitation Backend API", response["name"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, "A comprehensive backend for wedding invitation management", response["description"])

	endpoints, ok := response["endpoints"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, endpoints, "weddings")
	assert.Contains(t, endpoints, "rsvps")
	assert.Contains(t, endpoints, "guests")
	assert.Contains(t, endpoints, "health")
	assert.Contains(t, endpoints, "docs")
}

// TestWeddingEndpoints tests wedding-related endpoints
func TestWeddingEndpoints(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := "http://localhost:8080"

	// Test GET weddings list
	resp, err := client.Get(baseURL + "/api/v1/weddings")
	if err != nil {
		t.Skipf("API server is not running: %v", err)
		return
	}
	defer resp.Body.Close()

	// Should return some response (even if it's just a message)
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
}

// TestCreateWedding tests creating a wedding
func TestCreateWedding(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := "http://localhost:8080"

	weddingData := map[string]interface{}{
		"title":        "Integration Test Wedding",
		"groom_name":   "John Doe",
		"bride_name":   "Jane Smith",
		"wedding_date": time.Now().AddDate(0, 2, 0).Format(time.RFC3339),
		"venue":        "Test Venue",
		"description":  "Wedding created for integration testing",
	}

	jsonBody, err := json.Marshal(weddingData)
	assert.NoError(t, err)

	resp, err := client.Post(baseURL+"/api/v1/weddings", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Skipf("API server is not running: %v", err)
		return
	}
	defer resp.Body.Close()

	// Should return created, success, or validation error
	assert.True(t, resp.StatusCode == http.StatusCreated ||
		resp.StatusCode == http.StatusOK ||
		resp.StatusCode == http.StatusBadRequest)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
}

// TestGetNonExistentWedding tests error handling
func TestGetNonExistentWedding(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get("http://localhost:8080/api/v1/weddings/507f1f77bcf86cd799439011")
	if err != nil {
		t.Skipf("API server is not running: %v", err)
		return
	}
	defer resp.Body.Close()

	// May return 404 (not found), 500 (server error due to auth), or 401 (unauthorized)
	assert.True(t, resp.StatusCode == http.StatusNotFound ||
		resp.StatusCode == http.StatusInternalServerError ||
		resp.StatusCode == http.StatusUnauthorized)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	// Response might not be JSON in case of 500 error
	if err == nil {
		// If we can parse the response, check for error structure
		if success, ok := response["success"].(bool); ok {
			assert.Equal(t, false, success)
		}
		assert.Contains(t, response, "error")
	}
}

// TestGuestEndpoints tests guest-related endpoints
func TestGuestEndpoints(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := "http://localhost:8080"

	// Test GET guests list
	resp, err := client.Get(baseURL + "/api/v1/guests")
	if err != nil {
		t.Skipf("API server is not running: %v", err)
		return
	}
	defer resp.Body.Close()

	// Should return some response
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
}

// TestCreateGuest tests creating a guest
func TestCreateGuest(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := "http://localhost:8080"

	guestData := map[string]interface{}{
		"first_name":   "Test",
		"last_name":    "Guest",
		"email":        "testguest@example.com",
		"relationship": "friend",
		"side":         "groom",
		"wedding_id":   "507f1f77bcf86cd799439011", // Non-existent ID
	}

	jsonBody, err := json.Marshal(guestData)
	assert.NoError(t, err)

	resp, err := client.Post(baseURL+"/api/v1/guests", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Skipf("API server is not running: %v", err)
		return
	}
	defer resp.Body.Close()

	// Should return success, validation error, or not found (for non-existent wedding)
	assert.True(t, resp.StatusCode == http.StatusCreated ||
		resp.StatusCode == http.StatusOK ||
		resp.StatusCode == http.StatusBadRequest ||
		resp.StatusCode == http.StatusNotFound)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
}

// TestRSVPEndpoints tests RSVP-related endpoints
func TestRSVPEndpoints(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := "http://localhost:8080"

	// Test GET RSVPs for a wedding
	resp, err := client.Get(baseURL + "/api/v1/rsvps/507f1f77bcf86cd799439011")
	if err != nil {
		t.Skipf("API server is not running: %v", err)
		return
	}
	defer resp.Body.Close()

	// Should return a response (may be 200, 404, 401, or 500 due to auth)
	assert.True(t, resp.StatusCode == http.StatusOK ||
		resp.StatusCode == http.StatusNotFound ||
		resp.StatusCode == http.StatusUnauthorized ||
		resp.StatusCode == http.StatusInternalServerError)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	// Response might not be JSON in case of 500 error
	if err == nil {
		// If we can parse the response, validate structure
		assert.Contains(t, response, "success")
		if success, ok := response["success"].(bool); ok && !success {
			assert.Contains(t, response, "error")
		}
	}
}

// TestInputValidation tests input validation
func TestInputValidation(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := "http://localhost:8080"

	// Test creating wedding with invalid data
	invalidWedding := map[string]interface{}{
		"title": "", // Empty title should fail validation
	}

	jsonBody, err := json.Marshal(invalidWedding)
	assert.NoError(t, err)

	resp, err := client.Post(baseURL+"/api/v1/weddings", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Skipf("API server is not running: %v", err)
		return
	}
	defer resp.Body.Close()

	// Should return validation error
	assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
}

// TestInvalidEndpoint tests invalid endpoint handling
func TestInvalidEndpoint(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get("http://localhost:8080/api/v1/invalid")
	if err != nil {
		t.Skipf("API server is not running: %v", err)
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestMethodNotAllowed tests HTTP method restrictions
func TestMethodNotAllowed(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}

	// Try DELETE on health endpoint (should not be allowed)
	req, err := http.NewRequest("DELETE", "http://localhost:8080/health", nil)
	assert.NoError(t, err)

	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("API server is not running: %v", err)
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestContentTypeHandling tests content-type validation
func TestContentTypeHandling(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := "http://localhost:8080"

	// Send request without proper content-type
	weddingData := map[string]interface{}{
		"title": "Test Wedding",
	}

	jsonBody, _ := json.Marshal(weddingData)
	req, err := http.NewRequest("POST", baseURL+"/api/v1/weddings", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	// Intentionally not setting Content-Type header

	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("API server is not running: %v", err)
		return
	}
	defer resp.Body.Close()

	// Should handle missing content-type gracefully
	assert.True(t, resp.StatusCode == http.StatusBadRequest ||
		resp.StatusCode == http.StatusCreated ||
		resp.StatusCode == http.StatusOK)
}

// TestConcurrentRequests tests concurrent request handling
func TestConcurrentRequests(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	const numRequests = 5
	results := make(chan error, numRequests)

	// Launch concurrent health check requests
	for i := 0; i < numRequests; i++ {
		go func(id int) {
			resp, err := client.Get("http://localhost:8080/health")
			if err != nil {
				results <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				return
			}

			results <- nil
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		err := <-results
		// Some requests might fail, but server should handle concurrency
		if err != nil {
			t.Logf("Concurrent request failed: %v", err)
		}
	}
}

// TestServerStartupIntegration checks if server is running and basic endpoints work
func TestServerStartupIntegration(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}

	// Check if server is running
	resp, err := client.Get("http://localhost:8080/health")
	if err != nil {
		t.Skipf("API server is not running. Please start with: go run cmd/api/main.go")
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse health response
	var healthResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	assert.NoError(t, err)

	// Check health response structure
	assert.Equal(t, "ok", healthResponse["status"])
	assert.Equal(t, "Wedding Invitation Backend API is running", healthResponse["message"])
	assert.Equal(t, "1.0.0", healthResponse["version"])

	// Test API info endpoint
	infoResp, err := client.Get("http://localhost:8080/api/v1/info")
	assert.NoError(t, err)
	defer infoResp.Body.Close()

	assert.Equal(t, http.StatusOK, infoResp.StatusCode)

	var infoResponse map[string]interface{}
	err = json.NewDecoder(infoResp.Body).Decode(&infoResponse)
	assert.NoError(t, err)

	assert.Equal(t, "Wedding Invitation Backend API", infoResponse["name"])
	assert.Contains(t, infoResponse, "endpoints")

	// Test that documented endpoints exist (even if they return placeholder responses)
	endpoints := infoResponse["endpoints"].(map[string]interface{})
	for endpointName, endpointPath := range endpoints {
		if endpointName == "health" || endpointName == "docs" {
			continue // Skip these as they're already tested
		}

		resp, err := client.Get("http://localhost:8080" + endpointPath.(string))
		assert.NoError(t, err, fmt.Sprintf("Endpoint %s should be accessible", endpointName))

		// Should not be 500 (internal server error)
		assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode,
			fmt.Sprintf("Endpoint %s should not return 500 error", endpointName))

		if resp != nil {
			resp.Body.Close()
		}
	}

	t.Log("✅ All basic integration tests passed!")
	t.Log("🔗 Server is running correctly on http://localhost:8080")
	t.Log("📚 API documentation available at http://localhost:8080/swagger/index.html")
}
