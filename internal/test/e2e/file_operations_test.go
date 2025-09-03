package e2e

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/anh-nguyen/resource-server/internal/test/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type FileOperationsTestSuite struct {
	E2ETestSuite
}

// Test Cases for File Operations APIs

// FO-001: List files without filters
func (s *FileOperationsTestSuite) TestListFiles_NoFilters() {
	resp, err := s.GET("/api/v1/resources/r2/achievement")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	s.ParseSuccessResponse(resp, &result)

	s.Contains(result, "files")
	s.Contains(result, "continuationToken")

	files := result["files"].([]interface{})
	s.IsType([]interface{}{}, files)
}

// FO-002: List with max_keys limit
func (s *FileOperationsTestSuite) TestListFiles_MaxKeys() {
	resp, err := s.GET("/api/v1/resources/r2/achievement?max_keys=5")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	s.ParseSuccessResponse(resp, &result)

	files := result["files"].([]interface{})
	s.LessOrEqual(len(files), 5)
}

// FO-003: List with continuation token
func (s *FileOperationsTestSuite) TestListFiles_Pagination() {
	// First request
	resp1, err := s.GET("/api/v1/resources/r2/achievement?max_keys=2")
	s.Require().NoError(err)
	defer resp1.Body.Close()

	var result1 map[string]interface{}
	s.ParseSuccessResponse(resp1, &result1)

	if token, ok := result1["continuationToken"]; ok && token != "" {
		// Second request with continuation token
		resp2, err := s.GET(fmt.Sprintf("/api/v1/resources/r2/achievement?max_keys=2&continuation_token=%s", token))
		s.Require().NoError(err)
		defer resp2.Body.Close()

		s.Equal(http.StatusOK, resp2.StatusCode)

		var result2 map[string]interface{}
		s.ParseSuccessResponse(resp2, &result2)

		// Should get different files
		files2 := result2["files"].([]interface{})
		s.NotNil(files2)
	}
}

// FO-004: List with prefix filter
func (s *FileOperationsTestSuite) TestListFiles_WithPrefix() {
	resp, err := s.GET("/api/v1/resources/r2/achievement?prefix=icons/")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	s.ParseSuccessResponse(resp, &result)

	files := result["files"].([]interface{})

	// All files should have the prefix
	for _, f := range files {
		file := f.(map[string]interface{})
		key := file["key"].(string)
		s.Contains(key, "icons/")
	}
}

