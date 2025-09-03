package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type E2ETestSuite struct {
	suite.Suite
	client  *http.Client
	baseURL string
	ctx     context.Context
	cancel  context.CancelFunc
}

func (s *E2ETestSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	
	// Get server port from environment
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8081"
	}

	// Setup HTTP client
	s.client = &http.Client{
		Timeout: 10 * time.Second,
	}
	s.baseURL = fmt.Sprintf("http://localhost:%s", port)
}

func (s *E2ETestSuite) TearDownSuite() {
	s.cancel()
}

// Helper methods for API calls
func (s *E2ETestSuite) GET(path string) (*http.Response, error) {
	return s.client.Get(s.baseURL + path)
}

func (s *E2ETestSuite) POST(path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(data)
	}
	
	req, err := http.NewRequest("POST", s.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	
	return s.client.Do(req)
}

func (s *E2ETestSuite) PUT(path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(data)
	}
	
	req, err := http.NewRequest("PUT", s.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	
	return s.client.Do(req)
}

func (s *E2ETestSuite) DELETE(path string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", s.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	
	return s.client.Do(req)
}

// Response parsing helpers
func (s *E2ETestSuite) ParseSuccessResponse(resp *http.Response, data interface{}) {
	defer resp.Body.Close()
	
	var result struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
		Message string      `json:"message"`
		Error   interface{} `json:"error"`
	}
	
	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err, "Failed to read response body")
	
	err = json.Unmarshal(body, &result)
	s.Require().NoError(err, "Failed to parse response: %s", string(body))
	
	s.True(result.Success, "Expected success response, got error: %v", result.Error)
	
	if data != nil && result.Data != nil {
		// Re-marshal and unmarshal to handle interface{} -> struct conversion
		dataBytes, err := json.Marshal(result.Data)
		s.Require().NoError(err)
		err = json.Unmarshal(dataBytes, data)
		s.Require().NoError(err)
	}
}

func (s *E2ETestSuite) ParseErrorResponse(resp *http.Response) map[string]interface{} {
	defer resp.Body.Close()
	
	var result struct {
		Success bool                   `json:"success"`
		Data    interface{}            `json:"data"`
		Message string                 `json:"message"`
		Error   map[string]interface{} `json:"error"`
	}
	
	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err, "Failed to read response body")
	
	err = json.Unmarshal(body, &result)
	s.Require().NoError(err, "Failed to parse response: %s", string(body))
	
	s.False(result.Success, "Expected error response, got success")
	s.NotNil(result.Error, "Expected error details")
	
	return result.Error
}

func TestE2ESuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}
	
	suite.Run(t, new(E2ETestSuite))
}