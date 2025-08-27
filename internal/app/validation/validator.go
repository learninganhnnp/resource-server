package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate

	// Provider name validation - only allow known providers
	validProviders = map[string]bool{
		"cdn": true,
		"gcs": true,
		"r2":  true,
	}

	// File path validation - prevent directory traversal and invalid characters
	filePathRegex = regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
)

func init() {
	validate = validator.New()

	// Register custom validators
	validate.RegisterValidation("provider", validateProvider)
	validate.RegisterValidation("filepath", validateFilePath)
	validate.RegisterValidation("definition", validateDefinition)
	validate.RegisterValidation("duration", validateDuration)
}

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	messages := make([]string, len(e))
	for i, err := range e {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

// ValidateStruct validates a struct using go-playground/validator
func ValidateStruct(s interface{}) ValidationErrors {
	if err := validate.Struct(s); err != nil {
		var errors ValidationErrors
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, ValidationError{
				Field:   err.Field(),
				Message: formatValidationMessage(err),
			})
		}
		return errors
	}
	return nil
}

// ValidateProvider validates a provider name
func ValidateProvider(provider string) error {
	if provider == "" {
		return ValidationError{Field: "provider", Message: "provider is required"}
	}

	if !validProviders[provider] {
		return ValidationError{Field: "provider", Message: "invalid provider, must be one of: cdn, gcs, r2"}
	}

	return nil
}

// ValidateDefinition validates a resource definition name
func ValidateDefinition(definition string) error {
	if definition == "" {
		return ValidationError{Field: "definition", Message: "definition is required"}
	}

	if len(definition) > 64 {
		return ValidationError{Field: "definition", Message: "definition name too long (max 64 characters)"}
	}

	// Basic alphanumeric with underscore and dash
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(definition) {
		return ValidationError{Field: "definition", Message: "definition name contains invalid characters"}
	}

	return nil
}

// ValidateFilePath validates a file path
func ValidateFilePath(filePath string) error {
	if filePath == "" {
		return ValidationError{Field: "filePath", Message: "file path is required"}
	}

	// Check for directory traversal
	if strings.Contains(filePath, "..") {
		return ValidationError{Field: "filePath", Message: "file path cannot contain '..' (directory traversal)"}
	}

	// Check for valid characters
	if !filePathRegex.MatchString(filePath) {
		return ValidationError{Field: "filePath", Message: "file path contains invalid characters"}
	}

	// Check length
	if len(filePath) > 1024 {
		return ValidationError{Field: "filePath", Message: "file path too long (max 1024 characters)"}
	}

	return nil
}

// ValidateListParameters validates list operation parameters
func ValidateListParameters(maxKeys int, continuationToken, prefix string) ValidationErrors {
	var errors ValidationErrors

	// Validate maxKeys
	if maxKeys < 1 {
		errors = append(errors, ValidationError{Field: "maxKeys", Message: "maxKeys must be at least 1"})
	}
	if maxKeys > 10000 {
		errors = append(errors, ValidationError{Field: "maxKeys", Message: "maxKeys cannot exceed 10000"})
	}

	// Validate continuation token length
	if len(continuationToken) > 1024 {
		errors = append(errors, ValidationError{Field: "continuationToken", Message: "continuation token too long (max 1024 characters)"})
	}

	// Validate prefix
	if len(prefix) > 512 {
		errors = append(errors, ValidationError{Field: "prefix", Message: "prefix too long (max 512 characters)"})
	}

	return errors
}

// Custom validator functions
func validateProvider(fl validator.FieldLevel) bool {
	provider := fl.Field().String()
	return validProviders[provider]
}

func validateFilePath(fl validator.FieldLevel) bool {
	filePath := fl.Field().String()
	if strings.Contains(filePath, "..") {
		return false
	}
	if len(filePath) > 1024 {
		return false
	}
	return filePathRegex.MatchString(filePath)
}

func validateDefinition(fl validator.FieldLevel) bool {
	definition := fl.Field().String()
	if len(definition) > 64 {
		return false
	}
	return regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(definition)
}

func validateDuration(fl validator.FieldLevel) bool {
	duration := fl.Field().String()
	if duration == "" {
		return true // Empty duration is valid for omitempty fields
	}
	_, err := time.ParseDuration(duration)
	return err == nil
}

// formatValidationMessage formats validation error messages
func formatValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field is required"
	case "min":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "alphanum":
		return "must contain only alphanumeric characters"
	case "duration":
		return "must be a valid duration (e.g., '1h', '30m', '10s')"
	case "provider":
		return "must be a valid provider (cdn, gcs, r2)"
	case "filepath":
		return "must be a valid file path"
	case "definition":
		return "must be a valid definition name"
	default:
		return fmt.Sprintf("validation failed on tag '%s'", fe.Tag())
	}
}
