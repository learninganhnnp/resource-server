package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/anh-nguyen/resource-server/internal/test/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type AchievementTestSuite struct {
	E2ETestSuite
	testDB *helpers.TestDatabase
}

func (s *AchievementTestSuite) SetupSuite() {
	s.E2ETestSuite.SetupSuite()
	s.testDB = helpers.SetupTestDatabase(s.T())
}

func (s *AchievementTestSuite) TearDownSuite() {
	if s.testDB != nil {
		s.testDB.Close()
	}
	s.E2ETestSuite.TearDownSuite()
}

func (s *AchievementTestSuite) SetupTest() {
	s.testDB.Cleanup(s.T())
}

// Test Cases for Achievement APIs

// AC-001: List all achievements
func (s *AchievementTestSuite) TestListAchievements_Success() {
	resp, err := s.GET("/api/v1/achievements/")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]any
	s.ParseSuccessResponse(resp, &result)

	s.Contains(result, "achievements")
	s.Contains(result, "pagination")

	achievements := result["achievements"].([]any)
	pagination := result["pagination"].(map[string]any)

	s.NotNil(achievements)
	s.Contains(pagination, "page")
	s.Contains(pagination, "pageSize")
	s.Contains(pagination, "total")
}

// AC-002: List with pagination
func (s *AchievementTestSuite) TestListAchievements_Pagination() {
	resp, err := s.GET("/api/v1/achievements/?page=1&pageSize=5")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]any
	s.ParseSuccessResponse(resp, &result)

	achievements := result["achievements"].([]any)
	pagination := result["pagination"].(map[string]any)

	s.LessOrEqual(len(achievements), 5)
	s.Equal(float64(1), pagination["page"])
	s.Equal(float64(5), pagination["pageSize"])
}

// AC-003: List only active achievements
func (s *AchievementTestSuite) TestListAchievements_OnlyActive() {
	resp, err := s.GET("/api/v1/achievements/?only_active=true")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]any
	s.ParseSuccessResponse(resp, &result)

	achievements := result["achievements"].([]any)

	// All returned achievements should be active
	for _, a := range achievements {
		achievement := a.(map[string]any)
		s.True(achievement["is_active"].(bool))
	}
}

// AC-004: List with custom page size
func (s *AchievementTestSuite) TestListAchievements_CustomPageSize() {
	resp, err := s.GET("/api/v1/achievements/?pageSize=10")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]any
	s.ParseSuccessResponse(resp, &result)

	achievements := result["achievements"].([]any)
	pagination := result["pagination"].(map[string]any)

	s.LessOrEqual(len(achievements), 10)
	s.Equal(float64(10), pagination["pageSize"])
}

// AC-005: Invalid page number
func (s *AchievementTestSuite) TestListAchievements_InvalidPage() {
	resp, err := s.GET("/api/v1/achievements/?page=0")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]any
	s.ParseSuccessResponse(resp, &result)

	pagination := result["pagination"].(map[string]any)

	// Should default to page 1
	s.Equal(float64(1), pagination["page"])
}

// AC-006: Page size exceeds limit
func (s *AchievementTestSuite) TestListAchievements_PageSizeLimit() {
	resp, err := s.GET("/api/v1/achievements/?pageSize=200")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]any
	s.ParseSuccessResponse(resp, &result)

	pagination := result["pagination"].(map[string]any)

	// Should be capped at 100
	s.LessOrEqual(pagination["pageSize"].(float64), 100)
}

// AC-007: Create achievement without icon
func (s *AchievementTestSuite) TestCreateAchievement_WithoutIcon() {
	body := map[string]any{
		"name":        "Test Achievement",
		"description": "A test achievement",
		"category":    "test",
		"points":      100,
	}

	resp, err := s.POST("/api/v1/achievements/", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusCreated, resp.StatusCode)

	var achievement map[string]any
	s.ParseSuccessResponse(resp, &achievement)

	s.Equal("Test Achievement", achievement["name"])
	s.Equal("A test achievement", achievement["description"])
	s.Equal("test", achievement["category"])
	s.Equal(float64(100), achievement["points"])
	s.NotEmpty(achievement["id"])
}

