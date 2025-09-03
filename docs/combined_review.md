# Resource Package: Comprehensive Documentation & Usage Guide

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Core Components](#core-components)
4. [Quick Start](#quick-start)
5. [Resource Management](#resource-management)
6. [Parameter System](#parameter-system)
7. [Template Resolution](#template-resolution)
8. [Provider System](#provider-system)
9. [Scope Management](#scope-management)
10. [Signed URLs](#signed-urls)
11. [Error Handling](#error-handling)
12. [Advanced Usage](#advanced-usage)
13. [Integration Examples](#integration-examples)
14. [Testing](#testing)
15. [Best Practices](#best-practices)
16. [Performance Optimization](#performance-optimization)

## Overview

The Resource package provides a unified system for managing and resolving resource paths in the Aviron game server ecosystem. It enables flexible resource path definitions with multi-provider support, parameterized templates, and scope-based resolution. The package handles URL generation for different providers (CDN, Google Cloud Storage) while supporting signed URLs, parameter resolution from context, and secure template processing.

### Key Features

- **Dynamic Path Resolution**: Template-based path generation with parameter substitution
- **Multi-Provider Support**: Unified interface supporting CDN and Google Cloud Storage providers
- **Scope-Based Access**: Hierarchical scope system (Global, App, ClientApp) for resource isolation
- **Signed URL Generation**: Support for time-limited signed URLs with configurable expiry
- **Parameter Validation**: Type-safe parameter handling with validation and fallback resolution
- **Template Security**: Built-in protection against path traversal and injection attacks
- **Thread-Safe Operations**: Concurrent access support with proper synchronization
- **Flexible Architecture**: Extensible provider system and parameter resolution chains

### Resource Resolution Flow

1. **Path Definition** → Contains patterns and metadata for resource paths
2. **Scope Selection** → Determines the appropriate scope (Global, App, ClientApp) based on context
3. **Parameter Resolution** → Resolves template parameters from various sources (context, values, fallbacks)
4. **Provider Selection** → Chooses storage provider based on URL type and preferences
5. **Template Processing** → Applies parameters to pattern templates with security validation
6. **URL Generation** → Creates final URLs with optional signing and expiry

## Architecture

The resource system follows a layered architecture:

```
ResourceManager
├── DefinitionResolver
│   ├── PathResolver
│   │   └── TemplateResolver
│   ├── ScopeSelector
│   └── PathDefinitionRegistry
└── URLResolver
    └── ProviderRegistry
```

### Architecture Components

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

**Location**: `resource.go:39-50`

Central interface for managing resource definitions and resolution.

```go
type ResourceManager interface {
    URLResolver() URLResolver
    DefinitionResolver() DefinitionResolver
    
    GetDefinition(name DefinitionName) (*PathDefinition, error)
    AddDefinition(definition PathDefinition) error
    RemoveDefinition(defName DefinitionName) error
    GetAllDefinitions() []*PathDefinition
}

// Constructor
func NewResourceManager(
    pathResolver PathResolver,
    scopeSelector ScopeSelector,
    definitionRegistry PathDefinitionRegistry,
    providerRegistry provider.Registry,
) ResourceManager
```

**Key responsibilities**:
- Manages path definitions registry
- Provides access to URL and definition resolvers
- Handles CRUD operations for path definitions
- Orchestrates all resource resolution operations

### PathDefinition

**Location**: `definition.go:24-44`

Defines resource path patterns with provider and scope support.

```go
type PathDefinition struct {
    Name          string                                    // Unique identifier
    DisplayName   string                                    // Human-readable name
    Description   string                                    // Description of the resource
    AllowedScopes []ScopeType                              // Permitted scopes (ordered by priority)
    Patterns      map[provider.ProviderName]PathPatterns   // Provider-specific patterns
    Parameters    []*ParameterDefinition                   // Parameter definitions
    URLOptions    *provider.URLOptions                     // Default URL generation options
}

type PathPatterns struct {
    Patterns map[ScopeType]string  // Scope-specific path patterns
    URLType  URLType               // Content or Operation URLs
}
```

### ResolveOptions

**Location**: `resource.go:8-14`

Configuration for resource resolution with fluent API.

```go
type ResolveOptions struct {
    URLOptions    *provider.URLOptions   // URL generation options (signed URL, expiry)
    Provider      *provider.ProviderName // Optional provider selection
    Scope         *ScopeType             // Optional scope override
    ParamResolver ParameterResolver      // Optional parameter resolver
    URLType       *URLType               // Intended URL type for provider selection
}

// Fluent API methods
func (opts ResolveOptions) WithValues(params map[ParameterName]string) *ResolveOptions
func (opts ResolveOptions) WithProvider(providerName provider.ProviderName) *ResolveOptions
func (opts ResolveOptions) WithURLOptions(urlOpts *provider.URLOptions) *ResolveOptions
func (opts ResolveOptions) WithURLType(urlType URLType) *ResolveOptions
```

### ResolvedResource

**Location**: `resource.go:32-40`

Result of resource resolution.

```go
type ResolvedResource struct {
    URL          string                 // Generated URL
    ResolvedPath string                 // Resolved path template
    Provider     provider.ProviderName  // Selected provider
}
```

## Quick Start

### Basic Setup

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "avironactive.com/resource"
    "avironactive.com/resource/provider"
)

func main() {
    // Create template resolver with security settings
    templateResolver := resource.NewBracesTemplateResolver()
    
    // Create fallback parameter resolver
    fallbackResolver := resource.DefaultFallbackParameterResolver(
        map[int16]string{1: "ios", 2: "android"}, // Client apps
        map[int16]string{1: "bike", 2: "rower"},  // Apps
    )
    
    // Create path resolver
    pathResolver := resource.NewPathResolver(templateResolver, fallbackResolver)
    
    // Create scope selector
    scopeSelector := resource.NewScopeSelector()
    
    // Create path definition registry
    definitionRegistry := resource.NewPathDefinitionRegistry()
    
    // Create provider registry
    providerRegistry := provider.NewRegistry()
    
    // Add CDN provider
    cdnProvider := provider.NewCDNProvider(provider.CDNConfig{
        BaseURL:    "https://cdn1.avironactive.com",
        SigningKey: "secret-key",
        Expiry:     1 * time.Hour,
    })
    err := providerRegistry.Register(cdnProvider)
    if err != nil {
        panic(err)
    }
    
    // Create resource manager
    manager := resource.NewResourceManager(
        pathResolver,
        scopeSelector,
        definitionRegistry,
        providerRegistry,
    )
    
    // Add a path definition
    definition := resource.PathDefinition{
        Name:        "user-avatar",
        DisplayName: "User Avatar",
        Description: "User profile avatar images",
        AllowedScopes: []resource.ScopeType{
            resource.ScopeClientApp,
            resource.ScopeApp,
            resource.ScopeGlobal,
        },
        Patterns: map[provider.ProviderName]resource.PathPatterns{
            provider.ProviderCDN: {
                Patterns: map[resource.ScopeType]string{
                    resource.ScopeClientApp: "/avatars/{client_app}/{app}/{resource_id}.{format}",
                    resource.ScopeApp:       "/avatars/shared/{app}/{resource_id}.{format}",
                    resource.ScopeGlobal:    "/avatars/global/{resource_id}.{format}",
                },
                URLType: resource.URLTypeContent,
            },
        },
        Parameters: []*resource.ParameterDefinition{
            {Name: resource.ResourceID, Required: true},
            {Name: resource.ResourceFormat, DefaultValue: "jpg"},
        },
    }
    
    err = manager.AddDefinition(definition)
    if err != nil {
        panic(err)
    }
    
    // Resolve a resource
    ctx := context.Background()
    opts := resource.ResolveOptions{}.WithValues(map[resource.ParameterName]string{
        resource.ResourceID:     "user123",
        resource.ResourceFormat: "png",
    })
    
    result, err := manager.DefinitionResolver().Resolve(ctx, "user-avatar", &opts)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("URL: %s\n", result.URL)
    fmt.Printf("Path: %s\n", result.ResolvedPath)
    fmt.Printf("Provider: %s\n", result.Provider)
}
```

## Resource Management

### Dynamic Path Definition Registration

```go
func registerWorkoutVideoDefinition(manager resource.ResourceManager) {
    workoutVideoDef := resource.PathDefinition{
        Name:        "workout-video",
        DisplayName: "Workout Video Content",
        Description: "Exercise videos with multi-platform support",
        AllowedScopes: []resource.ScopeType{
            resource.ScopeClientApp,
            resource.ScopeApp,
            resource.ScopeGlobal,
        },
        Patterns: map[provider.ProviderName]resource.PathPatterns{
            provider.ProviderCDN: {
                Patterns: map[resource.ScopeType]string{
                    resource.ScopeClientApp: "videos/{client_app}/{app}/workouts/{resource_id}.{format}",
                    resource.ScopeApp:       "videos/shared/{app}/workouts/{resource_id}.{format}",
                    resource.ScopeGlobal:    "videos/global/workouts/{resource_id}.{format}",
                },
                URLType: resource.URLTypeContent,
            },
            provider.ProviderGCS: {
                Patterns: map[resource.ScopeType]string{
                    resource.ScopeGlobal: "workout-videos/{resource_id}.{format}",
                },
                URLType: resource.URLTypeOperation,
            },
        },
        Parameters: []*resource.ParameterDefinition{
            {Name: resource.ResourceID, Required: true, Description: "Workout video ID"},
            {Name: resource.ResourceFormat, DefaultValue: "mp4", Description: "Video format"},
            {Name: resource.Version, DefaultValue: "1.0.0", Description: "Content version"},
            {Name: "duration", Required: false, Description: "Video duration in seconds"},
            {Name: "quality", DefaultValue: "hd", Description: "Video quality"},
        ],
        URLOptions: &provider.URLOptions{
            SignedURL:    false, // Public videos by default
            SignedExpiry: 0,
        },
    }
    
    err := manager.AddDefinition(workoutVideoDef)
    if err != nil {
        log.Printf("Failed to register workout video definition: %v", err)
        return
    }
    
    fmt.Printf("Successfully registered workout video definition\n")
}
```

### Pattern Modes

**Location**: `definition.go:11-15`

```go
type PatternMode string

const (
    PatternModePublic  PatternMode = "public"  // URL generation only
    PatternModeStorage PatternMode = "storage" // Storage operations only
    PatternModeBoth    PatternMode = "both"    // Both URL + storage
)
```

### Basic Resource Resolution

```go
func resolveBasicResource(manager resource.ResourceManager) {
    ctx := context.Background()
    
    // Basic resolution with minimal parameters
    opts := resource.ResolveOptions{}.WithValues(map[resource.ParameterName]string{
        resource.ResourceID: "avatar_123",
    })
    
    result, err := manager.DefinitionResolver().Resolve(ctx, "user-avatar", &opts)
    if err != nil {
        log.Printf("Failed to resolve resource: %v", err)
        return
    }
    
    fmt.Printf("Resolved URL: %s\n", result.URL)
    fmt.Printf("Resolved Path: %s\n", result.ResolvedPath)
    fmt.Printf("Provider Used: %s\n", result.Provider)
}
```

## Parameter System

### Parameter Names

**Location**: `parameter.go:17-23`

Predefined parameter constants:

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

### Parameter Definitions

**Location**: `parameter.go:46-51`

```go
type ParameterDefinition struct {
    Name         ParameterName // Parameter name
    DefaultValue string        // Default value if not provided
    Description  string        // Parameter description  
    Required     bool          // Whether parameter is required
}
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

```go
// Values-based resolver
valuesResolver := resource.NewValuesParameterResolver(map[resource.ParameterName]string{
    resource.ResourceID:     "12345",
    resource.Version:        "2.0.0",
    resource.ResourceFormat: "webp",
})

// Context-aware resolvers
clientAppResolver := resource.NewClientAppParameterResolver(map[int16]string{
    1: "ios",
    2: "android",
    3: "web",
})

appResolver := resource.NewAppParameterResolver(map[int16]string{
    1: "bike",
    2: "rower",
    3: "strength",
})

// Chain multiple resolvers
chainedResolver, err := resource.NewParameterResolvers(
    valuesResolver,
    clientAppResolver,
    appResolver,
)
if err != nil {
    log.Fatal(err)
}
```

### Custom Parameter Resolver

```go
// Custom parameter resolver for user-specific data
type UserParameterResolver struct {
    userID   string
    userData map[string]string
}

func NewUserParameterResolver(userID string, userData map[string]string) *UserParameterResolver {
    return &UserParameterResolver{
        userID:   userID,
        userData: userData,
    }
}

func (r *UserParameterResolver) Resolve(ctx context.Context, paramName resource.ParameterName) (string, error) {
    switch paramName {
    case "user_id":
        return r.userID, nil
    case "user_tier":
        return r.userData["tier"], nil
    case "subscription_level":
        return r.userData["subscription"], nil
    default:
        return "", nil // Let other resolvers handle it
    }
}
```

## Template Resolution

### Template System

**Location**: `template_resolver.go:24-28`

The template resolver supports multiple template formats with security validation:

```go
type TemplateResolver interface {
    Resolve(ctx context.Context, template string, resolver ParameterResolver) (string, error)
    Validate(ctx context.Context, template string) error
}
```

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

### Supported Template Formats

1. **Braces Templates**: `{parameter_name}` syntax
2. **Named Groups**: `:parameter_name` syntax (regex-based)

### Template Example

```go
func demonstrateTemplatePatterns(templateResolver resource.TemplateResolver) {
    ctx := context.Background()
    
    // Braces template format
    bracesTemplate := "/assets/{client_app}/{resource_id}.{format}"
    params := map[resource.ParameterName]string{
        "client_app":   "ios",
        "resource_id":  "icon_123",
        "format":       "png",
    }
    
    result, err := templateResolver.Resolve(ctx, bracesTemplate, NewValuesParameterResolver(params))
    if err != nil {
        log.Printf("Braces template failed: %v", err)
        return
    }
    fmt.Printf("Braces result: %s\n", result)
    // Result: "/assets/ios/icon_123.png"
}
```

### Security Features

**Location**: `template_resolver.go:96-106`

- **Path Length Limits**: Maximum 2000 characters for resolved paths
- **Parameter Value Limits**: Maximum 500 characters per parameter
- **Path Traversal Protection**: Blocks `../` and similar patterns
- **Dangerous Character Filtering**: Prevents null bytes and path separators in parameters
- **Absolute Path Enforcement**: All resolved paths must start with `/`

### Safety Patterns

Default patterns blocked in templates:
- `../` and `..\\` (path traversal)
- `/etc/`, `/proc/`, `/sys/` (system directories)
- `\\windows\\`, `\\system32\\` (Windows system paths)

## Provider System

### Provider Interface

**Location**: `provider/provider.go:39-51`

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
    ProjectID          string        `yaml:"project_id"`
    CredentialsPath    string        `yaml:"credentials_path"`
    Expiry             time.Duration `yaml:"expiry"`
    ServiceAccountJSON string        `yaml:"service_account_json"`
}
```

**Features**:
- Storage operations (upload, delete, list, metadata)
- GCS native signed URLs
- Flexible authentication (file, JSON, ADC)

### Provider Registry Configuration

```go
// Create provider registry
providerRegistry := provider.NewRegistry()

// Add CDN provider
cdnProvider := provider.NewCDNProvider("https://cdn.aviron.com")
err := providerRegistry.Register(cdnProvider)
if err != nil {
    log.Fatal(err)
}

// Add Google Cloud Storage provider
gcsProvider := provider.NewGCSProvider(&provider.GCSConfig{
    BucketName:      "aviron-resources",
    ProjectID:       "aviron-prod",
    CredentialsPath: "/path/to/credentials.json",
})
err = providerRegistry.Register(gcsProvider)
if err != nil {
    log.Fatal(err)
}

// Set default provider
err = providerRegistry.SetDefault(provider.ProviderCDN)
if err != nil {
    log.Fatal(err)
}
```

### Provider-Specific Resolution

```go
func resolveWithSpecificProvider(manager resource.ResourceManager) {
    ctx := context.Background()
    
    // Force use of GCS for backup operations
    opts := resource.ResolveOptions{}.
        WithProvider(provider.ProviderGCS).
        WithValues(map[resource.ParameterName]string{
            resource.ResourceID: "backup_20241201",
            resource.Version:    "1.0.0",
        })
    
    result, err := manager.DefinitionResolver().Resolve(ctx, "backup-file", &opts)
    if err != nil {
        log.Printf("Failed to resolve with GCS: %v", err)
        return
    }
    
    fmt.Printf("GCS URL: %s\n", result.URL)
    fmt.Printf("Provider: %s\n", result.Provider)
}
```

## Scope Management

### Scope Types

**Location**: `scope_selector.go:10-16`

Hierarchical scope system:

```go
const (
    ScopeGlobal    ScopeType = "G"   // Global resources (no context required)
    ScopeApp       ScopeType = "A"   // App-scoped resources (requires app context)
    ScopeClientApp ScopeType = "CA"  // Client app-scoped resources (requires client app context)
)
```

### Scope Selection

**Location**: `scope_selector.go:24-27`

Automatic scope selection based on context:

```go
type ScopeSelector interface {
    Select(ctx context.Context, definition *PathDefinition) (ScopeType, error)
    IsAllowed(ctx context.Context, scope ScopeType, definition *PathDefinition) (bool, error)
}
```

### Context Requirements

**Location**: `scope_selector.go:72-86`

- **ScopeGlobal**: No additional requirements
- **ScopeApp**: Requires valid app context with ID > 0
- **ScopeClientApp**: Requires valid user agent with ClientAppID > 0

**Selection logic**:
1. Check context requirements for each allowed scope
2. Return first scope with satisfied context requirements
3. Global scope has no requirements (always available)
4. App scope requires valid app context
5. ClientApp scope requires valid user agent context

### Automatic Scope Selection

```go
func resolveWithAutomaticScope(manager resource.ResourceManager) {
    // Setup context with app and client app information
    ctx := context.Background()
    ctx = app.WithApp(ctx, &app.App{ID: 1, Name: "bike"})
    ctx = useragent.WithUserAgent(ctx, &useragent.UserAgent{
        ClientAppID: 2, // Android
        Version:     "1.5.0",
    })
    
    // Let the system automatically select the most specific scope
    opts := resource.ResolveOptions{}.WithValues(map[resource.ParameterName]string{
        resource.ResourceID: "feature_config_123",
    })
    
    result, err := manager.DefinitionResolver().Resolve(ctx, "app-config", &opts)
    if err != nil {
        log.Printf("Scope resolution failed: %v", err)
        return
    }
    
    fmt.Printf("Auto-scoped URL: %s\n", result.URL)
    // Will likely resolve to ClientApp scope if available
}
```

### Explicit Scope Selection

```go
func resolveWithExplicitScope(manager resource.ResourceManager) {
    ctx := context.Background()
    
    // Force Global scope for shared resources
    globalScope := resource.ScopeGlobal
    globalOpts := resource.ResolveOptions{
        Scope: &globalScope,
    }.WithValues(map[resource.ParameterName]string{
        resource.ResourceID: "shared_achievement_badge",
    })
    
    globalResult, err := manager.DefinitionResolver().Resolve(ctx, "achievement-badge", &globalOpts)
    if err != nil {
        log.Printf("Global scope resolution failed: %v", err)
        return
    }
    
    fmt.Printf("Global resource: %s\n", globalResult.URL)
}
```

## Signed URLs

### Basic Signed URL Generation

```go
func generateSignedURL(manager resource.ResourceManager) {
    ctx := context.Background()
    
    // Generate signed URL with 1-hour expiry
    opts := resource.ResolveOptions{
        URLOptions: &provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: time.Hour,
        },
    }.WithValues(map[resource.ParameterName]string{
        resource.ResourceID: "secure_document_789",
    })
    
    result, err := manager.DefinitionResolver().Resolve(ctx, "secure-document", &opts)
    if err != nil {
        log.Printf("Failed to generate signed URL: %v", err)
        return
    }
    
    fmt.Printf("Signed URL: %s\n", result.URL)
    fmt.Printf("Expires in: %v\n", time.Hour)
}
```

### Conditional Signed URLs

```go
func conditionalSignedURL(manager resource.ResourceManager, isPrivate bool) {
    ctx := context.Background()
    
    var urlOptions *provider.URLOptions
    if isPrivate {
        urlOptions = &provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: 30 * time.Minute, // 30-minute expiry for private content
        }
    }
    
    opts := resource.ResolveOptions{
        URLOptions: urlOptions,
    }.WithValues(map[resource.ParameterName]string{
        resource.ResourceID: "content_123",
        "access_level": func() string {
            if isPrivate {
                return "private"
            }
            return "public"
        }(),
    })
    
    result, err := manager.DefinitionResolver().Resolve(ctx, "user-content", &opts)
    if err != nil {
        log.Printf("Resolution failed: %v", err)
        return
    }
    
    if isPrivate {
        fmt.Printf("Private content URL (signed): %s\n", result.URL)
    } else {
        fmt.Printf("Public content URL: %s\n", result.URL)
    }
}
```

### Dynamic Expiry Based on Content Type

```go
func getExpiryForContentType(contentType string) time.Duration {
    switch contentType {
    case "video":
        return 4 * time.Hour // Long-form content
    case "image":
        return 2 * time.Hour // Medium duration
    case "document":
        return 1 * time.Hour // Shorter for documents
    case "temp":
        return 15 * time.Minute // Very short for temporary content
    default:
        return time.Hour
    }
}

func resolveWithDynamicExpiry(manager resource.ResourceManager, contentType string) {
    ctx := context.Background()
    
    opts := resource.ResolveOptions{
        URLOptions: &provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: getExpiryForContentType(contentType),
        },
    }.WithValues(map[resource.ParameterName]string{
        resource.ResourceID: "content_456",
        "content_type":      contentType,
    })
    
    result, err := manager.DefinitionResolver().Resolve(ctx, "dynamic-content", &opts)
    if err != nil {
        log.Printf("Failed to resolve: %v", err)
        return
    }
    
    fmt.Printf("Content URL: %s\n", result.URL)
}
```

## Error Handling

### Resource-Specific Errors

**Location**: `model.go:7-22`

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

### Error Handling Patterns

```go
func handleResolutionErrors(manager resource.ResourceManager) {
    ctx := context.Background()
    
    opts := resource.ResolveOptions{}.WithValues(map[resource.ParameterName]string{
        resource.ResourceID: "test123",
    })
    
    result, err := manager.DefinitionResolver().Resolve(ctx, "user-avatar", &opts)
    if err != nil {
        // Check for specific error types
        if errors.Is(err, resource.ErrDefinitionNotFound) {
            fmt.Printf("Definition not found: %v\n", err)
            return
        }
        
        // Use anerror for structured error information
        if anerror.As(err).IsInvalidArgument() {
            fmt.Printf("Invalid argument: %v\n", err)
            return
        }
        
        fmt.Printf("Unexpected error: %v\n", err)
        return
    }
    
    fmt.Printf("Success: %s\n", result.URL)
}
```

### Error Recovery Strategies

```go
func demonstrateErrorRecovery(manager resource.ResourceManager) {
    ctx := context.Background()
    resourceID := "critical_asset_123"
    
    // Strategy 1: Provider fallback
    providers := []provider.ProviderName{
        provider.ProviderCDN,
        provider.ProviderGCS,
    }
    
    for _, providerName := range providers {
        opts := resource.ResolveOptions{}.
            WithProvider(providerName).
            WithValues(map[resource.ParameterName]string{
                resource.ResourceID: resourceID,
            })
        
        result, err := manager.DefinitionResolver().Resolve(ctx, "critical-asset", &opts)
        if err == nil {
            fmt.Printf("Success with provider %s: %s\n", providerName, result.URL)
            return
        }
        
        fmt.Printf("Provider %s failed: %v\n", providerName, err)
    }
    
    fmt.Printf("All providers failed\n")
}
```

## Advanced Usage

### Batch Resource Resolution

```go
func batchResolveResources(manager resource.ResourceManager) {
    ctx := context.Background()
    
    // Define batch requests
    requests := []struct {
        Name    string
        DefName string
        Opts    *resource.ResolveOptions
    }{
        {
            Name:    "User Avatar",
            DefName: "user-avatar",
            Opts: resource.ResolveOptions{}.WithValues(map[resource.ParameterName]string{
                resource.ResourceID: "user123",
            }),
        },
        {
            Name:    "Achievement Badge",
            DefName: "achievement-badge",
            Opts: resource.ResolveOptions{}.WithValues(map[resource.ParameterName]string{
                resource.ResourceID: "first_workout",
            }),
        },
        {
            Name:    "Workout Video",
            DefName: "workout-video",
            Opts: resource.ResolveOptions{
                URLOptions: &provider.URLOptions{
                    SignedURL:    true,
                    SignedExpiry: 2 * time.Hour,
                },
            }.WithValues(map[resource.ParameterName]string{
                resource.ResourceID: "beginner_bike_ride",
            }),
        },
    }
    
    // Resolve concurrently
    type result struct {
        Name     string
        Resource *resource.ResolvedResource
        Error    error
    }
    
    resultChan := make(chan result, len(requests))
    
    for _, req := range requests {
        go func(r struct {
            Name    string
            DefName string
            Opts    *resource.ResolveOptions
        }) {
            resource, err := manager.DefinitionResolver().Resolve(ctx, r.DefName, r.Opts)
            resultChan <- result{
                Name:     r.Name,
                Resource: resource,
                Error:    err,
            }
        }(req)
    }
    
    // Collect results
    for i := 0; i < len(requests); i++ {
        res := <-resultChan
        if res.Error != nil {
            fmt.Printf("%s: ERROR - %v\n", res.Name, res.Error)
        } else {
            fmt.Printf("%s: %s\n", res.Name, res.Resource.URL)
        }
    }
}
```

### Resource Resolution Middleware

```go
type ResolutionMiddleware func(resource.DefinitionResolver) resource.DefinitionResolver

type LoggingMiddleware struct {
    logger *log.Logger
    next   resource.DefinitionResolver
}

func NewLoggingMiddleware(logger *log.Logger) ResolutionMiddleware {
    return func(next resource.DefinitionResolver) resource.DefinitionResolver {
        return &LoggingMiddleware{
            logger: logger,
            next:   next,
        }
    }
}

func (m *LoggingMiddleware) Resolve(ctx context.Context, defName string, opts *resource.ResolveOptions) (*resource.ResolvedResource, error) {
    start := time.Now()
    
    m.logger.Printf("Resolving resource: %s", defName)
    
    result, err := m.next.Resolve(ctx, defName, opts)
    
    duration := time.Since(start)
    
    if err != nil {
        m.logger.Printf("Resolution failed: %s - %v (took %v)", defName, err, duration)
    } else {
        m.logger.Printf("Resolution success: %s -> %s (took %v)", defName, result.URL, duration)
    }
    
    return result, err
}

// Usage
func useMiddleware(baseResolver resource.DefinitionResolver, logger *log.Logger) resource.DefinitionResolver {
    // Chain middleware
    resolver := baseResolver
    resolver = NewLoggingMiddleware(logger)(resolver)
    
    return resolver
}
```

## Integration Examples

### HTTP Handler Integration

```go
func createResourceHandler(manager resource.ResourceManager) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Extract path parameters
        defName := r.URL.Query().Get("definition")
        if defName == "" {
            http.Error(w, "definition parameter required", http.StatusBadRequest)
            return
        }
        
        // Extract resource parameters from query string
        params := make(map[resource.ParameterName]string)
        for key, values := range r.URL.Query() {
            if key != "definition" && len(values) > 0 {
                params[resource.ParameterName(key)] = values[0]
            }
        }
        
        // Create resolution options
        opts := resource.ResolveOptions{}.WithValues(params)
        
        // Add signed URL support for authenticated requests
        if r.Header.Get("Authorization") != "" {
            opts = opts.WithURLOptions(&provider.URLOptions{
                SignedURL:    true,
                SignedExpiry: time.Hour,
            })
        }
        
        // Add context information
        ctx := r.Context()
        if userAgent := r.Header.Get("User-Agent"); userAgent != "" {
            // Parse and add user agent context
            // ctx = useragent.WithUserAgent(ctx, parseUserAgent(userAgent))
        }
        
        // Resolve resource
        result, err := manager.DefinitionResolver().Resolve(ctx, defName, &opts)
        if err != nil {
            var anerr *anerror.Error
            if errors.As(err, &anerr) {
                switch anerr.Type() {
                case "notFound":
                    http.Error(w, "Resource not found", http.StatusNotFound)
                case "invalidArgument":
                    http.Error(w, "Invalid parameters", http.StatusBadRequest)
                default:
                    http.Error(w, "Internal error", http.StatusInternalServerError)
                }
            } else {
                http.Error(w, "Internal error", http.StatusInternalServerError)
            }
            return
        }
        
        // Return JSON response
        response := map[string]interface{}{
            "url":      result.URL,
            "path":     result.ResolvedPath,
            "provider": result.Provider,
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    }
}
```

### GraphQL Integration

```go
type ResourceResolver struct {
    manager resource.ResourceManager
}

func (r *ResourceResolver) ResolveResource(ctx context.Context, args struct {
    Definition string
    Parameters map[string]string
    Provider   *string
    Signed     *bool
    Expiry     *int
}) (*ResolvedResourceType, error) {
    // Convert parameters
    params := make(map[resource.ParameterName]string)
    for k, v := range args.Parameters {
        params[resource.ParameterName(k)] = v
    }
    
    // Create options
    opts := resource.ResolveOptions{}.WithValues(params)
    
    if args.Provider != nil {
        opts = opts.WithProvider(provider.ProviderName(*args.Provider))
    }
    
    if args.Signed != nil && *args.Signed {
        expiry := time.Hour // default
        if args.Expiry != nil {
            expiry = time.Duration(*args.Expiry) * time.Second
        }
        opts = opts.WithURLOptions(&provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: expiry,
        })
    }
    
    // Resolve
    result, err := r.manager.DefinitionResolver().Resolve(ctx, args.Definition, &opts)
    if err != nil {
        return nil, err
    }
    
    return &ResolvedResourceType{
        URL:      result.URL,
        Path:     result.ResolvedPath,
        Provider: string(result.Provider),
    }, nil
}

type ResolvedResourceType struct {
    URL      string
    Path     string
    Provider string
}
```

## Testing

### Test Structure

- **Unit tests**: `*_test.go` files for individual component testing
- **Integration tests**: `*_integration_test.go` test complete workflows
- **Test helpers**: Located in `tests/path_definitions.go`
- **Mock Support**: Test helpers for mocking interfaces

### Unit Testing

```go
func TestResourceResolution(t *testing.T) {
    // Setup test manager
    manager := setupTestResourceManager()
    
    tests := []struct {
        name        string
        definition  string
        params      map[resource.ParameterName]string
        expectError bool
        expectURL   string
    }{
        {
            name:       "Basic resolution",
            definition: "user-avatar",
            params: map[resource.ParameterName]string{
                resource.ResourceID: "test123",
            },
            expectError: false,
            expectURL:   "https://cdn.example.com/avatars/global/test123.jpg",
        },
        {
            name:        "Missing required parameter",
            definition:  "user-avatar",
            params:      map[resource.ParameterName]string{},
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            opts := resource.ResolveOptions{}.WithValues(tt.params)
            
            result, err := manager.DefinitionResolver().Resolve(ctx, tt.definition, &opts)
            
            if tt.expectError {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expectURL, result.URL)
        })
    }
}
```

### Integration Testing

```go
func TestResourceIntegration(t *testing.T) {
    // Setup with real providers (or test containers)
    manager := setupIntegrationManager()
    
    t.Run("CDN Resolution", func(t *testing.T) {
        ctx := context.Background()
        opts := resource.ResolveOptions{}.
            WithProvider(provider.ProviderCDN).
            WithValues(map[resource.ParameterName]string{
                resource.ResourceID: "integration_test_image",
            })
        
        result, err := manager.DefinitionResolver().Resolve(ctx, "test-image", &opts)
        require.NoError(t, err)
        
        // Verify URL is accessible
        resp, err := http.Get(result.URL)
        require.NoError(t, err)
        defer resp.Body.Close()
        
        assert.Equal(t, http.StatusOK, resp.StatusCode)
    })
}
```

### Performance Testing

```go
func BenchmarkResourceResolution(b *testing.B) {
    manager := setupBenchmarkManager()
    ctx := context.Background()
    
    opts := resource.ResolveOptions{}.WithValues(map[resource.ParameterName]string{
        resource.ResourceID: "benchmark_resource",
    })
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := manager.DefinitionResolver().Resolve(ctx, "benchmark-resource", &opts)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Key Test Coverage

1. **Path Definition Validation**: Ensures definitions are properly structured
2. **Parameter Resolution**: Tests all resolver types and chains
3. **Template Processing**: Validates template syntax and parameter substitution
4. **Provider Integration**: Tests URL generation and storage operations
5. **Scope Selection**: Verifies context-based scope resolution
6. **Error Scenarios**: Tests error handling for various failure cases

### Running Tests

```bash
# Run all resource package tests
go test -v ./resource/...

# Run specific test files
go test -v ./resource -run TestPathDefinition
go test -v ./resource -run TestParameterResolvers
go test -v ./resource -run TestTemplateResolver

# Run integration tests
go test -v -integration ./resource/...
```

## Best Practices

### 1. Parameter Validation

- Always validate required parameters before resolution
- Use default values for optional parameters
- Implement proper parameter name validation using `ParameterName.Validate()`

### 2. Scope Design

- Order allowed scopes by priority (most specific first)
- Ensure patterns exist for all allowed scopes
- Use hierarchical fallback (ClientApp → App → Global)

### 3. Template Security

- Never trust user input in templates without validation
- Use parameterized templates instead of string concatenation
- Leverage built-in safety patterns and validation

### 4. Provider Configuration

- Configure appropriate URL types for each provider pattern
- Use signed URLs for sensitive or temporary resources
- Set reasonable expiry times for signed URLs

### 5. Error Handling

- Check for specific error types using `errors.Is()`
- Provide meaningful error context for debugging
- Handle provider-specific errors gracefully

### 6. Performance

- Use parameter resolver caching for repeated lookups
- Implement connection pooling for provider backends
- Cache frequently accessed path definitions

### 7. Testing

- Test all scope combinations for path definitions
- Validate template resolution with edge cases
- Mock provider interfaces for unit testing

## Performance Optimization

### Connection Pooling

```go
type OptimizedResourceManager struct {
    *resource.ResourceManagerImpl
    providerPool map[provider.ProviderName]*sync.Pool
}

func NewOptimizedResourceManager(baseManager resource.ResourceManager) *OptimizedResourceManager {
    return &OptimizedResourceManager{
        ResourceManagerImpl: baseManager.(*resource.ResourceManagerImpl),
        providerPool:        make(map[provider.ProviderName]*sync.Pool),
    }
}

func (m *OptimizedResourceManager) getPooledProvider(name provider.ProviderName) provider.URLProvider {
    pool, exists := m.providerPool[name]
    if !exists {
        pool = &sync.Pool{
            New: func() interface{} {
                // Create new provider instance
                return m.createProvider(name)
            },
        }
        m.providerPool[name] = pool
    }
    
    return pool.Get().(provider.URLProvider)
}
```

### Memory Optimization

```go
// String interning for parameter names
var parameterNameIntern = make(map[string]resource.ParameterName)
var parameterNameMutex sync.RWMutex

func InternParameterName(name string) resource.ParameterName {
    parameterNameMutex.RLock()
    if interned, exists := parameterNameIntern[name]; exists {
        parameterNameMutex.RUnlock()
        return interned
    }
    parameterNameMutex.RUnlock()
    
    parameterNameMutex.Lock()
    defer parameterNameMutex.Unlock()
    
    // Double-check after acquiring write lock
    if interned, exists := parameterNameIntern[name]; exists {
        return interned
    }
    
    paramName := resource.ParameterName(name)
    parameterNameIntern[name] = paramName
    return paramName
}
```

### Performance Considerations

- **Template Caching**: Compiled templates are cached with configurable intervals
- **Registry Optimization**: Path definitions are cached for O(1) lookup by name
- **Concurrent Safety**: All components are thread-safe with read-write mutexes
- **Provider Selection**: Cached provider lookups by URL type for O(1) access
- **Parameter Resolution**: Chained resolvers with short-circuiting for efficiency

---

## Related Components

- **Provider Package** (`resource/provider`): Storage and URL generation backends
- **Common/ANError**: Structured error handling
- **Common/Context**: Request context management
- **Common/App**: Application context extraction
- **Common/UserAgent**: Client application identification

---

This comprehensive documentation combines architectural overview, API reference, detailed usage examples, and practical implementation guidance to provide complete coverage of the Resource package's capabilities for the Aviron Game Server ecosystem.