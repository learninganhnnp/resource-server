package dto

import (
	"fmt"
	"reflect"

	"avironactive.com/resource"
	"avironactive.com/resource/provider"
)

// NewPathDefinitionResponse creates PathDefinitionResponse from resource.PathDefinition
func NewPathDefinitionResponse(def resource.PathDefinition) PathDefinitionResponse {
	parameters := make([]PathParameterResponse, 0, len(def.Parameters))
	for _, param := range def.Parameters {
		// Convert validation.Rule slice to string slice
		var rules []string
		if param.Rules != nil {
			rules = make([]string, 0, len(param.Rules))
			for _, rule := range param.Rules {
				// Convert validation rule to string representation
				// Use reflection to get the rule type name
				ruleType := reflect.TypeOf(rule)
				if ruleType != nil {
					rules = append(rules, ruleType.String())
				} else {
					rules = append(rules, fmt.Sprintf("%T", rule))
				}
			}
		}

		parameters = append(parameters, PathParameterResponse{
			Name:         string(param.Name),
			Rules:        rules,
			Description:  param.Description,
			DefaultValue: param.DefaultValue,
		})
	}

	providers := make([]string, 0, len(def.Patterns))
	for provider := range def.Patterns {
		providers = append(providers, string(provider))
	}

	allowedScopes := make([]string, 0, len(def.AllowedScopes))
	for _, scope := range def.AllowedScopes {
		allowedScopes = append(allowedScopes, string(scope))
	}

	return PathDefinitionResponse{
		Name:          string(def.Name),
		DisplayName:   def.DisplayName,
		Description:   def.Description,
		AllowedScopes: allowedScopes,
		Parameters:    parameters,
		Providers:     providers,
	}
}

// NewMultipartInitResponse creates MultipartInitResponse from provider data
func NewMultipartInitResponse(uploadID, path, provider string, caps *provider.MultipartCapabilities) *MultipartInitResponse {
	response := &MultipartInitResponse{
		UploadID: uploadID,
		Path:     path,
		Provider: provider,
	}

	if caps != nil {
		response.MaxPartSize = caps.MaxPartSize
		response.MinPartSize = caps.MinPartSize
		response.MaxParts = caps.MaxParts
	}

	return response
}

// NewMultipartURLsResponse creates MultipartURLsResponse from provider data
func NewMultipartURLsResponse(partURLs []provider.PartURL, completeURL, abortURL *provider.ObjectURL) *MultipartURLsResponse {
	partResponses := make([]MultipartPartURL, 0, len(partURLs))
	for _, partURL := range partURLs {
		partResponses = append(partResponses, MultipartPartURL{
			PartNumber:        partURL.PartNumber,
			SignedURLResponse: *NewSignedURLResponse(&partURL.ObjectURL),
		})
	}

	return &MultipartURLsResponse{
		PartURLs:    partResponses,
		CompleteURL: *NewSignedURLResponse(completeURL),
		AbortURL:    *NewSignedURLResponse(abortURL),
	}
}

// NewAPIResponse creates a standard API response
func NewAPIResponse(success bool, data any, message string, err *APIError) APIResponse {
	return APIResponse{
		Success: success,
		Data:    data,
		Message: message,
		Error:   err,
	}
}

// NewSuccessResponse creates a successful API response
func NewSuccessResponse(data any) APIResponse {
	return NewAPIResponse(true, data, "", nil)
}

// NewSuccessResponseWithMessage creates a successful API response with message
func NewSuccessResponseWithMessage(data any, message string) APIResponse {
	return NewAPIResponse(true, data, message, nil)
}

// NewErrorResponse creates an error API response
func NewErrorResponse(code, message, details string) APIResponse {
	return NewAPIResponse(false, nil, message, &APIError{
		Code:    code,
		Details: details,
	})
}

// NewFileListResponseFromProvider creates FileListResponse from provider results
func NewFileListResponseFromProvider(result *provider.ListObjectsResult, maxKeys int) *FileListResponse {
	files := make([]FileInfo, 0, len(result.Objects))
	for _, obj := range result.Objects {
		files = append(files, FileInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			ETag:         obj.ETag,
			LastModified: *obj.LastModified,
		})
	}

	var nextContinuationToken string
	if result.NextContinuationToken != nil {
		nextContinuationToken = *result.NextContinuationToken
	}

	return &FileListResponse{
		Files:             files,
		ContinuationToken: nextContinuationToken,
		IsTruncated:       result.IsTruncated,
		MaxKeys:           maxKeys,
	}
}

// NewSignedURLResponseFromProvider creates SignedURLResponse from provider SignedURL
func NewSignedURLResponseFromProvider(resolvedResource *resource.ResolvedResource) *SignedURLResponse {
	return &SignedURLResponse{
		URL:         resolvedResource.URL.URL,
		Method:      resolvedResource.URL.Method,
		Headers:     resolvedResource.URL.Headers,
		ExpiresAt:   resolvedResource.URL.ExpiresAt,
		ResolvePath: resolvedResource.ResolvedPath,
	}
}

func NewSignedURLResponse(objectURL *provider.ObjectURL) *SignedURLResponse {
	return &SignedURLResponse{
		URL:       objectURL.URL,
		Method:    objectURL.Method,
		Headers:   objectURL.Headers,
		ExpiresAt: objectURL.ExpiresAt,
	}
}

// NewFileMetadataFromProvider creates FileMetadata from provider metadata
func NewFileMetadataFromProvider(metadata *provider.ObjectMetadata) *FileMetadata {
	// Convert checksums using factory function
	checksums := make([]ChecksumInfo, 0, len(metadata.Checksums))
	for _, checksum := range metadata.Checksums {
		checksums = append(checksums, NewChecksumInfoFromProvider(checksum))
	}

	// Convert storage class and ACL to strings
	var storageClass string
	if metadata.StorageClass != "" {
		storageClass = string(metadata.StorageClass)
	}

	var acl string
	if metadata.ACL != "" {
		acl = string(metadata.ACL)
	}

	return &FileMetadata{
		Key:                metadata.Key,
		Size:               metadata.Size,
		ContentType:        metadata.ContentType,
		ETag:               metadata.ETag,
		Created:            metadata.Created,
		LastModified:       metadata.LastModified,
		StorageClass:       storageClass,
		CacheControl:       metadata.CacheControl,
		ContentEncoding:    metadata.ContentEncoding,
		ContentDisposition: metadata.ContentDisposition,
		ContentLanguage:    metadata.ContentLanguage,
		Metadata:           metadata.Metadata,
		Checksums:          checksums,
		ACL:                acl,
		ExpirationTime:     metadata.ExpirationTime,
	}
}

// NewChecksumInfoFromProvider creates ChecksumInfo from provider.Checksum
func NewChecksumInfoFromProvider(checksum provider.Checksum) ChecksumInfo {
	return ChecksumInfo{
		Algorithm: string(checksum.Algorithm),
		Value:     checksum.Value,
	}
}
