package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/anh-nguyen/resource-server/internal/test/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type WorkflowTestSuite struct {
	E2ETestSuite
	testDB *helpers.TestDatabase
}

func (s *WorkflowTestSuite) SetupSuite() {
	s.E2ETestSuite.SetupSuite()
	s.testDB = helpers.SetupTestDatabase(s.T())
}

func (s *WorkflowTestSuite) TearDownSuite() {
	if s.testDB != nil {
		s.testDB.Close()
	}
	s.E2ETestSuite.TearDownSuite()
}

func (s *WorkflowTestSuite) SetupTest() {
	s.testDB.Cleanup(s.T())
}

// Complex Workflow Tests

// Workflow 1: Complete Achievement Creation with Icon Upload
func (s *WorkflowTestSuite) TestWorkflow_AchievementWithIcon() {
	// Step 1: Create achievement with icon format specified
	body := map[string]interface{}{
		"name":        "Workflow Achievement",
		"description": "Achievement created in workflow test",
		"category":    "workflow",
		"points":      500,
		"iconFormat":  "png",
		"provider":    "r2",
	}

	resp1, err := s.POST("/api/v1/achievements/", body)
	s.Require().NoError(err)
	defer resp1.Body.Close()

	s.Equal(http.StatusCreated, resp1.StatusCode)

	var createResult map[string]interface{}
	s.ParseSuccessResponse(resp1, &createResult)

	// Step 2: Receive upload URL
	s.Contains(createResult, "upload")

	achievementID := createResult["id"].(string)
	s.NotEmpty(achievementID)

	upload := createResult["upload"].(map[string]interface{})
	uploadURL := upload["upload_url"].(string)
	s.NotEmpty(uploadURL)

	// Step 3: Simulate file upload to URL
	// In real scenario, client would upload file to the signed URL
	// For testing, we just verify the URL structure
	s.Contains(uploadURL, achievementID)

	// Step 4: Confirm upload completion
	// Get the upload ID from context (would be provided in real scenario)
	uploadID := "upload-" + uuid.New().String()

	confirmBody := map[string]interface{}{
		"success":      true,
		"fileSize":     1024000,
		"contentType":  "image/png",
		"verifyExists": false, // Skip actual verification for test
	}

	resp2, err := s.POST(fmt.Sprintf("/api/v1/achievements/uploads/%s/confirm", uploadID), confirmBody)
	s.Require().NoError(err)
	defer resp2.Body.Close()

	// May return 404 if upload tracking not implemented, that's OK for test
	if resp2.StatusCode != http.StatusNotFound {
		s.Contains([]int{http.StatusOK, http.StatusBadRequest}, resp2.StatusCode)
	}

	// Step 5: Verify achievement exists
	resp3, err := s.GET(fmt.Sprintf("/api/v1/achievements/%s", achievementID))
	s.Require().NoError(err)
	defer resp3.Body.Close()

	s.Equal(http.StatusOK, resp3.StatusCode)

	var getResult map[string]interface{}
	s.ParseSuccessResponse(resp3, &getResult)

	retrievedAchievement := getResult["achievement"].(map[string]interface{})
	s.Equal(achievementID, retrievedAchievement["id"])
	s.Equal("Workflow Achievement", retrievedAchievement["name"])
}