// FO-005: Invalid provider
func (s *FileOperationsTestSuite) TestListFiles_InvalidProvider() {
	resp, err := s.GET("/api/v1/resources/invalid/achievement")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// FO-006: Invalid definition
func (s *FileOperationsTestSuite) TestListFiles_InvalidDefinition() {
	resp, err := s.GET("/api/v1/resources/r2/invalid")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// FO-007: Empty bucket
func (s *FileOperationsTestSuite) TestListFiles_EmptyResult() {
	// Use a very specific prefix unlikely to match
	resp, err := s.GET("/api/v1/resources/r2/achievement?prefix=nonexistent/path/")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	s.ParseSuccessResponse(resp, &result)

	files := result["files"].([]interface{})
	s.Empty(files)
}

// FO-008: Generate upload URL with minimal params
func (s *FileOperationsTestSuite) TestGenerateUploadURL_Minimal() {
	achievementID := uuid.New().String()

	body := map[string]interface{}{
		"parameters": map[string]string{
			"achievementId": achievementID,
		},
		"scope":      "G",
		"scopeValue": 0,
	}

	resp, err := s.POST("/api/v1/resources/r2/achievement/upload", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	s.ParseSuccessResponse(resp, &result)

	s.Contains(result, "url")
	s.Contains(result, "path")
	s.Contains(result, "method")

	// Verify URL is valid
	uploadURL := result["url"].(string)
	_, err = url.Parse(uploadURL)
	s.NoError(err)

	// Verify method
	s.Equal("PUT", result["method"])

	// Verify path contains achievementId
	path := result["path"].(string)
	s.Contains(path, achievementID)
}

// FO-009-011: Upload with different scopes
func (s *FileOperationsTestSuite) TestGenerateUploadURL_Scopes() {
	testCases := []struct {
		name       string
		scope      string
		scopeValue interface{}
	}{
		{"Global", "G", 0},
		{"App", "A", 12345},
		{"ClientApp", "CA", 67890},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			achievementID := uuid.New().String()

			body := map[string]interface{}{
				"parameters": map[string]string{
					"achievementId": achievementID,
				},
				"scope":      tc.scope,
				"scopeValue": tc.scopeValue,
			}

			resp, err := s.POST("/api/v1/resources/r2/achievement/upload", body)
			s.Require().NoError(err)
			defer resp.Body.Close()

			s.Equal(http.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			s.ParseSuccessResponse(resp, &result)

			// Path should reflect scope
			path := result["path"].(string)
			if tc.scope != "G" {
				s.Contains(path, fmt.Sprintf("%v", tc.scopeValue))
			}
		})
	}
}

// FO-012: Upload with metadata
func (s *FileOperationsTestSuite) TestGenerateUploadURL_WithMetadata() {
	achievementID := uuid.New().String()

	body := map[string]interface{}{
		"parameters": map[string]string{
			"achievementId": achievementID,
		},
		"scope":      "G",
		"scopeValue": 0,
		"metadata": map[string]string{
			"contentType":  "image/png",
			"cacheControl": "max-age=3600",
		},
	}

	resp, err := s.POST("/api/v1/resources/r2/achievement/upload", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	s.ParseSuccessResponse(resp, &result)

	// Check if headers are included
	if headers, ok := result["headers"]; ok {
		headersMap := headers.(map[string]interface{})
		s.Contains(headersMap, "Content-Type")
		s.Contains(headersMap, "Cache-Control")
	}
}

// FO-013: Upload with custom expiry
func (s *FileOperationsTestSuite) TestGenerateUploadURL_CustomExpiry() {
	achievementID := uuid.New().String()

	body := map[string]interface{}{
		"parameters": map[string]string{
			"achievementId": achievementID,
		},
		"scope":      "G",
		"scopeValue": 0,
		"expiry":     "2h",
	}

	resp, err := s.POST("/api/v1/resources/r2/achievement/upload", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	s.ParseSuccessResponse(resp, &result)

	// URL should be generated with custom expiry
	s.NotEmpty(result["url"])
}

// FO-014: Missing required parameters
func (s *FileOperationsTestSuite) TestGenerateUploadURL_MissingParams() {
	body := map[string]interface{}{
		"parameters": map[string]string{
			// Missing achievementId
		},
		"scope":      "G",
		"scopeValue": 0,
	}

	resp, err := s.POST("/api/v1/resources/r2/achievement/upload", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// FO-015: Invalid scope value
func (s *FileOperationsTestSuite) TestGenerateUploadURL_InvalidScope() {
	achievementID := uuid.New().String()

	body := map[string]interface{}{
		"parameters": map[string]string{
			"achievementId": achievementID,
		},
		"scope":      "INVALID",
		"scopeValue": 0,
	}

	resp, err := s.POST("/api/v1/resources/r2/achievement/upload", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// FO-016-020: Download URL tests
func (s *FileOperationsTestSuite) TestGenerateDownloadURL_Success() {
	body := map[string]interface{}{
		"path":   "achievements/icons/test.png",
		"expiry": "1h",
	}

	resp, err := s.POST("/api/v1/resources/r2/*/download", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	// May return 404 if file doesn't exist, or 200 with URL
	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		s.ParseSuccessResponse(resp, &result)

		s.Contains(result, "url")
		s.Contains(result, "method")
		s.Equal("GET", result["method"])

		// Verify URL is valid
		downloadURL := result["url"].(string)
		_, err = url.Parse(downloadURL)
		s.NoError(err)
	} else {
		s.Equal(http.StatusNotFound, resp.StatusCode)
	}
}

// FO-021-024: Metadata operations
func (s *FileOperationsTestSuite) TestGetMetadata() {
	path := "achievements/icons/test.png"

	resp, err := s.GET(fmt.Sprintf("/api/v1/resources/r2/*/metadata?path=%s", url.QueryEscape(path)))
	s.Require().NoError(err)
	defer resp.Body.Close()

	// May return 404 if file doesn't exist
	if resp.StatusCode == http.StatusOK {
		var metadata map[string]interface{}
		s.ParseSuccessResponse(resp, &metadata)

		// Should have standard metadata fields
		s.Contains(metadata, "contentType")
		s.Contains(metadata, "contentLength")
		s.Contains(metadata, "lastModified")
	} else {
		s.Equal(http.StatusNotFound, resp.StatusCode)
	}
}

// FO-025-029: Update metadata
func (s *FileOperationsTestSuite) TestUpdateMetadata() {
	body := map[string]interface{}{
		"path": "achievements/icons/test.png",
		"metadata": map[string]string{
			"contentType":  "image/webp",
			"cacheControl": "max-age=7200",
		},
	}

	resp, err := s.PUT("/api/v1/resources/r2/*/metadata", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	// May return 404 if file doesn't exist
	if resp.StatusCode == http.StatusOK {
		helpers.AssertSuccessResponse(s.T(), resp)
	} else {
		s.Equal(http.StatusNotFound, resp.StatusCode)
	}
}

// FO-030-033: Delete file
func (s *FileOperationsTestSuite) TestDeleteFile() {
	path := "achievements/icons/test-delete.png"

	resp, err := s.DELETE(fmt.Sprintf("/api/v1/resources/r2/*?path=%s", url.QueryEscape(path)))
	s.Require().NoError(err)
	defer resp.Body.Close()

	// May return 404 if file doesn't exist, or 200 if deleted
	s.Contains([]int{http.StatusOK, http.StatusNotFound}, resp.StatusCode)
}

func TestFileOperationsSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping file operations tests in short mode")
	}

	suite.Run(t, new(FileOperationsTestSuite))
}
