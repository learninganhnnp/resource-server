package e2e

import (
	"net/http"
	"testing"

	"github.com/anh-nguyen/resource-server/internal/test/helpers"
	"github.com/stretchr/testify/suite"
)

type ProviderTestSuite struct {
	E2ETestSuite
}

// Test Cases for Provider APIs

// PR-001: List all providers
func (s *ProviderTestSuite) TestListProviders_Success() {
	resp, err := s.GET("/api/v1/resources/providers")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)
	helpers.AssertSuccessResponse(s.T(), resp)

	var providers []map[string]interface{}
	s.ParseSuccessResponse(resp, &providers)

	// Should have cdn, gcs, r2 providers
	s.GreaterOrEqual(len(providers), 3)

	// Check for expected providers
	providerNames := make(map[string]bool)
	for _, p := range providers {
		providerNames[p["name"].(string)] = true
	}

	s.True(providerNames["cdn"], "CDN provider should exist")
	s.True(providerNames["gcs"], "GCS provider should exist")
	s.True(providerNames["r2"], "R2 provider should exist")
}

// PR-002: Verify capabilities structure
func (s *ProviderTestSuite) TestListProviders_CapabilitiesStructure() {
	resp, err := s.GET("/api/v1/resources/providers")
	s.Require().NoError(err)
	defer resp.Body.Close()

	var providers []map[string]interface{}
	s.ParseSuccessResponse(resp, &providers)

	for _, provider := range providers {
		s.Contains(provider, "name")
		s.Contains(provider, "capabilities")

		caps := provider["capabilities"].(map[string]interface{})

		// Verify capability fields
		s.Contains(caps, "supportsRead")
		s.Contains(caps, "supportsWrite")
		s.Contains(caps, "supportsDelete")
		s.Contains(caps, "supportsListing")
		s.Contains(caps, "supportsMetadata")
		s.Contains(caps, "supportsMultipart")
		s.Contains(caps, "supportsSignedUrls")

		// All should be boolean
		s.IsType(true, caps["supportsRead"])
		s.IsType(true, caps["supportsWrite"])
		s.IsType(true, caps["supportsDelete"])
	}
}

// PR-003: Check multipart support
func (s *ProviderTestSuite) TestListProviders_MultipartCapabilities() {
	resp, err := s.GET("/api/v1/resources/providers")
	s.Require().NoError(err)
	defer resp.Body.Close()

	var providers []map[string]interface{}
	s.ParseSuccessResponse(resp, &providers)

	// Find R2 provider (known to support multipart)
	var r2Provider map[string]interface{}
	for _, p := range providers {
		if p["name"] == "r2" {
			r2Provider = p
			break
		}
	}

	s.NotNil(r2Provider, "R2 provider should exist")

	caps := r2Provider["capabilities"].(map[string]interface{})
	s.True(caps["supportsMultipart"].(bool))

	// Check multipart details if present
	if multipart, ok := caps["multipart"]; ok && multipart != nil {
		mp := multipart.(map[string]interface{})
		s.Contains(mp, "minPartSize")
		s.Contains(mp, "maxPartSize")
		s.Contains(mp, "maxParts")

		// Verify reasonable values
		minSize := int(mp["minPartSize"].(float64))
		maxSize := int(mp["maxPartSize"].(float64))
		maxParts := int(mp["maxParts"].(float64))

		s.Greater(minSize, 0)
		s.Greater(maxSize, minSize)
		s.Greater(maxParts, 0)

		// Common S3 limits
		s.GreaterOrEqual(minSize, 5*1024*1024) // 5MB minimum for S3
		s.LessOrEqual(maxParts, 10000)         // S3 max parts limit
	}
}

// PR-004: Validate checksum algorithms
func (s *ProviderTestSuite) TestListProviders_ChecksumAlgorithms() {
	resp, err := s.GET("/api/v1/resources/providers")
	s.Require().NoError(err)
	defer resp.Body.Close()

	var providers []map[string]interface{}
	s.ParseSuccessResponse(resp, &providers)

	for _, provider := range providers {
		caps := provider["capabilities"].(map[string]interface{})

		if checksumAlgos, ok := caps["supportedChecksumAlgorithms"]; ok && checksumAlgos != nil {
			algos := checksumAlgos.([]interface{})

			// Common algorithms
			validAlgos := map[string]bool{
				"MD5":    true,
				"SHA1":   true,
				"SHA256": true,
				"CRC32":  true,
				"CRC32C": true,
			}

			for _, algo := range algos {
				s.True(validAlgos[algo.(string)], "Invalid algorithm: %s", algo)
			}
		}
	}
}

// PR-005: Get existing provider by name
func (s *ProviderTestSuite) TestGetProvider_Success() {
	resp, err := s.GET("/api/v1/resources/providers/r2")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var provider map[string]interface{}
	s.ParseSuccessResponse(resp, &provider)

	s.Equal("r2", provider["name"])
	s.Contains(provider, "capabilities")

	// Verify R2 specific capabilities
	caps := provider["capabilities"].(map[string]interface{})
	s.True(caps["supportsMultipart"].(bool))
	s.True(caps["supportsSignedUrls"].(bool))
}

// PR-006: Get non-existent provider
func (s *ProviderTestSuite) TestGetProvider_NotFound() {
	resp, err := s.GET("/api/v1/resources/providers/nonexistent")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusNotFound, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// PR-007: Invalid provider name
func (s *ProviderTestSuite) TestGetProvider_InvalidName() {
	resp, err := s.GET("/api/v1/resources/providers/invalid@provider!")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// Additional test: CDN provider capabilities
func (s *ProviderTestSuite) TestCDNProvider_Capabilities() {
	resp, err := s.GET("/api/v1/resources/providers/cdn")
	s.Require().NoError(err)
	defer resp.Body.Close()

	var provider map[string]interface{}
	s.ParseSuccessResponse(resp, &provider)

	caps := provider["capabilities"].(map[string]interface{})

	// CDN should support signed URLs for security
	s.True(caps["supportsSignedUrls"].(bool))

	// CDN should support read operations
	s.True(caps["supportsRead"].(bool))
}

// Additional test: GCS provider capabilities
func (s *ProviderTestSuite) TestGCSProvider_Capabilities() {
	resp, err := s.GET("/api/v1/resources/providers/gcs")
	s.Require().NoError(err)
	defer resp.Body.Close()

	var provider map[string]interface{}
	s.ParseSuccessResponse(resp, &provider)

	caps := provider["capabilities"].(map[string]interface{})

	// GCS should support all basic operations
	s.True(caps["supportsRead"].(bool))
	s.True(caps["supportsWrite"].(bool))
	s.True(caps["supportsDelete"].(bool))
	s.True(caps["supportsListing"].(bool))
	s.True(caps["supportsMetadata"].(bool))
}

// Additional test: Content type verification
func (s *ProviderTestSuite) TestListProviders_ContentType() {
	resp, err := s.GET("/api/v1/resources/providers")
	s.Require().NoError(err)
	defer resp.Body.Close()

	helpers.AssertContentType(s.T(), resp, "application/json")
}

func TestProviderSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping provider tests in short mode")
	}

	suite.Run(t, new(ProviderTestSuite))
}