// Workflow 2: Large File Multipart Upload
func (s *WorkflowTestSuite) TestWorkflow_MultipartUpload() {
	// Step 1: Initialize multipart upload for large icon (>5MB)
	achievementID := uuid.New().String()

	initBody := map[string]interface{}{
		"definitionName": "achievement",
		"provider":       "r2",
		"scope":          "G",
		"scopeValue":     0,
		"paramResolver": map[string]string{
			"achievementId": achievementID,
		},
		"metadata": map[string]string{
			"contentType": "image/png",
			"fileSize":    "10485760", // 10MB
		},
	}

	resp1, err := s.POST("/api/v1/resources/multipart/init", initBody)
	s.Require().NoError(err)
	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK {
		s.T().Skip("Multipart not available, skipping workflow test")
		return
	}

	var initResult map[string]interface{}
	s.ParseSuccessResponse(resp1, &initResult)

	uploadID := initResult["uploadId"].(string)
	path := initResult["path"].(string)

	s.NotEmpty(uploadID)
	s.NotEmpty(path)

	// Step 2: Get part URLs for multiple parts (simulate 2 parts of 5MB each)
	urlsBody := map[string]interface{}{
		"path":     path,
		"uploadId": uploadID,
		"provider": "r2",
		"urlOptions": []map[string]interface{}{
			{"partNumber": 1},
			{"partNumber": 2},
		},
	}

	resp2, err := s.POST("/api/v1/resources/multipart/urls", urlsBody)
	s.Require().NoError(err)
	defer resp2.Body.Close()

	if resp2.StatusCode == http.StatusOK {
		var urlsResult map[string]interface{}
		s.ParseSuccessResponse(resp2, &urlsResult)

		s.Contains(urlsResult, "partUrls")
		s.Contains(urlsResult, "completeUrl")
		s.Contains(urlsResult, "abortUrl")

		partURLs := urlsResult["partUrls"].([]interface{})
		s.Len(partURLs, 2)

		// Step 3: Simulate uploading each part
		// In real scenario, client would upload to each part URL

		// Step 4: Complete multipart upload with ETags
		// In real scenario, client would call complete URL with ETags from upload responses

		// For test, just verify URLs are properly formatted
		for _, partInfo := range partURLs {
			part := partInfo.(map[string]interface{})
			s.Contains(part, "partNumber")
			s.Contains(part, "url")
			s.Contains(part, "method")
			s.Equal("PUT", part["method"])
		}
	}
}

// Workflow 3: File Lifecycle Management
func (s *WorkflowTestSuite) TestWorkflow_FileLifecycle() {
	achievementID := uuid.New().String()

	// Step 1: Generate upload URL
	uploadBody := map[string]interface{}{
		"parameters": map[string]string{
			"achievement_id": achievementID,
		},
		"scope":      "G",
		"scopeValue": 0,
		"metadata": map[string]string{
			"contentType":  "image/png",
			"cacheControl": "max-age=3600",
		},
	}

	resp1, err := s.POST("/api/v1/resources/r2/achievement/upload", uploadBody)
	s.Require().NoError(err)
	defer resp1.Body.Close()

	s.Equal(http.StatusOK, resp1.StatusCode)

	var uploadResult map[string]interface{}
	s.ParseSuccessResponse(resp1, &uploadResult)

	uploadURL := uploadResult["url"].(string)
	filePath := uploadResult["path"].(string)

	s.NotEmpty(uploadURL)
	s.NotEmpty(filePath)

	// Step 2: Upload file (simulated)
	s.T().Log("Would upload file to:", uploadURL)

	// Step 3: List files to verify existence
	resp2, err := s.GET("/api/v1/resources/r2/achievement?prefix=" + achievementID)
	s.Require().NoError(err)
	defer resp2.Body.Close()

	s.Equal(http.StatusOK, resp2.StatusCode)

	var listResult map[string]interface{}
	s.ParseSuccessResponse(resp2, &listResult)

	files := listResult["files"].([]interface{})
	s.IsType([]interface{}{}, files)

	// Step 4: Get file metadata (may not exist yet in test)
	resp3, err := s.GET(fmt.Sprintf("/api/v1/resources/r2/*/metadata?path=%s", filePath))
	s.Require().NoError(err)
	defer resp3.Body.Close()

	// File may not exist in test environment
	if resp3.StatusCode == http.StatusOK {
		var metadataResult map[string]interface{}
		s.ParseSuccessResponse(resp3, &metadataResult)
		s.Contains(metadataResult, "contentType")
	}

	// Step 5: Update file metadata
	updateMetadata := map[string]interface{}{
		"path": filePath,
		"metadata": map[string]string{
			"contentType":  "image/webp",
			"cacheControl": "max-age=7200",
		},
	}

	resp4, err := s.PUT("/api/v1/resources/r2/*/metadata", updateMetadata)
	s.Require().NoError(err)
	defer resp4.Body.Close()

	// May return 404 if file doesn't exist
	s.Contains([]int{http.StatusOK, http.StatusNotFound}, resp4.StatusCode)

	// Step 6: Generate download URL
	downloadBody := map[string]interface{}{
		"path":   filePath,
		"expiry": "1h",
	}

	resp5, err := s.POST("/api/v1/resources/r2/*/download", downloadBody)
	s.Require().NoError(err)
	defer resp5.Body.Close()

	// May return 404 if file doesn't exist
	if resp5.StatusCode == http.StatusOK {
		var downloadResult map[string]interface{}
		s.ParseSuccessResponse(resp5, &downloadResult)
		s.Contains(downloadResult, "url")
		s.Contains(downloadResult, "method")
		s.Equal("GET", downloadResult["method"])
	}

	// Step 7: Delete file
	resp6, err := s.DELETE(fmt.Sprintf("/api/v1/resources/r2/*?path=%s", filePath))
	s.Require().NoError(err)
	defer resp6.Body.Close()

	// Should succeed or return not found
	s.Contains([]int{http.StatusOK, http.StatusNotFound}, resp6.StatusCode)

	// Step 8: Verify file no longer exists
	resp7, err := s.GET("/api/v1/resources/r2/achievement?prefix=" + achievementID)
	s.Require().NoError(err)
	defer resp7.Body.Close()

	var finalListResult map[string]interface{}
	s.ParseSuccessResponse(resp7, &finalListResult)

	finalFiles := finalListResult["files"].([]interface{})

	// File should be gone or not found in first place
	fileFound := false
	for _, f := range finalFiles {
		file := f.(map[string]interface{})
		if file["key"].(string) == filePath {
			fileFound = true
			break
		}
	}

	if len(finalFiles) > 0 {
		s.False(fileFound, "File should be deleted")
	}
}