// AC-008: Create with all fields
func (s *AchievementTestSuite) TestCreateAchievement_AllFields() {
	body := map[string]any{
		"name":        "Complete Achievement",
		"description": "Achievement with all fields",
		"category":    "fitness",
		"points":      250,
		"iconFormat":  "png",
		"provider":    "r2",
	}

	resp, err := s.POST("/api/v1/achievements/", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusCreated, resp.StatusCode)

	var result map[string]any
	s.ParseSuccessResponse(resp, &result)

	s.Contains(result, "upload")
	s.Equal("Complete Achievement", result["name"])

	// Should provide upload info for icon
	upload := result["upload"].(map[string]any)
	s.Contains(upload, "upload_url")
	s.NotEmpty(upload["upload_url"])
}

// AC-009: Create with icon format
func (s *AchievementTestSuite) TestCreateAchievement_IconFormat() {
	body := map[string]any{
		"name":        "Icon Achievement",
		"description": "Achievement with icon",
		"category":    "design",
		"points":      150,
		"iconFormat":  "webp",
		"provider":    "r2",
	}

	resp, err := s.POST("/api/v1/achievements/", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusCreated, resp.StatusCode)

	var result map[string]any
	s.ParseSuccessResponse(resp, &result)

	// Should include upload info for specified format
	s.Contains(result, "upload")
	upload := result["upload"].(map[string]any)
	s.Contains(upload, "upload_url")
	s.NotEmpty(upload["upload_url"])
}

