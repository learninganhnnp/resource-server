package e2e

import (
	"net/http"
	"testing"

	"github.com/anh-nguyen/resource-server/internal/test/helpers"
	"github.com/stretchr/testify/suite"
)

type HealthTestSuite struct {
	E2ETestSuite
}

// Test Cases for Health Check API

// HC-001: Successful health check
func (s *HealthTestSuite) TestHealthCheck_Success() {
	resp, err := s.GET("/api/v1/health")
	s.Require().NoError(err)
	defer resp.Body.Close()

	// Verify status code
	s.Equal(http.StatusOK, resp.StatusCode)

	// Verify success response
	helpers.AssertSuccessResponse(s.T(), resp)
}

// HC-002: Verify response format
func (s *HealthTestSuite) TestHealthCheck_ResponseFormat() {
	resp, err := s.GET("/api/v1/health")
	s.Require().NoError(err)
	defer resp.Body.Close()

	var data map[string]interface{}
	s.ParseSuccessResponse(resp, &data)

	// Verify response contains expected fields
	s.Contains(data, "status")
	s.Contains(data, "timestamp")
	s.Equal("healthy", data["status"])
}

// HC-003: Check content-type header
func (s *HealthTestSuite) TestHealthCheck_ContentType() {
	resp, err := s.GET("/api/v1/health")
	s.Require().NoError(err)
	defer resp.Body.Close()

	// Verify content type
	helpers.AssertContentType(s.T(), resp, "application/json")
}

// HC-004: Verify server readiness
func (s *HealthTestSuite) TestHealthCheck_ServerReadiness() {
	resp, err := s.GET("/api/v1/health")
	s.Require().NoError(err)
	defer resp.Body.Close()

	// Server should respond within reasonable time
	s.Equal(http.StatusOK, resp.StatusCode)

	var data map[string]interface{}
	s.ParseSuccessResponse(resp, &data)

	// Verify server reports healthy status
	s.Equal("healthy", data["status"])

	// Timestamp should be present and valid
	s.NotEmpty(data["timestamp"])
}

// Additional test: Multiple concurrent health checks
func (s *HealthTestSuite) TestHealthCheck_Concurrent() {
	const numRequests = 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := s.GET("/api/v1/health")
			if err != nil {
				results <- 0
				return
			}
			defer resp.Body.Close()
			results <- resp.StatusCode
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < numRequests; i++ {
		statusCode := <-results
		if statusCode == http.StatusOK {
			successCount++
		}
	}

	// All requests should succeed
	s.Equal(numRequests, successCount)
}

func TestHealthSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping health endpoint tests in short mode")
	}

	suite.Run(t, new(HealthTestSuite))
}
