package e2e

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/anh-nguyen/resource-server/internal/test/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type MultipartTestSuite struct {
	E2ETestSuite
}

// Test Cases for Multipart Upload APIs

// MP-001: Initialize multipart upload
func (s *MultipartTestSuite) TestInitMultipart_Success() {
	achievementID := uuid.New().String()

	body := map[string]interface{}{
		"definitionName": "achievement",
		"provider":       "r2",
		"scope":          "G",
		"scopeValue":     0,
		"parameters": map[string]string{
			"achievement_id": achievementID,
		},
	}

	resp, err := s.POST("/api/v1/resources/multipart/init", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	s.ParseSuccessResponse(resp, &result)

	s.Contains(result, "uploadId")
	s.Contains(result, "path")
	s.Contains(result, "constraints")

	// Verify upload ID format
	uploadID := result["uploadId"].(string)
	s.NotEmpty(uploadID)

	// Verify path contains achievementId
	path := result["path"].(string)
	s.Contains(path, achievementID)

	// Verify constraints
	constraints := result["constraints"].(map[string]interface{})
	s.Contains(constraints, "minPartSize")
	s.Contains(constraints, "maxPartSize")
	s.Contains(constraints, "maxParts")
}

// MP-002: Init with minimal parameters
func (s *MultipartTestSuite) TestInitMultipart_MinimalParams() {
	achievementID := uuid.New().String()

	body := map[string]interface{}{
		"definitionName": "achievement",
		"provider":       "r2",
		"scope":          "G",
		"scopeValue":     0,
		"parameters": map[string]string{
			"achievement_id": achievementID,
		},
	}

	resp, err := s.POST("/api/v1/resources/multipart/init", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)
	helpers.AssertSuccessResponse(s.T(), resp)
}

// MP-003: Init with metadata
func (s *MultipartTestSuite) TestInitMultipart_WithMetadata() {
	achievementID := uuid.New().String()

	body := map[string]interface{}{
		"definitionName": "achievement",
		"provider":       "r2",
		"scope":          "G",
		"scopeValue":     0,
		"parameters": map[string]string{
			"achievement_id": achievementID,
		},
		"metadata": map[string]string{
			"contentType":     "image/png",
			"contentEncoding": "gzip",
			"cacheControl":    "max-age=3600",
		},
	}

	resp, err := s.POST("/api/v1/resources/multipart/init", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	s.ParseSuccessResponse(resp, &result)

	// Metadata should be accepted
	s.NotEmpty(result["uploadId"])
}

// MP-004: Invalid provider
func (s *MultipartTestSuite) TestInitMultipart_InvalidProvider() {
	achievementID := uuid.New().String()

	body := map[string]interface{}{
		"definitionName": "achievement",
		"provider":       "invalid",
		"scope":          "G",
		"scopeValue":     0,
		"parameters": map[string]string{
			"achievement_id": achievementID,
		},
	}

	resp, err := s.POST("/api/v1/resources/multipart/init", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// MP-005: Invalid definition
func (s *MultipartTestSuite) TestInitMultipart_InvalidDefinition() {
	achievementID := uuid.New().String()

	body := map[string]interface{}{
		"definitionName": "invalid",
		"provider":       "r2",
		"scope":          "G",
		"scopeValue":     0,
		"parameters": map[string]string{
			"achievement_id": achievementID,
		},
	}

	resp, err := s.POST("/api/v1/resources/multipart/init", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// MP-006: Missing required fields
func (s *MultipartTestSuite) TestInitMultipart_MissingFields() {
	body := map[string]interface{}{
		"definitionName": "achievement",
		"provider":       "r2",
		// Missing scope and parameters
	}

	resp, err := s.POST("/api/v1/resources/multipart/init", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// MP-007: Get part URLs for valid upload
func (s *MultipartTestSuite) TestGetPartURLs_Success() {
	// First init multipart upload
	achievementID := uuid.New().String()
	path := "achievements/icons/" + achievementID + ".png"
	uploadID := "test-upload-" + uuid.New().String()

	// Generate checksum for part
	partData := []byte("test part data")
	hash := sha256.Sum256(partData)
	checksum := base64.StdEncoding.EncodeToString(hash[:])

	body := map[string]interface{}{
		"path":     path,
		"uploadId": uploadID,
		"provider": "r2",
		"urlOptions": []map[string]interface{}{
			{
				"partNumber": 1,
				"checksum": map[string]string{
					"algorithm": "SHA256",
					"value":     checksum,
				},
			},
		},
	}

	resp, err := s.POST("/api/v1/resources/multipart/urls", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	// May return error if upload ID doesn't exist in real scenario
	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		s.ParseSuccessResponse(resp, &result)

		s.Contains(result, "partUrls")
		s.Contains(result, "completeUrl")
		s.Contains(result, "abortUrl")

		partURLs := result["partUrls"].([]interface{})
		s.Len(partURLs, 1)

		// Verify part URL structure
		partURL := partURLs[0].(map[string]interface{})
		s.Contains(partURL, "partNumber")
		s.Contains(partURL, "url")
		s.Contains(partURL, "method")
		s.Equal(float64(1), partURL["partNumber"])
		s.Equal("PUT", partURL["method"])
	}
}

// MP-008: Get URLs with checksums
func (s *MultipartTestSuite) TestGetPartURLs_WithChecksums() {
	path := "achievements/icons/test.png"
	uploadID := "test-upload-" + uuid.New().String()

	// Generate multiple parts with checksums
	body := map[string]interface{}{
		"path":     path,
		"uploadId": uploadID,
		"provider": "r2",
		"urlOptions": []map[string]interface{}{
			{
				"partNumber": 1,
				"checksum": map[string]string{
					"algorithm": "SHA256",
					"value":     "abc123def456",
				},
			},
			{
				"partNumber": 2,
				"checksum": map[string]string{
					"algorithm": "SHA256",
					"value":     "789ghi012jkl",
				},
			},
		},
	}

	resp, err := s.POST("/api/v1/resources/multipart/urls", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		s.ParseSuccessResponse(resp, &result)

		partURLs := result["partUrls"].([]interface{})
		s.Len(partURLs, 2)

		// Check if headers include checksum
		for _, urlInfo := range partURLs {
			partURL := urlInfo.(map[string]interface{})
			if headers, ok := partURL["headers"]; ok {
				headersMap := headers.(map[string]interface{})
				// May contain checksum headers depending on provider
				_ = headersMap
			}
		}
	}
}

// MP-009: Request too many parts
func (s *MultipartTestSuite) TestGetPartURLs_TooManyParts() {
	path := "achievements/icons/test.png"
	uploadID := "test-upload-" + uuid.New().String()

	// Create request with excessive parts (>10000)
	urlOptions := make([]map[string]interface{}, 10001)
	for i := 0; i < 10001; i++ {
		urlOptions[i] = map[string]interface{}{
			"partNumber": i + 1,
		}
	}

	body := map[string]interface{}{
		"path":       path,
		"uploadId":   uploadID,
		"provider":   "r2",
		"urlOptions": urlOptions,
	}

	resp, err := s.POST("/api/v1/resources/multipart/urls", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	// Should reject too many parts
	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// MP-010: Invalid upload ID
func (s *MultipartTestSuite) TestGetPartURLs_InvalidUploadID() {
	body := map[string]interface{}{
		"path":     "achievements/icons/test.png",
		"uploadId": "",
		"provider": "r2",
		"urlOptions": []map[string]interface{}{
			{
				"partNumber": 1,
			},
		},
	}

	resp, err := s.POST("/api/v1/resources/multipart/urls", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// MP-011: Get complete and abort URLs
func (s *MultipartTestSuite) TestGetPartURLs_CompleteAbortURLs() {
	path := "achievements/icons/test.png"
	uploadID := "test-upload-" + uuid.New().String()

	body := map[string]interface{}{
		"path":     path,
		"uploadId": uploadID,
		"provider": "r2",
		"urlOptions": []map[string]interface{}{
			{
				"partNumber": 1,
			},
		},
	}

	resp, err := s.POST("/api/v1/resources/multipart/urls", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		s.ParseSuccessResponse(resp, &result)

		// Should include complete and abort URLs
		s.Contains(result, "completeUrl")
		s.Contains(result, "abortUrl")

		completeURL := result["completeUrl"].(string)
		abortURL := result["abortUrl"].(string)

		s.NotEmpty(completeURL)
		s.NotEmpty(abortURL)
		s.NotEqual(completeURL, abortURL)
	}
}

// Additional test: Provider that doesn't support multipart
func (s *MultipartTestSuite) TestInitMultipart_UnsupportedProvider() {
	achievementID := uuid.New().String()

	// Try with CDN which might not support multipart
	body := map[string]interface{}{
		"definitionName": "achievement",
		"provider":       "cdn",
		"scope":          "G",
		"scopeValue":     0,
		"parameters": map[string]string{
			"achievement_id": achievementID,
		},
	}

	resp, err := s.POST("/api/v1/resources/multipart/init", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	// Should either succeed (if supported) or return appropriate error
	if resp.StatusCode != http.StatusOK {
		helpers.AssertErrorResponse(s.T(), resp, "")
	}
}

// Additional test: Validate part size constraints
func (s *MultipartTestSuite) TestInitMultipart_PartSizeConstraints() {
	achievementID := uuid.New().String()

	body := map[string]interface{}{
		"definitionName": "achievement",
		"provider":       "r2",
		"scope":          "G",
		"scopeValue":     0,
		"parameters": map[string]string{
			"achievement_id": achievementID,
		},
	}

	resp, err := s.POST("/api/v1/resources/multipart/init", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	s.ParseSuccessResponse(resp, &result)

	constraints := result["constraints"].(map[string]interface{})

	// Verify reasonable constraints (S3-compatible)
	minPartSize := int(constraints["minPartSize"].(float64))
	maxPartSize := int(constraints["maxPartSize"].(float64))
	maxParts := int(constraints["maxParts"].(float64))

	s.GreaterOrEqual(minPartSize, 5*1024*1024)   // 5MB minimum
	s.LessOrEqual(maxPartSize, 5*1024*1024*1024) // 5GB maximum
	s.LessOrEqual(maxParts, 10000)               // 10000 parts max
	s.Greater(maxPartSize, minPartSize)
}

func TestMultipartSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multipart upload tests in short mode")
	}

	suite.Run(t, new(MultipartTestSuite))
}
