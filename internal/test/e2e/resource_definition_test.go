package e2e

import (
	"net/http"
	"testing"

	"github.com/anh-nguyen/resource-server/internal/test/helpers"
	"github.com/stretchr/testify/suite"
)

type ResourceDefinitionTestSuite struct {
	E2ETestSuite
}

// Test Cases for Resource Definition APIs

// RD-001: List all definitions
func (s *ResourceDefinitionTestSuite) TestListDefinitions_Success() {
	resp, err := s.GET("/api/v1/resources/definitions")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)
	helpers.AssertSuccessResponse(s.T(), resp)

	var definitions []map[string]interface{}
	s.ParseSuccessResponse(resp, &definitions)

	// Should have at least achievement and workout definitions
	s.GreaterOrEqual(len(definitions), 2)

	// Find achievement definition
	var achievementDef map[string]interface{}
	for _, def := range definitions {
		if def["name"] == "achievement" {
			achievementDef = def
			break
		}
	}

	s.NotNil(achievementDef, "Achievement definition should exist")
}

// RD-002: Verify definition structure
func (s *ResourceDefinitionTestSuite) TestListDefinitions_Structure() {
	resp, err := s.GET("/api/v1/resources/definitions")
	s.Require().NoError(err)
	defer resp.Body.Close()

	var definitions []map[string]interface{}
	s.ParseSuccessResponse(resp, &definitions)

	for _, def := range definitions {
		// Verify required fields
		s.Contains(def, "name")
		s.Contains(def, "displayName")
		s.Contains(def, "description")
		s.Contains(def, "parameters")

		// Verify types
		s.IsType("", def["name"])
		s.IsType("", def["displayName"])
		s.IsType("", def["description"])
		s.IsType([]interface{}{}, def["parameters"])
	}
}

// RD-003: Check allowed scopes
func (s *ResourceDefinitionTestSuite) TestListDefinitions_Scopes() {
	resp, err := s.GET("/api/v1/resources/definitions")
	s.Require().NoError(err)
	defer resp.Body.Close()

	var definitions []map[string]interface{}
	s.ParseSuccessResponse(resp, &definitions)

	for _, def := range definitions {
		s.Contains(def, "allowedScopes")
		scopes := def["allowedScopes"].([]interface{})

		// Verify valid scope values (G, A, CA)
		for _, scope := range scopes {
			s.Contains([]string{"G", "A", "CA"}, scope)
		}
	}
}

// RD-004: Validate providers list
func (s *ResourceDefinitionTestSuite) TestListDefinitions_Providers() {
	resp, err := s.GET("/api/v1/resources/definitions")
	s.Require().NoError(err)
	defer resp.Body.Close()

	var definitions []map[string]interface{}
	s.ParseSuccessResponse(resp, &definitions)

	for _, def := range definitions {
		s.Contains(def, "providers")
		providers := def["providers"].([]interface{})

		// Should include standard providers
		providerSet := make(map[string]bool)
		for _, p := range providers {
			providerSet[p.(string)] = true
		}

		s.True(providerSet["cdn"] || providerSet["gcs"] || providerSet["r2"],
			"Should have at least one standard provider")
	}
}

// RD-005: Get existing definition by name
func (s *ResourceDefinitionTestSuite) TestGetDefinition_Success() {
	resp, err := s.GET("/api/v1/resources/definitions/achievement")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var definition map[string]interface{}
	s.ParseSuccessResponse(resp, &definition)

	// Verify achievement definition
	s.Equal("achievement", definition["name"])
	s.Equal("Achievement", definition["displayName"])
	s.NotEmpty(definition["description"])

	// Verify parameters
	params := definition["parameters"].([]interface{})
	s.Greater(len(params), 0)

	// Check for achievementId parameter
	var hasAchievementID bool
	for _, p := range params {
		param := p.(map[string]interface{})
		if param["name"] == "achievementId" {
			hasAchievementID = true
			s.Contains(param, "rules")
			s.Contains(param, "description")
		}
	}
	s.True(hasAchievementID, "Should have achievementId parameter")
}

// RD-006: Get non-existent definition
func (s *ResourceDefinitionTestSuite) TestGetDefinition_NotFound() {
	resp, err := s.GET("/api/v1/resources/definitions/nonexistent")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusNotFound, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// RD-007: Invalid definition name format
func (s *ResourceDefinitionTestSuite) TestGetDefinition_InvalidFormat() {
	// Test with special characters
	resp, err := s.GET("/api/v1/resources/definitions/invalid@name!")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
	helpers.AssertErrorResponse(s.T(), resp, "")
}

// RD-008: Case sensitivity check
func (s *ResourceDefinitionTestSuite) TestGetDefinition_CaseSensitive() {
	// Try uppercase version
	resp, err := s.GET("/api/v1/resources/definitions/ACHIEVEMENT")
	s.Require().NoError(err)
	defer resp.Body.Close()

	// Should be case-sensitive and not found
	s.Equal(http.StatusNotFound, resp.StatusCode)

	// Try lowercase version (should work)
	resp2, err := s.GET("/api/v1/resources/definitions/achievement")
	s.Require().NoError(err)
	defer resp2.Body.Close()

	s.Equal(http.StatusOK, resp2.StatusCode)
}

// Additional test: Verify parameter validation rules
func (s *ResourceDefinitionTestSuite) TestDefinitionParameters_ValidationRules() {
	resp, err := s.GET("/api/v1/resources/definitions/achievement")
	s.Require().NoError(err)
	defer resp.Body.Close()

	var definition map[string]interface{}
	s.ParseSuccessResponse(resp, &definition)

	params := definition["parameters"].([]interface{})
	for _, p := range params {
		param := p.(map[string]interface{})

		if param["name"] == "achievementId" {
			rules := param["rules"].([]interface{})
			// Should have UUID validation rule
			s.Contains(rules, "uuid")
		}
	}
}

// Additional test: List definitions content type
func (s *ResourceDefinitionTestSuite) TestListDefinitions_ContentType() {
	resp, err := s.GET("/api/v1/resources/definitions")
	s.Require().NoError(err)
	defer resp.Body.Close()

	helpers.AssertContentType(s.T(), resp, "application/json")
}

func TestResourceDefinitionSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource definition tests in short mode")
	}

	suite.Run(t, new(ResourceDefinitionTestSuite))
}
