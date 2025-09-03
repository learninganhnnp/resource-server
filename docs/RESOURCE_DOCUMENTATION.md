# Resource Package Documentation

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Core Components](#core-components)
4. [Resource Definitions](#resource-definitions)
5. [Parameter System](#parameter-system)
6. [Provider System](#provider-system)
7. [Resolution Process](#resolution-process)
8. [Template System](#template-system)
9. [Error Handling](#error-handling)
10. [Usage Examples](#usage-examples)
11. [Testing](#testing)

## Overview

The `resource` package provides a flexible, provider-agnostic system for managing resource paths and URLs across different storage backends (CDN, Google Cloud Storage, etc.). It supports dynamic path resolution with parameters, scope-based routing, and both public and signed URL generation.

### Key Features

- **Multi-provider support**: CDN, Google Cloud Storage with extensible architecture
- **Scoped resources**: Global, Application, and Client Application scopes
- **Dynamic path resolution**: Template-based paths with parameter substitution
- **URL generation options**: Public URLs and signed URLs with expiration
- **Type-safe parameter handling**: Validated parameter resolution with fallbacks
- **Thread-safe operations**: Concurrent access support with proper synchronization

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     ResourceManager                            │
├─────────────────────────────────────────────────────────────────┤
│ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐   │
│ │PathDefinition   │ │PathURL          │ │PathDefinition   │   │
│ │Resolver         │ │Resolver         │ │Registry         │   │
│ └─────────────────┘ └─────────────────┘ └─────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
           │                      │                      │
           ▼                      ▼                      ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│   PathResolver  │ │ProviderRegistry │ │ScopeSelector    │
│                 │ │                 │ │                 │
│ ┌─────────────┐ │ │ ┌─────────────┐ │ │                 │
│ │Template     │ │ │ │CDN Provider │ │ │                 │
│ │Resolver     │ │ │ │GCS Provider │ │ │                 │
│ └─────────────┘ │ │ └─────────────┘ │ │                 │
└─────────────────┘ └─────────────────┘ └─────────────────┘
           │
           ▼
┌─────────────────┐
│Parameter        │
│Resolvers        │
│                 │
│ ┌─────────────┐ │
│ │Context      │ │
│ │Values       │ │
│ │Routing      │ │
│ │Definition   │ │
│ └─────────────┘ │
└─────────────────┘
```

## Core Components

### ResourceManager

**Location**: `resource.go:39`

The main entry point that orchestrates all resource resolution operations.

```go
type ResourceManager interface {
    PathURLResolver() PathURLResolver
    PathDefinitionResolver() PathDefinitionResolver
    GetDefinition(name string) (*PathDefinition, error)
    AddDefinition(definition PathDefinition) error
    RemoveDefinition(defName string) error
    GetAllDefinitions() []*PathDefinition
}
```

**Key responsibilities**:
- Manages path definitions registry
- Provides access to URL and definition resolvers
- Handles CRUD operations for path definitions

### ResolvedResource

**Location**: `resource.go:32`

Represents the final result of resource resolution.

```go
type ResolvedResource struct {
    URL          string                 // Generated URL
    ResolvedPath string                 // Resolved path template
    ExpiresAt    *time.Time            // Expiration time for signed URLs
    Provider     provider.ProviderName  // Used provider
}
```

## Resource Definitions

**Location**: `definition.go:24`

Path definitions define how resources are organized and accessed across different providers and scopes.

```go
type PathDefinition struct {
    Name          string                                    // Unique identifier
    DisplayName   string                                    // Human-readable name
    Description   string                                    // Description
    AllowedScopes []ScopeType                              // Priority-ordered scopes
    Patterns      map[provider.ProviderName]Pattern        // Provider-specific patterns
    Parameters    []*ParameterDefinition                   // Parameter definitions
    URLOptions    *provider.URLOptions                     // Default URL options
}
```

### Pattern Modes

**Location**: `definition.go:11`

```go
type PatternMode string

const (
    PatternModePublic  PatternMode = "public"  // URL generation only
    PatternModeStorage PatternMode = "storage" // Storage operations only
    PatternModeBoth    PatternMode = "both"    // Both URL + storage
)
```

### Example Definition

```go
var UserAvatarsPath = PathDefinition{
    Name:          "user-avatars",
    DisplayName:   "User Avatars",
    Description:   "User avatar images shared across all platforms",
    AllowedScopes: []ScopeType{ScopeGlobal},
    Patterns: map[provider.ProviderName]Pattern{
        provider.ProviderCDN: {
            Patterns: map[ScopeType]string{
                ScopeGlobal: "shared/global/users/avatars/{user_id}.png",
            },
            Mode: PatternModePublic,
        },
        provider.ProviderGCS: {
            Patterns: map[ScopeType]string{
                ScopeGlobal: "global-avatars/{user_id}.png",
            },
            Mode: PatternModeStorage,
        },
    },
    Parameters: []*ParameterDefinition{
        {Name: "user_id", Required: true, Description: "User identifier"},
    },
}
```

## Parameter System

### Parameter Types

**Location**: `parameter.go:17`

```go
const (
    ClientApp      ParameterName = "client_app"
    App            ParameterName = "app"
    ResourceID     ParameterName = "resource_id"
    Version        ParameterName = "version"
    Timestamp      ParameterName = "timestamp"
    ResourceFormat ParameterName = "format"
)
```

### Parameter Resolvers

**Location**: `parameter_resolvers.go`

The parameter resolution system supports multiple resolver types with fallback chains:

#### Context-Based Resolvers

- **ClientAppParameterResolver**: Resolves based on user agent context
- **AppParameterResolver**: Resolves based on application context

#### Value-Based Resolvers

- **ValuesParameterResolver**: Static key-value parameter mapping
- **SimpleValueParameterResolver**: Single parameter resolver
- **TimestampParameterResolver**: Time-based parameter generation

#### Composite Resolvers

- **RoutingParameterResolver**: Routes parameters to specific resolvers
- **DefinitionParameterResolver**: Applies parameter definitions with defaults and validation
- **ValidatingParameterResolver**: Adds validation to existing resolvers

### Parameter Resolution Chain

```go
// Example: Create a parameter resolver chain
fallbackResolver := DefaultFallbackParameterResolver(clientApps, apps)
userResolver := NewValuesParameterResolver(map[ParameterName]string{
    "user_id": "12345",
})
chainedResolver, _ := NewParameterResolvers(userResolver, fallbackResolver)
```

## Provider System

### Provider Interface

**Location**: `provider/provider.go:39`

```go
type Provider interface {
    URLProvider
    StorageProvider
}

type URLProvider interface {
    Name() ProviderName
    GenerateURL(path string, opts *URLOptions) (string, error)
}

type StorageProvider interface {
    Name() ProviderName
    UploadObject(ctx context.Context, path string, data []byte) error
    DeleteObject(ctx context.Context, path string) error
    ListObjects(ctx context.Context, prefix string) ([]string, error)
    GetObjectMetadata(ctx context.Context, path string) (map[string]string, error)
}
```

### CDN Provider

**Location**: `provider/cdn_provider.go`

Handles CDN URL generation with optional signed URL support:

```go
type CDNConfig struct {
    BaseURL    string        `yaml:"base_url"`
    SigningKey string        `yaml:"signing_key"`
    Expiry     time.Duration `yaml:"expiry"`
}
```

**Features**:
- Public URL generation
- HMAC-SHA256 signed URLs with expiration
- Query parameter-based signatures

### Google Cloud Storage Provider

**Location**: `provider/gcs_provider.go`

Full GCS integration with storage operations and signed URLs:

```go
type GCSConfig struct {
    BucketName         string        `yaml:"bucket_name"`
    CredentialsPath    string        `yaml:"credentials_path"`
    Expiry             time.Duration `yaml:"expiry"`
    ServiceAccountJSON string        `yaml:"service_account_json"`
}
```

**Features**:
- Storage operations (upload, delete, list, metadata)
- GCS native signed URLs
- Flexible authentication (file, JSON, ADC)

## Resolution Process

### Scope Selection

**Location**: `scope_selector.go:24`

Scopes define the context hierarchy for resource access:

```go
const (
    ScopeGlobal    ScopeType = "G"   // Global resources
    ScopeApp       ScopeType = "A"   // Application-specific
    ScopeClientApp ScopeType = "CA"  // Client application-specific
)
```

**Selection logic**:
1. Check context requirements for each allowed scope
2. Return first scope with satisfied context requirements
3. Global scope has no requirements (always available)
4. App scope requires valid app context
5. ClientApp scope requires valid user agent context

### Path Resolution Flow

1. **Definition Lookup**: Find path definition by name
2. **Scope Selection**: Determine appropriate scope from context
3. **Provider Selection**: Choose provider (explicit or default)
4. **Pattern Resolution**: Get template pattern for scope/provider combination
5. **Parameter Resolution**: Resolve all template parameters
6. **Template Processing**: Apply parameters to pattern template
7. **URL Generation**: Generate final URL with provider-specific options

## Template System

**Location**: `template_resolver.go`

### Template Resolvers

#### BracesTemplateResolver
Uses `{parameter}` syntax for parameter placeholders:

```go
resolver, _ := NewBracesTemplateResolver()
url, _ := resolver.Resolve(ctx, "users/{user_id}/avatar.{format}", params)
```

#### NamedGroupsTemplateResolver
Uses `:parameter` syntax for parameter placeholders:

```go
resolver, _ := NewNamedGroupsTemplateResolver()
url, _ := resolver.Resolve(ctx, "users/:user_id/avatar.:format", params)
```

### Safety Features

**Location**: `template_resolver.go:96`

- **Path traversal protection**: Detects `../` patterns
- **Length limits**: Maximum 2000 characters for paths, 500 for parameter values
- **Character validation**: Prevents dangerous characters in parameters
- **Absolute path enforcement**: All resolved paths must start with `/`

### Template Validation

```go
type TemplateResolver interface {
    Resolve(ctx context.Context, template string, params map[ParameterName]string) (string, error)
    ResolveWith(ctx context.Context, template string, resolver ParameterResolver) (string, error)
    Validate(ctx context.Context, template string) error
}
```

## Error Handling

**Location**: `model.go:7`

The package defines specific error types for different failure scenarios:

```go
var (
    ErrDefinitionNotFound       = anerror.NewNotFoundError(8001, ...)
    ErrDefinitionExists         = anerror.NewInvalidArgumentError(8002, ...)
    ErrInvalidDefinition        = anerror.NewInvalidArgumentError(8003, ...)
    ErrParameterResolutionFailed = anerror.NewInternalError(8004, ...)
    ErrConditionNotMet          = anerror.NewInvalidArgumentError(8005, ...)
    ErrScopeNotAllowed          = anerror.NewInvalidArgumentError(8006, ...)
    ErrTemplateResolutionFailed = anerror.NewInternalError(8007, ...)
    ErrInvalidTemplate          = anerror.NewInvalidArgumentError(8008, ...)
)
```

## Usage Examples

### Basic Setup

```go
// Create components
templateResolver, _ := NewBracesTemplateResolver()
fallbackResolver := DefaultFallbackParameterResolver(clientApps, apps)
pathResolver := NewPathResolver(templateResolver, fallbackResolver)
scopeSelector := NewScopeSelector()
definitionRegistry := NewPathDefinitionRegistry()
providerRegistry := provider.NewRegistry()

// Register providers
cdnProvider := provider.NewCDNProvider(cdnConfig)
gcsProvider, _ := provider.NewGCSProvider(ctx, gcsConfig)
providerRegistry.Register(provider.ProviderCDN, cdnProvider)
providerRegistry.Register(provider.ProviderGCS, gcsProvider)

// Create resource manager
resourceManager := NewResourceManager(
    pathResolver,
    scopeSelector,
    definitionRegistry,
    providerRegistry,
)

// Add path definitions
resourceManager.AddDefinition(UserAvatarsPath)
```

### Resolving Resources

```go
// Simple resolution with parameter values
opts := ResolveOptions{}.WithValues(map[ParameterName]string{
    "user_id": "12345",
})

result, err := resourceManager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &opts)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("URL: %s\n", result.URL)
fmt.Printf("Path: %s\n", result.ResolvedPath)
fmt.Printf("Provider: %s\n", result.Provider)
```

### Signed URL Generation

```go
// Generate signed URL with custom expiration
urlOpts := &provider.URLOptions{
    SignedURL:    true,
    SignedExpiry: 1 * time.Hour,
}

opts := ResolveOptions{
    URLOptions: urlOpts,
}.WithValues(map[ParameterName]string{
    "user_id": "12345",
})

result, err := resourceManager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &opts)
```

### Provider-Specific Resolution

```go
// Force specific provider
opts := ResolveOptions{
    Provider: &provider.ProviderGCS,
}.WithValues(map[ParameterName]string{
    "user_id": "12345",
})

result, err := resourceManager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &opts)
```

## Testing

### Test Structure

- **Unit tests**: `*_test.go` files for individual component testing
- **Integration tests**: Provider-specific integration testing
- **Test helpers**: Located in `tests/path_definitions.go`

### Example Test Path Definitions

**Location**: `tests/path_definitions.go:30`

```go
var UserAvatarsPath = PathDefinition{
    Name:          "user-avatars",
    DisplayName:   "User Avatars", 
    Description:   "User avatar images shared across all platforms",
    AllowedScopes: []ScopeType{ScopeGlobal},
    Patterns: map[provider.ProviderName]Pattern{
        provider.ProviderCDN: {
            Patterns: map[ScopeType]string{
                ScopeGlobal: "shared/global/users/avatars/{user_id}.png",
            },
            Mode: PatternModePublic,
        },
    },
    Parameters: []*ParameterDefinition{
        {Name: "user_id", Required: true, Description: "User identifier"},
    },
}
```

### Running Tests

```bash
# Run all resource package tests
go test -v ./resource/...

# Run specific test files
go test -v ./resource -run TestPathDefinition
go test -v ./resource -run TestParameterResolvers
go test -v ./resource -run TestTemplateResolver
```

### Key Test Coverage

1. **Path Definition Validation**: Ensures definitions are properly structured
2. **Parameter Resolution**: Tests all resolver types and chains
3. **Template Processing**: Validates template syntax and parameter substitution
4. **Provider Integration**: Tests URL generation and storage operations
5. **Scope Selection**: Verifies context-based scope resolution
6. **Error Scenarios**: Tests error handling for various failure cases

---

This documentation provides comprehensive coverage of the resource package's architecture, components, and usage patterns. The package follows Go best practices with clear interfaces, comprehensive error handling, and extensive test coverage.