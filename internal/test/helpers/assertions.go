package helpers

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertSuccessResponse(t *testing.T, resp *http.Response) {
	t.Helper()
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")
	
	var response Response
	err = json.Unmarshal(body, &response)
	require.NoError(t, err, "Failed to parse response: %s", string(body))
	
	assert.True(t, response.Success, "Expected success response, got error: %v", response.Error)
	assert.Nil(t, response.Error, "Expected no error in success response")
}

func AssertErrorResponse(t *testing.T, resp *http.Response, expectedCode string) {
	t.Helper()
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")
	
	var response Response
	err = json.Unmarshal(body, &response)
	require.NoError(t, err, "Failed to parse response: %s", string(body))
	
	assert.False(t, response.Success, "Expected error response, got success")
	assert.NotNil(t, response.Error, "Expected error details")
	
	if expectedCode != "" && response.Error != nil {
		assert.Equal(t, expectedCode, response.Error.Code, "Unexpected error code")
	}
}

func AssertStatusCode(t *testing.T, resp *http.Response, expectedStatus int) {
	t.Helper()
	assert.Equal(t, expectedStatus, resp.StatusCode, "Unexpected status code")
}

func AssertContentType(t *testing.T, resp *http.Response, expectedType string) {
	t.Helper()
	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, expectedType, "Unexpected content type")
}

func ParseResponseData(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")
	
	var response Response
	err = json.Unmarshal(body, &response)
	require.NoError(t, err, "Failed to parse response: %s", string(body))
	
	require.True(t, response.Success, "Expected success response, got error: %v", response.Error)
	
	if target != nil && response.Data != nil {
		dataBytes, err := json.Marshal(response.Data)
		require.NoError(t, err, "Failed to marshal data")
		
		err = json.Unmarshal(dataBytes, target)
		require.NoError(t, err, "Failed to unmarshal data into target")
	}
}

func GetResponseError(t *testing.T, resp *http.Response) *ErrorInfo {
	t.Helper()
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")
	
	var response Response
	err = json.Unmarshal(body, &response)
	require.NoError(t, err, "Failed to parse response: %s", string(body))
	
	return response.Error
}