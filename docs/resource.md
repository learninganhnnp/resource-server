# Resource Package Documentation

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Usage Guide](#usage-guide)
- [Core Components](#core-components)
- [Parameter System](#parameter-system)
- [Template Resolution](#template-resolution)
- [Provider System](#provider-system)
- [Scope Management](#scope-management)

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

## Architecture

The resource system follows a layered architecture:

```
ResourceManager
├── PathDefinitionResolver
│   ├── PathResolver
│   │   └── TemplateResolver
│   ├── ScopeSelector
│   └── PathDefinitionRegistry
└── PathURLResolver
    └── ProviderRegistry
```

### Resource Resolution Flow

1. **Path Definition**: Contains patterns and metadata for resource paths
2. **Scope Selection**: Determines the appropriate scope (Global, App, ClientApp) based on context
3. **Parameter Resolution**: Resolves template parameters from various sources (context, values, fallbacks)
4. **Provider Selection**: Chooses storage provider based on URL type and preferences
5. **Template Processing**: Applies parameters to pattern templates with security validation
6. **URL Generation**: Creates final URLs with optional signing and expiry


## Usage Guide

### Table of Contents

- [Quick Start](#quick-start) - Get up and running in minutes
- [Basic Operations](#basic-operations) - Common use cases
- [Advanced Features](#advanced-features) - Provider selection, signed URLs, scopes
- [Performance Patterns](#performance-patterns) - Best practices for production
- [Production Examples](#production-examples) - Real-world service implementations

---

## Quick Start

### 1. Setup ResourceManager

```go
func setupResourceManager() ResourceManager {
    // Create components
    templateResolver, _ := NewBracesTemplateResolver()
    scopeSelector := NewDefaultScopeSelector()
    pathResolver := NewPathResolver(templateResolver, nil)
    definitionRegistry := NewPathDefinitionRegistry()
    providerRegistry := provider.NewRegistry()

    // Register CDN provider
    cdnProvider := provider.NewCDNProvider(provider.CDNConfig{
        BaseURL:    "https://cdn.example.com",
        SigningKey: "your-signing-key",
        Expiry:     1 * time.Hour,
    })
    providerRegistry.Register(provider.ProviderCDN, cdnProvider)

    // Register GCS provider
    gcsProvider, _ := provider.NewGCSProvider(ctx, provider.GCSConfig{
        BucketName: "your-bucket",
        Expiry:     1 * time.Hour,
    })
    providerRegistry.Register(provider.ProviderGCS, gcsProvider)

    return NewResourceManager(pathResolver, scopeSelector, definitionRegistry, providerRegistry)
}
```

### 2. Define Resource Patterns

```go
manager.AddDefinition(PathDefinition{
    Name:          "user-avatars",
    AllowedScopes: []ScopeType{ScopeGlobal},
    Patterns: map[provider.ProviderName]PathPatterns{
        provider.ProviderCDN: {
            Patterns: map[ScopeType]string{
                ScopeGlobal: "users/avatars/{user_id}.png",
            },
            URLType: URLTypeContent,
        },
        provider.ProviderGCS: {
            Patterns: map[ScopeType]string{
                ScopeGlobal: "avatars/{user_id}.png",
            },
            URLType: URLTypeOperation,
        },
    },
    Parameters: []*ParameterDefinition{
        {Name: "user_id", Required: true},
    },
})
```

### 3. Resolve URLs

```go
// Simple URL resolution
result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", 
    &ResolveOptions{
        ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
            "user_id": "12345",
        }),
    })

fmt.Printf("URL: %s\n", result.URL)
// Output: https://cdn.example.com/users/avatars/12345.png
```

---

## Basic Operations

### Simple Resource Resolution

```go
// Resolve with default provider (CDN for content URLs)
result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", 
    &ResolveOptions{
        ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
            "user_id": "12345",
        }),
    })

fmt.Printf("Avatar URL: %s\n", result.URL)
// Output: https://cdn.example.com/users/avatars/12345.png
```

### Using Fluent API

```go
// Cleaner syntax with fluent API
result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars",
    ResolveOptions{}.WithValues(map[ParameterName]string{
        "user_id": "12345",
    }))
```

### Parameter Resolution Patterns

```go
// Static parameters
staticResolver := NewValuesParameterResolver(map[ParameterName]string{
    "user_id": "12345",
    "format":  "png",
})

// Context-based parameters
contextResolver := NewClientAppParameterResolver(clientApps)

// Chained resolvers (tries each in order)
chainedResolver, _ := NewParameterResolvers(
    staticResolver,     // First: check static values
    contextResolver,    // Then: check context
)
```

---

## Advanced Features

### Provider Selection

#### Explicit Provider Choice

```go
// Force GCS instead of default CDN
gcsProvider := provider.ProviderGCS
result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", 
    &ResolveOptions{
        Provider: &gcsProvider,
        ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
            "user_id": "12345",
        }),
    })

fmt.Printf("GCS URL: %s\n", result.URL)
// Output: https://storage.googleapis.com/avatars/12345.png
```

#### URL Type-Based Selection

```go
// Content URLs (optimized for delivery) → CDN
contentType := URLTypeContent
contentResult, err := manager.PathDefinitionResolver().Resolve(ctx, "resources", 
    &ResolveOptions{
        URLType: &contentType,
        ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
            "resource_id": "image123",
        }),
    })

// Operation URLs (for management) → GCS
operationType := URLTypeOperation
operationResult, err := manager.PathDefinitionResolver().Resolve(ctx, "resources", 
    &ResolveOptions{
        URLType: &operationType,
        ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
            "resource_id": "image123",
        }),
    })
```

### Signed URL Generation

#### Basic Signed URLs

```go
// Generate signed URL with default expiry
result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", 
    &ResolveOptions{
        URLOptions: &provider.URLOptions{
            SignedURL: true,
        },
        ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
            "user_id": "secure_user",
        }),
    })
```

#### Custom Expiry

```go
// 2-hour signed URL
result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", 
    &ResolveOptions{
        URLOptions: &provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: 2 * time.Hour,
        },
        ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
            "user_id": "temp_user",
        }),
    })
```

#### Security-Limited Expiry

```go
// PathDefinition with maximum expiry enforcement
definition := PathDefinition{
    Name: "secure-docs",
    URLOptions: &provider.URLOptions{
        SignedExpiry: 1 * time.Hour, // Maximum allowed expiry
    },
    // ... other fields
}

// Request for 3 hours gets capped to 1 hour
result, err := manager.PathDefinitionResolver().Resolve(ctx, "secure-docs", 
    &ResolveOptions{
        URLOptions: &provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: 3 * time.Hour, // Requested but will be capped
        },
    })
```

### Scope Management

#### Automatic Scope Selection

```go
// Context determines scope automatically
ctx := context.Background()
ctx = app.WithApp(ctx, &app.App{ID: 1, Name: "bike"})
ctx = useragent.WithUserAgent(ctx, &useragent.UserAgent{
    ClientAppID: 2,
    Version:     "1.5.0",
})

// System selects most specific available scope
result, err := manager.PathDefinitionResolver().Resolve(ctx, "app-resources", 
    &ResolveOptions{
        ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
            "resource_id": "config123",
        }),
    })
```

#### Explicit Scope Override

```go
// Force global scope
globalScope := ScopeGlobal
result, err := manager.PathDefinitionResolver().Resolve(ctx, "shared-resources", 
    &ResolveOptions{
        Scope: &globalScope,
        ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
            "resource_id": "global_logo",
        }),
    })
```

### Complex Combinations

```go
// Combine all options: GCS + operation URL + signed URL + app scope
result, err := manager.PathDefinitionResolver().Resolve(ctx, "admin-resources",
    ResolveOptions{}.
        WithProvider(provider.ProviderGCS).
        WithURLType(URLTypeOperation).
        WithURLOptions(&provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: 30 * time.Minute,
        }).
        WithValues(map[ParameterName]string{
            "resource_id": "admin_backup",
        }))
```

---

## Performance Patterns

### Reuse Options Structures

```go
// ❌ Bad: Creating new options every time
func getBadUserAvatar(userID string) string {
    result, _ := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", 
        &ResolveOptions{
            URLOptions: &provider.URLOptions{SignedURL: true, SignedExpiry: 1 * time.Hour},
            ParamResolver: NewValuesParameterResolver(map[ParameterName]string{"user_id": userID}),
        })
    return result.URL
}

// ✅ Good: Reuse base options
var baseSecureOptions = &ResolveOptions{
    URLOptions: &provider.URLOptions{
        SignedURL:    true,
        SignedExpiry: 1 * time.Hour,
    },
}

func getGoodUserAvatar(userID string) string {
    opts := *baseSecureOptions // Copy struct
    opts.ParamResolver = NewValuesParameterResolver(map[ParameterName]string{"user_id": userID})
    
    result, _ := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &opts)
    return result.URL
}
```

### Pre-built Resolver Chains

```go
// Create resolver chains once, reuse many times
var userContextResolver, _ = NewParameterResolvers(
    NewClientAppParameterResolver(clientApps),
    NewAppParameterResolver(apps),
    NewValuesParameterResolver(map[ParameterName]string{
        "format": "png", // fallback format
    }),
)

func resolveUserResource(resourceName string, userParams map[ParameterName]string) (*ResolvedResource, error) {
    userResolver := NewValuesParameterResolver(userParams)
    combinedResolver, _ := NewParameterResolvers(userResolver, userContextResolver)
    
    return manager.PathDefinitionResolver().Resolve(ctx, resourceName, &ResolveOptions{
        ParamResolver: combinedResolver,
    })
}
```

---

## Production Examples

### Service-Specific Configurations

#### User Service

```go
type UserService struct {
    resourceManager ResourceManager
    avatarOptions   *ResolveOptions
    profileOptions  *ResolveOptions
}

func NewUserService(manager ResourceManager) *UserService {
    return &UserService{
        resourceManager: manager,
        // Long-lived avatars with signing
        avatarOptions: &ResolveOptions{
            URLOptions: &provider.URLOptions{
                SignedURL:    true,
                SignedExpiry: 24 * time.Hour,
            },
        },
        // Public profile images, no signing needed
        profileOptions: &ResolveOptions{
            URLType: &[]URLType{URLTypeContent}[0],
        },
    }
}

func (s *UserService) GetUserAvatarURL(userID string) (string, error) {
    opts := *s.avatarOptions
    opts.ParamResolver = NewValuesParameterResolver(map[ParameterName]string{
        "user_id": userID,
    })
    
    result, err := s.resourceManager.PathDefinitionResolver().Resolve(context.Background(), "user-avatars", &opts)
    if err != nil {
        return "", err
    }
    return result.URL, nil
}
```

#### Admin Service

```go
type AdminService struct {
    resourceManager ResourceManager
    adminOptions    *ResolveOptions
}

func NewAdminService(manager ResourceManager) *AdminService {
    gcsProvider := provider.ProviderGCS
    operationType := URLTypeOperation
    
    return &AdminService{
        resourceManager: manager,
        // Short-lived, GCS operation URLs for admin tasks
        adminOptions: &ResolveOptions{
            Provider: &gcsProvider,
            URLType:  &operationType,
            URLOptions: &provider.URLOptions{
                SignedURL:    true,
                SignedExpiry: 15 * time.Minute, // Short expiry for security
            },
        },
    }
}

func (s *AdminService) GetManagementURL(resourceType, resourceID string) (string, error) {
    opts := *s.adminOptions
    opts.ParamResolver = NewValuesParameterResolver(map[ParameterName]string{
        "resource_id": resourceID,
    })
    
    result, err := s.resourceManager.PathDefinitionResolver().Resolve(context.Background(), resourceType, &opts)
    if err != nil {
        return "", err
    }
    return result.URL, nil
}
```

#### Content Delivery Service

```go
type CDNService struct {
    resourceManager ResourceManager
    publicOptions   *ResolveOptions
    secureOptions   *ResolveOptions
}

func NewCDNService(manager ResourceManager) *CDNService {
    contentType := URLTypeContent
    
    return &CDNService{
        resourceManager: manager,
        // Public content, no signing
        publicOptions: &ResolveOptions{
            URLType: &contentType,
        },
        // Secure content with time-limited access
        secureOptions: &ResolveOptions{
            URLType: &contentType,
            URLOptions: &provider.URLOptions{
                SignedURL:    true,
                SignedExpiry: 2 * time.Hour,
            },
        },
    }
}

func (s *CDNService) GetPublicContentURL(resourceID string) (string, error) {
    opts := *s.publicOptions
    opts.ParamResolver = NewValuesParameterResolver(map[ParameterName]string{
        "resource_id": resourceID,
    })
    
    result, err := s.resourceManager.PathDefinitionResolver().Resolve(context.Background(), "public-content", &opts)
    if err != nil {
        return "", err
    }
    return result.URL, nil
}

func (s *CDNService) GetSecureContentURL(resourceID string) (string, error) {
    opts := *s.secureOptions
    opts.ParamResolver = NewValuesParameterResolver(map[ParameterName]string{
        "resource_id": resourceID,
    })
    
    result, err := s.resourceManager.PathDefinitionResolver().Resolve(context.Background(), "secure-content", &opts)
    if err != nil {
        return "", err
    }
    return result.URL, nil
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

type ContentService struct {
    resourceManager ResourceManager
}

func (s *ContentService) GetContentURL(resourceID, contentType string) (string, error) {
    opts := &ResolveOptions{
        URLOptions: &provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: getExpiryForContentType(contentType),
        },
        ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
            "resource_id":  resourceID,
            "content_type": contentType,
        }),
    }
    
    result, err := s.resourceManager.PathDefinitionResolver().Resolve(context.Background(), "dynamic-content", opts)
    if err != nil {
        return "", err
    }
    return result.URL, nil
}
```

---

## Quick Reference

### Common Patterns Cheat Sheet

```go
// 1. Simple URL (default provider)
result, _ := manager.PathDefinitionResolver().Resolve(ctx, "resource-name",
    ResolveOptions{}.WithValues(map[ParameterName]string{
        "param": "value",
    }))

// 2. Signed URL (1 hour expiry)
result, _ := manager.PathDefinitionResolver().Resolve(ctx, "resource-name",
    ResolveOptions{}.
        WithValues(params).
        WithURLOptions(&provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: time.Hour,
        }))

// 3. Force specific provider
result, _ := manager.PathDefinitionResolver().Resolve(ctx, "resource-name",
    ResolveOptions{}.
        WithValues(params).
        WithProvider(provider.ProviderGCS))

// 4. Content vs Operation URLs
contentResult, _ := manager.PathDefinitionResolver().Resolve(ctx, "resource-name",
    ResolveOptions{}.WithValues(params).WithURLType(URLTypeContent))   // → CDN

operationResult, _ := manager.PathDefinitionResolver().Resolve(ctx, "resource-name",
    ResolveOptions{}.WithValues(params).WithURLType(URLTypeOperation)) // → GCS

// 5. Override scope
result, _ := manager.PathDefinitionResolver().Resolve(ctx, "resource-name",
    ResolveOptions{}.
        WithValues(params).
        WithScope(ScopeGlobal))

// 6. Complex combination
result, _ := manager.PathDefinitionResolver().Resolve(ctx, "resource-name",
    ResolveOptions{}.
        WithProvider(provider.ProviderGCS).
        WithURLType(URLTypeOperation).
        WithURLOptions(&provider.URLOptions{SignedURL: true, SignedExpiry: 30*time.Minute}).
        WithScope(ScopeApp).
        WithValues(params))
```

### Parameter Resolver Patterns

```go
// Static values
resolver := NewValuesParameterResolver(map[ParameterName]string{
    "user_id": "12345",
    "format":  "png",
})

// Context-based
resolver := NewClientAppParameterResolver(clientApps)

// Chained (tries in order)
resolver, _ := NewParameterResolvers(
    NewValuesParameterResolver(userParams),    // 1st: user provided
    NewClientAppParameterResolver(clientApps), // 2nd: from context
    NewValuesParameterResolver(defaults),      // 3rd: fallback defaults
)

// Routing (parameter-specific)
resolver := NewRoutingParameterResolver(map[ParameterName]ParameterResolver{
    "client_app": NewClientAppParameterResolver(clientApps),
    "app":        NewAppParameterResolver(apps),
    "timestamp":  NewCurrentTimestampParameterResolver("2006-01-02"),
})
```

### Service Configuration Templates

```go
// Public content service
type PublicContentService struct {
    options *ResolveOptions
}
func NewPublicContentService(manager ResourceManager) *PublicContentService {
    return &PublicContentService{
        options: &ResolveOptions{URLType: &[]URLType{URLTypeContent}[0]},
    }
}

// Secure content service  
type SecureContentService struct {
    options *ResolveOptions
}
func NewSecureContentService(manager ResourceManager) *SecureContentService {
    return &SecureContentService{
        options: &ResolveOptions{
            URLOptions: &provider.URLOptions{SignedURL: true, SignedExpiry: 2*time.Hour},
        },
    }
}

// Admin operations service
type AdminService struct {
    options *ResolveOptions
}
func NewAdminService(manager ResourceManager) *AdminService {
    gcs := provider.ProviderGCS
    op := URLTypeOperation
    return &AdminService{
        options: &ResolveOptions{
            Provider: &gcs,
            URLType:  &op,
            URLOptions: &provider.URLOptions{SignedURL: true, SignedExpiry: 15*time.Minute},
        },
    }
}
```

### URL Types Decision Matrix

| Use Case | URL Type | Provider | Signed | Expiry |
|----------|----------|----------|--------|--------|
| Public images/assets | `URLTypeContent` | CDN | No | - |
| User profile pics | `URLTypeContent` | CDN | Yes | 24h |
| Temp file downloads | `URLTypeContent` | CDN | Yes | 15m |
| Admin file uploads | `URLTypeOperation` | GCS | Yes | 15m |
| Backup operations | `URLTypeOperation` | GCS | Yes | 1h |
| Bulk data exports | `URLTypeOperation` | GCS | Yes | 4h |

---

## Core Components

### ResourceManager

Central interface for managing resource definitions and resolution.

```go
type ResourceManager interface {
    PathURLResolver() PathURLResolver
    PathDefinitionResolver() PathDefinitionResolver

    GetDefinition(name PathDefinitionName) (*PathDefinition, error)
    AddDefinition(definition PathDefinition) error
    RemoveDefinition(defName PathDefinitionName) error
    GetAllDefinitions() []*PathDefinition
}
```

**Constructor**:

```go
func NewResourceManager(
    pathResolver PathResolver,
    scopeSelector ScopeSelector,
    definitionRegistry PathDefinitionRegistry,
    providerRegistry provider.Registry,
) ResourceManager
```

### PathDefinition

Defines resource path patterns with provider and scope support.

```go
type PathDefinition struct {
    Name          PathDefinitionName                      // Unique identifier
    DisplayName   string                                  // Human-readable name
    Description   string                                  // Description of the resource
    AllowedScopes []ScopeType                            // Permitted scopes (ordered by priority)
    Patterns      map[provider.ProviderName]PathPatterns // Provider-specific patterns
    Parameters    []*ParameterDefinition                 // Parameter definitions
    URLOptions    *provider.URLOptions                   // URL options with security limits. If set, definition's expiry is the maximum allowed time.
}

type PathPatterns struct {
    Patterns map[ScopeType]string  // Scope-specific path patterns
    URLType  URLType               // Content or Operation URLs
}

type URLType string
const (
    URLTypeContent   URLType = "content"   // Content delivery URLs
    URLTypeOperation URLType = "operation" // Storage operation URLs like uploads, reads, deletes
)
```

### ResolveOptions

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

Result of resource resolution.

```go
type ResolvedResource struct {
    URL          string                 // Generated URL
    ResolvedPath string                 // Resolved path template
    Provider     provider.ProviderName  // Selected provider
}
```

## Parameter System

### Parameter Names

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

```go
type ParameterDefinition struct {
    Name         ParameterName // Parameter name
    DefaultValue string        // Default value if not provided
    Description  string        // Parameter description
    Required     bool          // Whether parameter is required
}
```

### Parameter Resolvers

The parameter resolution system supports multiple resolver types with fallback chains:

#### Context-Based Resolvers

- **ClientAppParameterResolver**: Resolves based on user agent context
- **AppParameterResolver**: Resolves based on application context

#### Value-Based Resolvers

- **ValuesParameterResolver**: Static key-value parameter mapping
- **ValueParameterResolver**: Single parameter resolver
- **TimestampParameterResolver**: Time-based parameter generation

#### Composite Resolvers

Advanced parameter resolution patterns for complex scenarios:

##### NewParameterResolvers - Resolver Chaining

Creates a chain of parameter resolvers that are called in sequence until a parameter is resolved:

```go
func NewParameterResolvers(resolvers ...ParameterResolver) (ParameterResolver, error)
```

**Features:**
- Resolvers are tried in order until one returns a non-empty value
- Early termination on first successful resolution
- Automatic fallback through the resolver chain
- Thread-safe operation

**Usage:**
```go
// Create a resolver chain with priority order
chainedResolver, err := resource.NewParameterResolvers(
    resource.NewValuesParameterResolver(map[resolver.ParameterName]string{
        resource.ResourceID: "override_id",
    }),
    resource.NewClientAppParameterResolver(clientApps),
    resource.NewAppParameterResolver(apps),
    resource.NewValuesParameterResolver(map[resolver.ParameterName]string{
        resource.ResourceFormat: "jpg", // fallback format
    }),
)
```

##### NewRoutingParameterResolver - Parameter-Specific Routing

Routes specific parameters to designated resolvers for specialized handling:

```go
func NewRoutingParameterResolver(resolvers map[ParameterName]ParameterResolver) *RoutingParameterResolver
```

**Features:**
- Parameter-specific resolver assignment
- Clean separation of concerns

**Usage:**
```go
routingResolver := resource.NewRoutingParameterResolver(map[resolver.ParameterName]resolver.ParameterResolver{
    resource.ClientApp:      resource.NewClientAppParameterResolver(clientApps),
    resource.App:            resource.NewAppParameterResolver(apps),
    resource.ResourceID:     resource.NewIDParameterResolver("default_id"),
    resource.Timestamp:      resource.NewCurrentTimestampParameterResolver("2006-01-02"), // custom format
})
```

##### NewDefinitionParameterResolver - Definition-Based Resolution

Wraps a resolver with parameter definition support, adding defaults and validation:

```go
func NewDefinitionParameterResolver(definition *ParameterDefinition, resolver ParameterResolver) (ParameterResolver, error)
```

**Features:**
- Automatic default value application
- Required parameter validation
- Definition-driven behavior

**Usage:**
```go
// Create parameter definition with validation
userIdDef := &resolver.ParameterDefinition{
    Name:         "user_id",
    Required:     true,
    DefaultValue: "",
    Description:  "User identifier for personalized resources",
}

// Wrap base resolver with definition behavior
defResolver, err := resource.NewDefinitionParameterResolver(
    userIdDef,
    resource.NewValuesParameterResolver(params),
)
```

##### NewValidatingParameterResolver - Validation Layer

Adds validation to any parameter resolver using custom validator functions:

```go
func NewValidatingParameterResolver(resolver ParameterResolver, validators ...ParameterValidator) (ParameterResolver, error)

type ParameterValidator func(ctx context.Context, paramName ParameterName, value string) error
```

**Built-in Validators:**
- `WithEmptyParameterValidation` - Ensures parameters are not empty

**Features:**
- Multiple validator chaining
- Custom validation logic support
- Validation error aggregation

**Usage:**
```go
// Custom validator for resource ID format
resourceIdValidator := func(ctx context.Context, paramName resolver.ParameterName, value string) error {
    if paramName == resource.ResourceID && !regexp.MustMatch(`^[a-zA-Z0-9_-]+$`, value) {
        return errors.New("resource ID must contain only alphanumeric characters, underscores, and hyphens")
    }
    return nil
}

// Apply validation to resolver
validatedResolver, err := resource.NewValidatingParameterResolver(
    baseResolver,
    resource.WithEmptyParameterValidation,
    resourceIdValidator,
)
```

##### Advanced Chaining Example

Combining multiple resolver patterns for complex parameter resolution:

```go
func createAdvancedParameterResolver(
    clientApps map[int16]string,
    apps map[int16]string,
    userParams map[resolver.ParameterName]string,
) (resolver.ParameterResolver, error) {
    
    // Create base resolvers
    contextResolver := resource.NewRoutingParameterResolver(map[resolver.ParameterName]resolver.ParameterResolver{
        resource.ClientApp: resource.WithClientAppResolver(clientApps),
        resource.App:       resource.WithAppResolver(apps),
    })
    
    userResolver := resource.NewValuesParameterResolver(userParams)
    
    // Chain resolvers with priority: user params > context > defaults
    chainedResolver, err := resource.NewParameterResolvers(
        userResolver,
        contextResolver,
        resource.DefaultFallbackParameterResolver(clientApps, apps),
    )
    if err != nil {
        return nil, err
    }
    
    // Add validation layer
    return resource.NewValidatingParameterResolver(
        chainedResolver,
        resource.WithEmptyParameterValidation,
    )
}

#### ParameterResolverFunc - Functional Interface

Function type that implements the `ParameterResolver` interface for simple resolver creation:

```go
type ParameterResolverFunc func(ctx context.Context, paramName ParameterName) (string, error)
```

**Usage:**
```go
// Create a simple resolver using a function
timestampResolver := resolver.ParameterResolverFunc(func(ctx context.Context, paramName resolver.ParameterName) (string, error) {
    if paramName == resource.Timestamp {
        return time.Now().Format(time.RFC3339), nil
    }
    return "", nil // Let other resolvers handle it
})

// Use in resolver chains
chainedResolver, err := resource.NewParameterResolvers(
    timestampResolver,
    otherResolvers...,
)

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

func (r *UserParameterResolver) Resolve(ctx context.Context, paramName resolver.ParameterName) (string, error) {
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
    params := map[resolver.ParameterName]string{
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

Handles CDN URL generation with optional signed URL support:

```go
type CDNConfig struct {
	BaseURL    string        `yaml:"base_url"`    // Base URL for CDN paths
	SigningKey string        `yaml:"signing_key"` // Key used for signing URLs
	Expiry     time.Duration `yaml:"expiry"`      // Default expiry for signed URLs if not specified
}
```

**Features**:

- Public URL generation
- HMAC-SHA256 signed URLs with expiration
- Query parameter-based signatures

### Google Cloud Storage Provider

Full GCS integration with storage operations and signed URLs:

```go
type GCSConfig struct {
	BucketName         string        `yaml:"bucket_name"`          // Default bucket name to use if not specified in paths
	CredentialsPath    string        `yaml:"credentials_path"`     // Path to GCS credentials file
	Expiry             time.Duration `yaml:"expiry"`               // Default expiry for signed URLs if not specified
	ServiceAccountJSON string        `yaml:"service_account_json"` // Alternative to credentials file
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
cdnProvider := provider.NewCDNProvider(provider.CDNConfig{
    BaseURL:    "https://cdn1.avironactive.com",
    SigningKey: "secret-key",
    Expiry:     1 * time.Hour,
})
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
        WithValues(map[resolver.ParameterName]string{
            resource.ResourceID: "backup_20241201",
            resource.Version:    "1.0.0",
        })

    result, err := manager.PathDefinitionResolver().Resolve(ctx, "backup-file", &opts)
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

Hierarchical scope system:

```go
const (
    ScopeGlobal    ScopeType = "G"   // Global resources (no context required)
    ScopeApp       ScopeType = "A"   // App-scoped resources (requires app context)
    ScopeClientApp ScopeType = "CA"  // Client app-scoped resources (requires client app context)
)
```

### Scope Selection

Automatic scope selection based on context:

```go
type ScopeSelector interface {
    Select(ctx context.Context, definition *PathDefinition) (ScopeType, error)
    IsAllowed(ctx context.Context, scope ScopeType, definition *PathDefinition) (bool, error)
}
```

### Context Requirements

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
    opts := resource.ResolveOptions{}.WithValues(map[resolver.ParameterName]string{
        resource.ResourceID: "feature_config_123",
    })

    result, err := manager.PathDefinitionResolver().Resolve(ctx, "app-config", &opts)
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
    globalScope := resolver.ScopeGlobal
    globalOpts := resource.ResolveOptions{
        Scope: &globalScope,
    }.WithValues(map[resolver.ParameterName]string{
        resource.ResourceID: "shared_achievement_badge",
    })

    globalResult, err := manager.PathDefinitionResolver().Resolve(ctx, "achievement-badge", &globalOpts)
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
    }.WithValues(map[resolver.ParameterName]string{
        resource.ResourceID: "secure_document_789",
    })

    result, err := manager.PathDefinitionResolver().Resolve(ctx, "secure-document", &opts)
    if err != nil {
        log.Printf("Failed to generate signed URL: %v", err)
        return
    }

    fmt.Printf("Signed URL: %s\n", result.URL)
    fmt.Printf("Expires in: %v\n", time.Hour)
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
    }.WithValues(map[resolver.ParameterName]string{
        resource.ResourceID: "content_456",
        "content_type":      contentType,
    })

    result, err := manager.PathDefinitionResolver().Resolve(ctx, "dynamic-content", &opts)
    if err != nil {
        log.Printf("Failed to resolve: %v", err)
        return
    }

    fmt.Printf("Content URL: %s\n", result.URL)
}
```
