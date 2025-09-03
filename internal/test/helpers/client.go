package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *APIClient) Get(path string, params map[string]string) (*http.Response, error) {
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return nil, err
	}
	
	if params != nil {
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}
	
	return c.HTTPClient.Get(u.String())
}

func (c *APIClient) Post(path string, body interface{}) (*http.Response, error) {
	return c.doJSON("POST", path, body)
}

func (c *APIClient) Put(path string, body interface{}) (*http.Response, error) {
	return c.doJSON("PUT", path, body)
}

func (c *APIClient) Delete(path string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	
	return c.HTTPClient.Do(req)
}

func (c *APIClient) doJSON(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(data)
	}
	
	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	
	return c.HTTPClient.Do(req)
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Error   *ErrorInfo  `json:"error"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

func ParseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	
	if !response.Success {
		if response.Error != nil {
			return fmt.Errorf("API error: %s - %s", response.Error.Code, response.Error.Message)
		}
		return fmt.Errorf("API error: unknown error")
	}
	
	if target != nil && response.Data != nil {
		dataBytes, err := json.Marshal(response.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		
		if err := json.Unmarshal(dataBytes, target); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}
	
	return nil
}