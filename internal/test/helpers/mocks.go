package helpers

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// MockStorageProvider mocks storage provider operations without external dependencies
type MockStorageProvider struct {
	Files             map[string][]byte
	ShouldReturnError bool
	ErrorToReturn     error
}

func NewMockStorageProvider() *MockStorageProvider {
	return &MockStorageProvider{
		Files: make(map[string][]byte),
	}
}

func (m *MockStorageProvider) GenerateUploadURL(ctx context.Context, path string, expiry time.Duration, metadata map[string]string) (string, error) {
	if m.ShouldReturnError {
		return "", m.ErrorToReturn
	}
	
	// Generate a mock signed URL
	u := &url.URL{
		Scheme: "https",
		Host:   "mock-storage.example.com",
		Path:   path,
	}
	
	q := u.Query()
	q.Set("expires", fmt.Sprintf("%d", time.Now().Add(expiry).Unix()))
	for k, v := range metadata {
		q.Set("x-meta-"+k, v)
	}
	u.RawQuery = q.Encode()
	
	return u.String(), nil
}

func (m *MockStorageProvider) GenerateDownloadURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	if m.ShouldReturnError {
		return "", m.ErrorToReturn
	}
	
	// Check if file exists
	if _, exists := m.Files[path]; !exists {
		return "", fmt.Errorf("file not found: %s", path)
	}
	
	// Generate a mock signed URL
	u := &url.URL{
		Scheme: "https",
		Host:   "mock-storage.example.com",
		Path:   path,
	}
	
	q := u.Query()
	q.Set("expires", fmt.Sprintf("%d", time.Now().Add(expiry).Unix()))
	u.RawQuery = q.Encode()
	
	return u.String(), nil
}

func (m *MockStorageProvider) DeleteFile(ctx context.Context, path string) error {
	if m.ShouldReturnError {
		return m.ErrorToReturn
	}
	
	delete(m.Files, path)
	return nil
}

func (m *MockStorageProvider) ListFiles(ctx context.Context, prefix string, maxKeys int, continuationToken string) ([]FileInfo, string, error) {
	if m.ShouldReturnError {
		return nil, "", m.ErrorToReturn
	}
	
	var files []FileInfo
	for path, data := range m.Files {
		if len(prefix) == 0 || len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			files = append(files, FileInfo{
				Key:          path,
				Size:         int64(len(data)),
				LastModified: time.Now(),
				ContentType:  "application/octet-stream",
			})
		}
	}
	
	// Simple pagination
	if len(files) > maxKeys {
		return files[:maxKeys], "next-token", nil
	}
	
	return files, "", nil
}

func (m *MockStorageProvider) GetFileMetadata(ctx context.Context, path string) (map[string]string, error) {
	if m.ShouldReturnError {
		return nil, m.ErrorToReturn
	}
	
	if _, exists := m.Files[path]; !exists {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	
	metadata := map[string]string{
		"content-type":   "application/octet-stream",
		"content-length": fmt.Sprintf("%d", len(m.Files[path])),
		"last-modified":  time.Now().Format(time.RFC3339),
	}
	
	return metadata, nil
}

func (m *MockStorageProvider) UpdateFileMetadata(ctx context.Context, path string, metadata map[string]string) error {
	if m.ShouldReturnError {
		return m.ErrorToReturn
	}
	
	if _, exists := m.Files[path]; !exists {
		return fmt.Errorf("file not found: %s", path)
	}
	
	return nil
}

func (m *MockStorageProvider) SimulateUpload(path string, data []byte) {
	m.Files[path] = data
}

func (m *MockStorageProvider) FileExists(path string) bool {
	_, exists := m.Files[path]
	return exists
}

type FileInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ContentType  string
	ETag         string
}

// MockMultipartUpload mocks multipart upload operations
type MockMultipartUpload struct {
	UploadID string
	Path     string
	Parts    map[int][]byte
	Metadata map[string]string
}

func NewMockMultipartUpload(uploadID, path string) *MockMultipartUpload {
	return &MockMultipartUpload{
		UploadID: uploadID,
		Path:     path,
		Parts:    make(map[int][]byte),
		Metadata: make(map[string]string),
	}
}

func (m *MockMultipartUpload) AddPart(partNumber int, data []byte) {
	m.Parts[partNumber] = data
}

func (m *MockMultipartUpload) Complete() ([]byte, error) {
	if len(m.Parts) == 0 {
		return nil, fmt.Errorf("no parts uploaded")
	}
	
	// Combine all parts
	var combined []byte
	for i := 1; i <= len(m.Parts); i++ {
		part, exists := m.Parts[i]
		if !exists {
			return nil, fmt.Errorf("missing part %d", i)
		}
		combined = append(combined, part...)
	}
	
	return combined, nil
}

func (m *MockMultipartUpload) Abort() error {
	m.Parts = make(map[int][]byte)
	return nil
}