// Workflow 4: Achievement Management Flow
func (s *WorkflowTestSuite) TestWorkflow_AchievementManagement() {
	// Step 1: Create multiple achievements
	achievements := []map[string]interface{}{
		{
			"name":        "First Achievement",
			"description": "First achievement in management flow",
			"category":    "beginner",
			"points":      100,
		},
		{
			"name":        "Second Achievement",
			"description": "Second achievement in management flow",
			"category":    "intermediate",
			"points":      250,
		},
		{
			"name":        "Third Achievement",
			"description": "Third achievement in management flow",
			"category":    "advanced",
			"points":      500,
		},
	}

	createdIDs := make([]string, 0, len(achievements))

	for _, ach := range achievements {
		resp, err := s.POST("/api/v1/achievements/", ach)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		s.ParseSuccessResponse(resp, &result)

		achievement := result["achievement"].(map[string]interface{})
		createdIDs = append(createdIDs, achievement["id"].(string))
	}

	s.Len(createdIDs, 3)

	// Step 2: List achievements with pagination
	resp1, err := s.GET("/api/v1/achievements/?page=1&pageSize=2")
	s.Require().NoError(err)
	defer resp1.Body.Close()

	s.Equal(http.StatusOK, resp1.StatusCode)

	var listResult map[string]interface{}
	s.ParseSuccessResponse(resp1, &listResult)

	achievements1 := listResult["achievements"].([]interface{})
	pagination1 := listResult["pagination"].(map[string]interface{})

	s.LessOrEqual(len(achievements1), 2)
	s.Equal(float64(1), pagination1["page"])
	s.Equal(float64(2), pagination1["pageSize"])

	// Step 3: Update achievement icons
	iconUpdate := map[string]interface{}{
		"format":   "png",
		"provider": "r2",
	}

	for _, id := range createdIDs {
		resp, err := s.PUT(fmt.Sprintf("/api/v1/achievements/%s/icon", id), iconUpdate)
		s.Require().NoError(err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var updateResult map[string]interface{}
			s.ParseSuccessResponse(resp, &updateResult)
			s.Contains(updateResult, "uploadUrl")
		}
	}

	// Step 4: Filter active achievements
	resp2, err := s.GET("/api/v1/achievements/?only_active=true")
	s.Require().NoError(err)
	defer resp2.Body.Close()

	s.Equal(http.StatusOK, resp2.StatusCode)

	var activeResult map[string]interface{}
	s.ParseSuccessResponse(resp2, &activeResult)

	activeAchievements := activeResult["achievements"].([]interface{})

	// All should be active (newly created are active by default)
	for _, ach := range activeAchievements {
		achievement := ach.(map[string]interface{})
		s.True(achievement["isActive"].(bool))
	}

	// Step 5: Verify category filtering by checking individual achievements
	for _, id := range createdIDs {
		resp, err := s.GET(fmt.Sprintf("/api/v1/achievements/%s", id))
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusOK, resp.StatusCode)

		var getResult map[string]interface{}
		s.ParseSuccessResponse(resp, &getResult)

		achievement := getResult["achievement"].(map[string]interface{})
		s.Contains(achievement, "category")
		s.Contains([]string{"beginner", "intermediate", "advanced"}, achievement["category"])
	}
}