// AC-010: Missing required name
func (s *AchievementTestSuite) TestCreateAchievement_MissingName() {
	body := map[string]any{
		"description": "Achievement without name",
		"category":    "test",
		"points":      100,
	}

	resp, err := s.POST("/api/v1/achievements/", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// AC-011: Invalid points range
func (s *AchievementTestSuite) TestCreateAchievement_InvalidPoints() {
	body := map[string]any{
		"name":        "Invalid Points",
		"description": "Achievement with invalid points",
		"category":    "test",
		"points":      -50,
	}

	resp, err := s.POST("/api/v1/achievements/", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// AC-012: Invalid icon format
func (s *AchievementTestSuite) TestCreateAchievement_InvalidIconFormat() {
	body := map[string]any{
		"name":        "Invalid Format",
		"description": "Achievement with invalid icon format",
		"category":    "test",
		"points":      100,
		"iconFormat":  "invalid",
		"provider":    "r2",
	}

	resp, err := s.POST("/api/v1/achievements/", body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// AC-013: Duplicate achievement name
// func (s *AchievementTestSuite) TestCreateAchievement_DuplicateName() {
// 	// Create first achievement
// 	body1 := map[string]any{
// 		"name":        "Unique Name",
// 		"description": "First achievement",
// 		"category":    "test",
// 		"points":      100,
// 	}

// 	resp1, err := s.POST("/api/v1/achievements/", body1)
// 	s.Require().NoError(err)
// 	defer resp1.Body.Close()
// 	s.Equal(http.StatusCreated, resp1.StatusCode)

// 	// Try to create duplicate
// 	body2 := map[string]any{
// 		"name":        "Unique Name",
// 		"description": "Second achievement",
// 		"category":    "test",
// 		"points":      200,
// 	}

// 	resp2, err := s.POST("/api/v1/achievements/", body2)
// 	s.Require().NoError(err)
// 	defer resp2.Body.Close()

// 	s.Equal(http.StatusConflict, resp2.StatusCode)
// 	helpers.AssertErrorResponse(s.T(), resp2, "")
// }

// AC-014: Get existing achievement
func (s *AchievementTestSuite) TestGetAchievement_Success() {
	// First create an achievement
	body := map[string]any{
		"name":        "Get Test Achievement",
		"description": "Achievement for get test",
		"category":    "test",
		"points":      175,
	}

	resp1, err := s.POST("/api/v1/achievements/", body)
	s.Require().NoError(err)
	defer resp1.Body.Close()

	var createResult map[string]any
	s.ParseSuccessResponse(resp1, &createResult)

	achievement := createResult["achievement"].(map[string]any)
	achievementID := achievement["id"].(string)

	// Now get the achievement
	resp2, err := s.GET(fmt.Sprintf("/api/v1/achievements/%s", achievementID))
	s.Require().NoError(err)
	defer resp2.Body.Close()

	s.Equal(http.StatusOK, resp2.StatusCode)

	var getResult map[string]any
	s.ParseSuccessResponse(resp2, &getResult)

	retrievedAchievement := getResult["achievement"].(map[string]any)
	s.Equal(achievementID, retrievedAchievement["id"])
	s.Equal("Get Test Achievement", retrievedAchievement["name"])
}

// AC-015: Get non-existent achievement
func (s *AchievementTestSuite) TestGetAchievement_NotFound() {
	nonExistentID := uuid.New().String()

	resp, err := s.GET(fmt.Sprintf("/api/v1/achievements/%s", nonExistentID))
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusNotFound, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// AC-016: Invalid UUID format
func (s *AchievementTestSuite) TestGetAchievement_InvalidUUID() {
	resp, err := s.GET("/api/v1/achievements/invalid-uuid")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// AC-017: Verify all fields returned
func (s *AchievementTestSuite) TestGetAchievement_AllFields() {
	body := map[string]any{
		"name":        "Full Field Achievement",
		"description": "Achievement with all fields",
		"category":    "complete",
		"points":      300,
	}

	resp1, err := s.POST("/api/v1/achievements/", body)
	s.Require().NoError(err)
	defer resp1.Body.Close()

	var result map[string]any
	s.ParseSuccessResponse(resp1, &result)

	achievementID := result["id"].(string)

	// Get the achievement
	resp2, err := s.GET(fmt.Sprintf("/api/v1/achievements/%s", achievementID))
	s.Require().NoError(err)
	defer resp2.Body.Close()

	var getResult map[string]any
	s.ParseSuccessResponse(resp2, &getResult)

	retrievedAchievement := getResult

	// Verify all expected fields
	s.Contains(retrievedAchievement, "id")
	s.Contains(retrievedAchievement, "name")
	s.Contains(retrievedAchievement, "description")
	s.Contains(retrievedAchievement, "category")
	s.Contains(retrievedAchievement, "points")
	s.Contains(retrievedAchievement, "isActive")
	s.Contains(retrievedAchievement, "createdAt")
	s.Contains(retrievedAchievement, "updatedAt")
}

// AC-018-021: Update achievement icon
func (s *AchievementTestSuite) TestUpdateAchievementIcon_Success() {
	// First create an achievement
	body := map[string]any{
		"name":        "Icon Update Achievement",
		"description": "Achievement for icon update test",
		"category":    "test",
		"points":      200,
	}

	resp1, err := s.POST("/api/v1/achievements/", body)
	s.Require().NoError(err)
	defer resp1.Body.Close()

	var achievement map[string]any
	s.ParseSuccessResponse(resp1, &achievement)
	achievementID := achievement["id"].(string)

	// Update icon
	updateBody := map[string]any{
		"format":   "webp",
		"provider": "r2",
	}

	resp2, err := s.PUT(fmt.Sprintf("/api/v1/achievements/%s/icon", achievementID), updateBody)
	s.Require().NoError(err)
	defer resp2.Body.Close()

	s.Equal(http.StatusOK, resp2.StatusCode)

	var updateResult map[string]any
	s.ParseSuccessResponse(resp2, &updateResult)

	s.Contains(updateResult, "upload_url")
}

func TestAchievementSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping achievement tests in short mode")
	}

	suite.Run(t, new(AchievementTestSuite))
}