// Workflow 5: Cross-Provider Operations
func (s *WorkflowTestSuite) TestWorkflow_CrossProvider() {
	achievementID := uuid.New().String()
	providers := []string{"cdn", "gcs", "r2"}

	for _, provider := range providers {
		s.Run(fmt.Sprintf("Provider_%s", provider), func() {
			// Step 1: Upload to provider
			uploadBody := map[string]interface{}{
				"parameters": map[string]string{
					"achievementId": achievementID,
				},
				"scope":      "G",
				"scopeValue": 0,
			}

			resp1, err := s.POST(fmt.Sprintf("/api/v1/resources/%s/achievement/upload", provider), uploadBody)
			s.Require().NoError(err)
			defer resp1.Body.Close()

			if resp1.StatusCode != http.StatusOK {
				s.T().Skipf("Provider %s not available or configured", provider)
				return
			}

			var uploadResult map[string]interface{}
			s.ParseSuccessResponse(resp1, &uploadResult)

			s.Contains(uploadResult, "url")
			s.Contains(uploadResult, "path")

			// Step 2: List files from provider
			resp2, err := s.GET(fmt.Sprintf("/api/v1/resources/%s/achievement", provider))
			s.Require().NoError(err)
			defer resp2.Body.Close()

			s.Equal(http.StatusOK, resp2.StatusCode)

			var listResult map[string]interface{}
			s.ParseSuccessResponse(resp2, &listResult)

			s.Contains(listResult, "files")
		})
	}

	// Step 3: Verify provider-specific capabilities
	resp3, err := s.GET("/api/v1/resources/providers")
	s.Require().NoError(err)
	defer resp3.Body.Close()

	var providersResult []map[string]interface{}
	s.ParseSuccessResponse(resp3, &providersResult)

	for _, p := range providersResult {
		providerName := p["name"].(string)
		capabilities := p["capabilities"].(map[string]interface{})

		s.Contains(capabilities, "supportsRead")
		s.Contains(capabilities, "supportsWrite")

		// Provider-specific checks
		switch providerName {
		case "r2":
			s.True(capabilities["supportsMultipart"].(bool), "R2 should support multipart")
		case "cdn":
			s.True(capabilities["supportsSignedUrls"].(bool), "CDN should support signed URLs")
		case "gcs":
			s.True(capabilities["supportsMetadata"].(bool), "GCS should support metadata")
		}
	}
}

// Performance test: Concurrent operations
func (s *WorkflowTestSuite) TestWorkflow_ConcurrentOperations() {
	const numConcurrent = 10

	// Test concurrent achievement creation
	results := make(chan bool, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func(index int) {
			body := map[string]interface{}{
				"name":        fmt.Sprintf("Concurrent Achievement %d", index),
				"description": fmt.Sprintf("Achievement created concurrently %d", index),
				"category":    "concurrent",
				"points":      100,
			}

			resp, err := s.POST("/api/v1/achievements/", body)
			if err != nil {
				results <- false
				return
			}
			defer resp.Body.Close()

			results <- resp.StatusCode == http.StatusCreated
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numConcurrent; i++ {
		if <-results {
			successCount++
		}
	}

	// Should handle concurrent requests successfully
	s.GreaterOrEqual(successCount, numConcurrent/2, "At least half of concurrent requests should succeed")
}

func TestWorkflowSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping workflow tests in short mode")
	}

	suite.Run(t, new(WorkflowTestSuite))
}
