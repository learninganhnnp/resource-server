# Resource Package Usage Guide

A comprehensive guide for using the Resource package in the Aviron game server ecosystem, organized from basic concepts to advanced patterns.

## Table of Contents

1. [Introduction](#introduction)
2. [Core Concepts](#core-concepts)
3. [Getting Started](#getting-started)
4. [Basic Usage](#basic-usage)
5. [Parameter Resolution](#parameter-resolution)
6. [Provider Selection](#provider-selection)
7. [Scope Management](#scope-management)
8. [Signed URLs](#signed-urls)
9. [Advanced Patterns](#advanced-patterns)
10. [Production Best Practices](#production-best-practices)
11. [Troubleshooting](#troubleshooting)
12. [Quick Reference](#quick-reference)

## Introduction

The Resource package provides a unified system for managing and resolving resource paths across different storage providers (CDN, Google Cloud Storage) with support for:

- **Dynamic path resolution** with parameter substitution
- **Multi-provider support** with automatic selection
- **Scope-based access control** (Global, App, ClientApp)
- **Signed URL generation** with configurable expiry
- **Security features** including path traversal protection

### Why Use the Resource Package?

Instead of hardcoding URLs throughout your codebase:
```go
// ❌ Bad: Hardcoded URLs scattered everywhere
avatarURL := "https://cdn.example.com/shared/global/users/avatars/" + userID + ".png"
```

Use the Resource package for centralized, flexible resource management:
```go
// ✅ Good: Centralized resource management
result, _ := manager.Resolve(ctx, "user-avatars", 
    ResolveOptions{}.WithValues(map[ParameterName]string{"user_id": userID}))
avatarURL := result.URL
```

## Core Concepts

### 1. Path Definitions
Path definitions describe how resources are organized across different providers and scopes:

```go
var UserAvatarsPath = PathDefinition{
    Name:          "user-avatars",
    DisplayName:   "User Avatars",
    Description:   "User avatar images shared across all platforms",
    AllowedScopes: []ScopeType{ScopeGlobal},
    Patterns: map[provider.ProviderName]PathPatterns{
        provider.ProviderCDN: {
            Patterns: map[ScopeType]string{
                ScopeGlobal: "shared/global/users/avatars/{user_id}.png",
            },
            URLType: URLTypeContent,
        },
        provider.ProviderGCS: {
            Patterns: map[ScopeType]string{
                ScopeGlobal: "global-avatars/{user_id}.png",
            },
            URLType: URLTypeOperation,
        },
    },
    Parameters: []*ParameterDefinition{
        {Name: "user_id", Required: true, Description: "User identifier"},
    },
}
```

### 2. URL Types
- **URLTypeContent**: For content delivery (typically CDN)
- **URLTypeOperation**: For storage operations (typically GCS)

### 3. Scopes
- **ScopeGlobal**: Resources shared across all applications
- **ScopeApp**: Resources specific to an app (bike, rower)
- **ScopeClientApp**: Resources specific to a client platform (iOS, Android)

### 4. Providers
- **CDN Provider**: Optimized for content delivery
- **GCS Provider**: For storage operations and management

## Getting Started

### Quick Start Example

```go
// Minimal setup using defaults
manager, err := NewResourceManager()
if err != nil {
    log.Fatal(err)
}

// Add providers
cdnProvider := provider.NewCDNProvider(provider.CDNConfig{
    BaseURL: "https://cdn.example.com",
})
manager.URLResolver().(*URLResolver).providerRegistry.Register(cdnProvider)

// Add path definitions
manager.AddDefinition(UserAvatarsPath)

// Start resolving resources
result, err := manager.DefinitionResolver().Resolve(ctx, "user-avatars",
    ResolveOptions{}.WithValues(map[ParameterName]string{
        "user_id": "12345",
    }))
```

### Step 1: Create Component Instances (Advanced Setup)

For more control, you can configure each component separately:

```go
// Create template resolver (handles {parameter} substitution)
templateResolver, err := NewBracesTemplateResolver()
if err != nil {
    log.Fatal(err)
}

// Create scope selector (determines resource access scope)
scopeSelector := NewScopeSelector()

// Setup client app and app mappings for parameter resolution
clientApps := map[int16]string{
    1: "ios",
    2: "android",
}
apps := map[int16]string{
    1: "bike",
    2: "rower",
}

// Create path resolver with default fallback
// The fallback resolver provides client_app and app parameters from context
fallbackResolver := DefaultFallbackParameterResolver(clientApps, apps)
pathResolver := NewPathResolver(templateResolver, fallbackResolver)

// Create registries
definitionRegistry := NewPathDefinitionRegistry()
providerRegistry := provider.NewRegistry()

// Create URLResolver (optional - created automatically if not provided)
URLResolver := NewURLResolver(providerRegistry)
```

### Step 2: Configure Providers

```go
// Configure CDN provider
cdnProvider := provider.NewCDNProvider(provider.CDNConfig{
    BaseURL:    "https://cdn.example.com",
    SigningKey: "your-signing-key",
    Expiry:     1 * time.Hour,
})
err := providerRegistry.Register(cdnProvider)
if err != nil {
    log.Fatal(err)
}

// Configure GCS provider
gcsProvider, _ := provider.NewGCSProvider(ctx, provider.GCSConfig{
    BucketName: "your-bucket",
    Expiry:     1 * time.Hour,
})
err = providerRegistry.Register(gcsProvider)
if err != nil {
    log.Fatal(err)
}
```

### Step 3: Create ResourceManager

```go
// Using options pattern for configuration
manager, err := NewResourceManager(
    WithPathResolver(pathResolver),
    WithScopeSelector(scopeSelector),
    WithDefinitionRegistry(definitionRegistry),
    WithProviderRegistry(providerRegistry),
    WithURLResolver(URLResolver),
)
if err != nil {
    log.Fatal(err)
}

// Or use defaults with minimal configuration
manager, err := NewResourceManager(
    WithProviderRegistry(providerRegistry),
)
if err != nil {
    log.Fatal(err)
}
```

#### Available ResourceManager Options

- `WithPathResolver(resolver)` - Custom path resolver for template substitution
- `WithScopeSelector(selector)` - Custom scope selection logic
- `WithDefinitionRegistry(registry)` - Pre-configured path definitions
- `WithProviderRegistry(registry)` - Storage provider configuration
- `WithTemplateResolver(resolver)` - Custom template resolver (defaults to BracesTemplateResolver)
- `WithFallbackParameterResolver(resolver)` - Default parameter resolver
- `WithURLResolver(resolver)` - Custom URL resolver
- `WithDefinitions(definitions...)` - Add multiple path definitions at creation
- `WithProviders(providers...)` - Register multiple providers at creation

### Step 4: Add Path Definitions

```go
// Add the user avatars path definition
err := manager.AddDefinition(UserAvatarsPath)
if err != nil {
    log.Fatal(err)
}
```

## Basic Usage

### Simple URL Resolution

The most basic use case - resolving a resource URL with parameters:

```go
// Resolve user avatar URL
result, err := manager.DefinitionResolver().Resolve(ctx, "user-avatars", 
    &ResolveOptions{
        ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
            "user_id": "12345",
        }),
    })

if err != nil {
    log.Printf("Failed to resolve: %v", err)
    return
}

fmt.Printf("Avatar URL: %s\n", result.URL)
// Output: https://cdn.example.com/shared/global/users/avatars/12345.png
```

### Using Fluent API

The fluent API provides a cleaner syntax:

```go
// Same resolution using fluent API
result, err := manager.DefinitionResolver().Resolve(ctx, "user-avatars",
    ResolveOptions{}.WithValues(map[ParameterName]string{
        "user_id": "12345",
    }))
```

### Multiple Parameters

For resources with multiple parameters:

```go
// Define a path with multiple parameters
var GameConfigPath = PathDefinition{
    Name: "game-configs",
    Parameters: []*ParameterDefinition{
        {Name: "app", Required: true},
        {Name: "version", Required: true},
        {Name: "config_name", Required: true},
    },
    // ... patterns configuration
}

// Resolve with multiple parameters
result, err := manager.DefinitionResolver().Resolve(ctx, "game-configs",
    ResolveOptions{}.WithValues(map[ParameterName]string{
        "app":         "bike",
        "version":     "1.2",
        "config_name": "bike_game_1.json",
    }))
```

## Parameter Resolution

The parameter system supports multiple resolution strategies:

### 1. Static Values

The simplest approach - provide values directly:

```go
resolver := NewValuesParameterResolver(map[ParameterName]string{
    "user_id": "12345",
    "format":  "png",
})
```

### 2. Context-Based Resolution

Automatically extract parameters from context:

```go
// Parameters from client app context
clientAppResolver := NewClientAppParameterResolver(clientApps)

// Parameters from app context
appResolver := NewAppParameterResolver(apps)
```

### 3. Chained Resolution

Try multiple resolvers in sequence:

```go
// Create a chain: user values → context → defaults
chainedResolver, err := NewParameterResolvers(
    NewValuesParameterResolver(userProvidedParams),    // First try user values
    NewClientAppParameterResolver(clientApps),         // Then context
    NewValuesParameterResolver(map[ParameterName]string{ // Finally defaults
        "format": "png",
    }),
)
```

### 4. Dynamic Resolution

Create custom resolvers for complex logic:

```go
// Timestamp resolver with custom format
timestampResolver := ParameterResolverFunc(func(ctx context.Context, paramName ParameterName) (string, error) {
    if paramName == "timestamp" {
        return time.Now().Format("2006-01-02"), nil
    }
    return "", nil
})

// User-specific resolver
type UserDataResolver struct {
    userID string
    userData map[string]string
}

func (r *UserDataResolver) Resolve(ctx context.Context, paramName ParameterName) (string, error) {
    switch paramName {
    case "user_id":
        return r.userID, nil
    case "user_tier":
        return r.userData["tier"], nil
    default:
        return "", nil
    }
}
```

### 5. Parameter Validation

Add validation to ensure parameter integrity:

```go
// Built-in empty parameter validation
validatedResolver, err := NewValidatingParameterResolver(
    baseResolver,
    WithEmptyParameterValidation,
)

// Custom validation
resourceIDValidator := func(ctx context.Context, paramName ParameterName, value string) error {
    if paramName == "resource_id" && !isValidResourceID(value) {
        return errors.New("invalid resource ID format")
    }
    return nil
}

validatedResolver, err := NewValidatingParameterResolver(
    baseResolver,
    WithEmptyParameterValidation,
    resourceIDValidator,
)
```

## Provider Selection

### Automatic Selection by URL Type

The system automatically selects providers based on URL type:

```go
// Content URLs → CDN
contentResult, err := manager.DefinitionResolver().Resolve(ctx, "achievements",
    ResolveOptions{}.
        WithURLType(URLTypeContent).
        WithValues(params))
// URL: https://cdn.example.com/shared/global/achievements/...

// Operation URLs → GCS
operationResult, err := manager.DefinitionResolver().Resolve(ctx, "achievements",
    ResolveOptions{}.
        WithURLType(URLTypeOperation).
        WithValues(params))
// URL: https://storage.googleapis.com/global-achievements/...
```

### Explicit Provider Selection

Override automatic selection when needed:

```go
// Force GCS provider for special cases
result, err := manager.DefinitionResolver().Resolve(ctx, "user-avatars",
    ResolveOptions{}.
        WithProvider(provider.ProviderGCS).
        WithValues(map[ParameterName]string{
            "user_id": "12345",
        }))
```

### Provider Fallback Pattern

Implement graceful degradation when a preferred provider fails:

```go
func getResourceWithFallback(ctx context.Context, manager ResourceManager, name string, params map[ParameterName]string) (string, error) {
    // Try primary provider (GCS for operations)
    result, err := manager.DefinitionResolver().Resolve(ctx, name,
        ResolveOptions{}.
            WithProvider(provider.ProviderGCS).
            WithValues(params))
    
    if err == nil {
        return result.URL, nil
    }
    
    // Log the failure but don't fail completely
    log.Printf("Primary provider failed, falling back to default: %v", err)
    
    // Fallback to default provider
    result, err = manager.DefinitionResolver().Resolve(ctx, name,
        ResolveOptions{}.WithValues(params))
    
    if err != nil {
        return "", fmt.Errorf("all providers failed: %w", err)
    }
    
    return result.URL, nil
}
```

### Parameter Fallback with Context

The `DefaultFallbackParameterResolver` automatically provides common parameters:

```go
// Setup context with app and client app information
ctx := context.Background()
ctx = app.WithApp(ctx, &app.App{ID: 1, Name: "bike"})
ctx = useragent.WithUserAgent(ctx, &useragent.UserAgent{
    ClientAppID: 2, // Android
})

// The fallback resolver will automatically provide:
// - "client_app" -> "android" (from ClientAppID: 2)
// - "app" -> "bike" (from App.Name)
// 
// So you only need to provide resource-specific parameters
result, err := manager.DefinitionResolver().Resolve(ctx, "app-configs",
    ResolveOptions{}.WithValues(map[ParameterName]string{
        "config_name": "game_settings.json", // Only need specific params
        // "app" and "client_app" are provided by fallback resolver
    }))
```

### Advanced Parameter Resolution Chain

Combine multiple resolution strategies for maximum flexibility:

```go
// Create a sophisticated parameter resolution chain
func createAdvancedParameterChain(userParams map[ParameterName]string, clientApps map[int16]string, apps map[int16]string) (ParameterResolver, error) {
    // 1. User-provided parameters (highest priority)
    userResolver := NewValuesParameterResolver(userParams)
    
    // 2. Context-based parameters for routing
    contextResolver := NewRoutingParameterResolver(map[ParameterName]ParameterResolver{
        "client_app": NewClientAppParameterResolver(clientApps),
        "app":        NewAppParameterResolver(apps),
        "timestamp":  NewCurrentTimestampParameterResolver("2006-01-02"),
    })
    
    // 3. Default fallback parameters
    fallbackResolver := DefaultFallbackParameterResolver(clientApps, apps)
    
    // 4. Final defaults for common parameters
    defaultsResolver := NewValuesParameterResolver(map[ParameterName]string{
        "format":  "png",
        "version": "latest",
    })
    
    // Chain them together with validation
    chainedResolver, err := NewParameterResolvers(
        userResolver,     // Try user params first
        contextResolver,  // Then context-based
        fallbackResolver, // Then standard fallbacks
        defaultsResolver, // Finally, hardcoded defaults
    )
    if err != nil {
        return nil, err
    }
    
    // Add validation layer
    return NewValidatingParameterResolver(
        chainedResolver,
        WithEmptyParameterValidation,
    )
}
```

## Scope Management

### Understanding Scopes

Scopes provide hierarchical resource organization:

```
ScopeGlobal    → Shared across all apps and platforms
ScopeApp       → Specific to an app (bike, rower)
ScopeClientApp → Specific to a client platform (iOS, Android)
```

### Automatic Scope Selection

The system selects the most specific available scope based on context:

```go
// Setup context with app and client information
ctx := context.Background()
ctx = app.WithApp(ctx, &app.App{ID: 1, Name: "bike"})
ctx = useragent.WithUserAgent(ctx, &useragent.UserAgent{
    ClientAppID: 2, // Android
})

// System automatically selects most specific scope
result, err := manager.DefinitionResolver().Resolve(ctx, "preferences",
    ResolveOptions{}.WithValues(map[ParameterName]string{
        "user_id": "12345",
    }))
// Will use ClientApp scope if available
```

### Explicit Scope Override

Force a specific scope when needed:

```go
// Force global scope for shared resources
result, err := manager.DefinitionResolver().Resolve(ctx, "achievements",
    ResolveOptions{}.
        WithScope(ScopeGlobal).
        WithValues(map[ParameterName]string{
            "achievement_id": "gold_medal",
            "format":         "png",
        }))
```

### Scope-Based Path Patterns

Define different paths for different scopes:

```go
var PreferencesPath = PathDefinition{
    Name:          "preferences",
    AllowedScopes: []ScopeType{ScopeClientApp, ScopeGlobal},
    Patterns: map[provider.ProviderName]PathPatterns{
        provider.ProviderCDN: {
            Patterns: map[ScopeType]string{
                ScopeClientApp: "{client_app}/global/users/preferences/{user_id}.json",
                ScopeGlobal:    "shared/global/users/preferences/{user_id}.json",
            },
        },
    },
}
```

## Signed URLs

### Basic Signed URLs

Generate time-limited URLs for secure access:

```go
// Generate signed URL with default expiry
result, err := manager.DefinitionResolver().Resolve(ctx, "user-avatars",
    ResolveOptions{}.
        WithURLOptions(&provider.URLOptions{
            SignedURL: true,
        }).
        WithValues(map[ParameterName]string{
            "user_id": "secure_user",
        }))
```

### Custom Expiry Times

Set specific expiration periods:

```go
// Short-lived URL for temporary access (15 minutes)
tempResult, err := manager.DefinitionResolver().Resolve(ctx, "temp-file",
    ResolveOptions{}.
        WithURLOptions(&provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: 15 * time.Minute,
        }).
        WithValues(params))

// Long-lived URL for premium users (24 hours)
premiumResult, err := manager.DefinitionResolver().Resolve(ctx, "premium-content",
    ResolveOptions{}.
        WithURLOptions(&provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: 24 * time.Hour,
        }).
        WithValues(params))
```

### Security-Limited Expiry

Path definitions can enforce maximum expiry times:

```go
var SecureDocumentsPath = PathDefinition{
    Name: "secure-docs",
    URLOptions: &provider.URLOptions{
        SignedURL:    true,
        SignedExpiry: 1 * time.Hour, // Maximum allowed
    },
    // ... other configuration
}

// Request for 3 hours gets capped to 1 hour
result, err := manager.DefinitionResolver().Resolve(ctx, "secure-docs",
    ResolveOptions{}.
        WithURLOptions(&provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: 3 * time.Hour, // Will be capped to 1 hour
        }).
        WithValues(params))
```

### Dynamic Expiry Based on Content

Adjust expiry based on resource type:

```go
func getContentURL(manager ResourceManager, contentType string, resourceID string) (string, error) {
    // Determine expiry based on content type
    expiry := map[string]time.Duration{
        "video":    4 * time.Hour,   // Long for streaming
        "image":    2 * time.Hour,   // Medium for viewing
        "document": 1 * time.Hour,   // Short for documents
        "temp":     15 * time.Minute, // Very short for temp
    }[contentType]
    
    if expiry == 0 {
        expiry = 1 * time.Hour // Default
    }
    
    return manager.DefinitionResolver().Resolve(ctx, "content",
        ResolveOptions{}.
            WithURLOptions(&provider.URLOptions{
                SignedURL:    true,
                SignedExpiry: expiry,
            }).
            WithValues(map[ParameterName]string{
                "resource_id":   resourceID,
                "content_type": contentType,
            }))
}
```

## Advanced Patterns

### 1. Service-Specific Configurations

Create specialized services with pre-configured options:

```go
// User Service with avatar management
type UserService struct {
    resourceManager ResourceManager
    avatarOptions   *ResolveOptions
}

func NewUserService(manager ResourceManager) *UserService {
    return &UserService{
        resourceManager: manager,
        avatarOptions: &ResolveOptions{
            URLOptions: &provider.URLOptions{
                SignedURL:    true,
                SignedExpiry: 24 * time.Hour,
            },
        },
    }
}

func (s *UserService) GetUserAvatarURL(userID string) (string, error) {
    opts := *s.avatarOptions // Copy base options
    opts.ParamResolver = NewValuesParameterResolver(map[ParameterName]string{
        "user_id": userID,
    })
    
    result, err := s.resourceManager.DefinitionResolver().Resolve(
        context.Background(), "user-avatars", &opts)
    if err != nil {
        return "", err
    }
    return result.URL, nil
}
```

### 2. Batch Processing

Efficiently process multiple resources:

```go
type BatchProcessor struct {
    manager      ResourceManager
    baseOptions  *ResolveOptions
}

func (b *BatchProcessor) ProcessUserBatch(ctx context.Context, userIDs []string) []BatchResult {
    results := make([]BatchResult, 0, len(userIDs))
    
    for _, userID := range userIDs {
        opts := *b.baseOptions
        opts.ParamResolver = NewValuesParameterResolver(map[ParameterName]string{
            "user_id": userID,
        })
        
        result, err := b.manager.DefinitionResolver().Resolve(ctx, "user-avatars", &opts)
        results = append(results, BatchResult{
            UserID: userID,
            URL:    result.URL,
            Error:  err,
        })
    }
    
    return results
}
```

### 3. Caching Resolved URLs

Implement caching for frequently accessed resources:

```go
type CachedResourceResolver struct {
    manager ResourceManager
    cache   map[string]*CachedEntry
    mu      sync.RWMutex
    ttl     time.Duration
}

type CachedEntry struct {
    URL       string
    ExpiresAt time.Time
}

func (c *CachedResourceResolver) Resolve(ctx context.Context, name string, params map[ParameterName]string) (string, error) {
    // Generate cache key
    key := generateCacheKey(name, params)
    
    // Check cache
    c.mu.RLock()
    if entry, exists := c.cache[key]; exists && entry.ExpiresAt.After(time.Now()) {
        c.mu.RUnlock()
        return entry.URL, nil
    }
    c.mu.RUnlock()
    
    // Resolve and cache
    result, err := c.manager.DefinitionResolver().Resolve(ctx, name,
        ResolveOptions{}.WithValues(params))
    if err != nil {
        return "", err
    }
    
    c.mu.Lock()
    c.cache[key] = &CachedEntry{
        URL:       result.URL,
        ExpiresAt: time.Now().Add(c.ttl),
    }
    c.mu.Unlock()
    
    return result.URL, nil
}
```

### 4. Multi-Environment Support

Configure different providers for different environments:

```go
func CreateEnvironmentSpecificManager(env string) (ResourceManager, error) {
    var cdnConfig provider.CDNConfig
    var gcsConfig provider.GCSConfig
    
    switch env {
    case "production":
        cdnConfig = provider.CDNConfig{
            BaseURL:    "https://cdn.aviron.com",
            SigningKey: os.Getenv("PROD_SIGNING_KEY"),
            Expiry:     24 * time.Hour,
        }
        gcsConfig = provider.GCSConfig{
            BucketName: "aviron-prod-resources",
            Expiry:     24 * time.Hour,
        }
        
    case "staging":
        cdnConfig = provider.CDNConfig{
            BaseURL:    "https://cdn-staging.aviron.com",
            SigningKey: os.Getenv("STAGING_SIGNING_KEY"),
            Expiry:     2 * time.Hour,
        }
        gcsConfig = provider.GCSConfig{
            BucketName: "aviron-staging-resources",
            Expiry:     2 * time.Hour,
        }
        
    default: // development
        cdnConfig = provider.CDNConfig{
            BaseURL:    "http://localhost:8080",
            SigningKey: "dev-key",
            Expiry:     30 * time.Minute,
        }
        gcsConfig = provider.GCSConfig{
            BucketName: "aviron-dev-resources",
            Expiry:     30 * time.Minute,
        }
    }
    
    // Setup providers
    cdnProvider := provider.NewCDNProvider(cdnConfig)
    gcsProvider, err := provider.NewGCSProvider(context.Background(), gcsConfig)
    if err != nil {
        return nil, err
    }
    
    // Create provider registry
    providerRegistry := provider.NewRegistry()
    providerRegistry.Register(cdnProvider)
    providerRegistry.Register(gcsProvider)
    
    // Create manager with options
    return NewResourceManager(
        WithProviderRegistry(providerRegistry),
        // Add other options as needed
    )
}
```

## Production Best Practices

### 1. Centralized Path Definitions

Keep all path definitions in a central location:

```go
// paths/definitions.go
package paths

var (
    UserAvatars = PathDefinition{...}
    Achievements = PathDefinition{...}
    GameConfigs = PathDefinition{...}
    WorkoutData = PathDefinition{...}
)

// paths/registry.go
func RegisterAllPaths(manager ResourceManager) error {
    definitions := []PathDefinition{
        UserAvatars,
        Achievements,
        GameConfigs,
        WorkoutData,
    }
    
    for _, def := range definitions {
        if err := manager.AddDefinition(def); err != nil {
            return fmt.Errorf("failed to register %s: %w", def.Name, err)
        }
    }
    return nil
}
```

### 2. Reusable Option Sets

Create standard option sets for common patterns:

```go
var (
    // Public content - no signing
    PublicContentOptions = &ResolveOptions{
        URLType: ptr(URLTypeContent),
    }
    
    // Secure content - signed with medium expiry
    SecureContentOptions = &ResolveOptions{
        URLType: ptr(URLTypeContent),
        URLOptions: &provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: 2 * time.Hour,
        },
    }
    
    // Admin operations - GCS with short expiry
    AdminOperationOptions = &ResolveOptions{
        Provider: ptr(provider.ProviderGCS),
        URLType:  ptr(URLTypeOperation),
        URLOptions: &provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: 15 * time.Minute,
        },
    }
)

func ptr[T any](v T) *T { return &v }
```

### 3. Error Handling

Implement comprehensive error handling:

```go
func resolveWithRetry(manager ResourceManager, name string, params map[ParameterName]string) (string, error) {
    var lastErr error
    
    for attempt := 0; attempt < 3; attempt++ {
        result, err := manager.DefinitionResolver().Resolve(
            context.Background(), name,
            ResolveOptions{}.WithValues(params))
        
        if err == nil {
            return result.URL, nil
        }
        
        lastErr = err
        
        // Check error type for retry decision
        if errors.Is(err, ErrDefinitionNotFound) {
            return "", err // Don't retry on definition errors
        }
        
        if errors.Is(err, ErrTemplateResolutionFailed) {
            return "", err // Don't retry on parameter errors
        }
        
        // Exponential backoff for other errors
        time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
    }
    
    return "", fmt.Errorf("failed after 3 attempts: %w", lastErr)
}
```

### 4. Monitoring and Metrics

Add instrumentation for production monitoring:

```go
type InstrumentedResolver struct {
    manager ResourceManager
    metrics MetricsCollector
}

func (i *InstrumentedResolver) Resolve(ctx context.Context, name string, opts *ResolveOptions) (*ResolvedResource, error) {
    start := time.Now()
    
    result, err := i.manager.DefinitionResolver().Resolve(ctx, name, opts)
    
    duration := time.Since(start)
    i.metrics.RecordResolution(name, err == nil, duration)
    
    if err != nil {
        i.metrics.RecordError(name, err)
    }
    
    return result, err
}
```

### 5. Graceful Degradation

Implement fallback strategies for resilience:

```go
type ResilientResolver struct {
    primary   ResourceManager
    fallback  ResourceManager
    cache     Cache
}

func (r *ResilientResolver) Resolve(ctx context.Context, name string, params map[ParameterName]string) (string, error) {
    // Try cache first
    if url, found := r.cache.Get(name, params); found {
        return url, nil
    }
    
    // Try primary resolver
    if result, err := r.primary.DefinitionResolver().Resolve(ctx, name,
        ResolveOptions{}.WithValues(params)); err == nil {
        r.cache.Set(name, params, result.URL)
        return result.URL, nil
    }
    
    // Fallback to secondary
    if r.fallback != nil {
        if result, err := r.fallback.DefinitionResolver().Resolve(ctx, name,
            ResolveOptions{}.WithValues(params)); err == nil {
            return result.URL, nil
        }
    }
    
    // Last resort: construct URL manually
    return constructFallbackURL(name, params), nil
}
```

## Troubleshooting

### Common Issues and Solutions

#### 1. Missing Required Parameters
```go
// Error: ErrTemplateResolutionFailed
// Solution: Ensure all required parameters are provided
result, err := manager.DefinitionResolver().Resolve(ctx, "user-avatars",
    ResolveOptions{}.WithValues(map[ParameterName]string{
        "user_id": "12345", // Required parameter
    }))
```

#### 2. Invalid Scope
```go
// Error: ErrScopeNotAllowed
// Solution: Use an allowed scope for the resource
definition, _ := manager.GetDefinition("user-avatars")
fmt.Printf("Allowed scopes: %v\n", definition.AllowedScopes)
```

#### 3. Provider Not Found
```go
// Error: Provider not found
// Solution: Ensure provider is registered
err := providerRegistry.Register(cdnProvider)
```

#### 4. Template Resolution Failed
```go
// Error: Template contains invalid characters
// Solution: Validate parameter values
if !isValidUserID(userID) {
    return errors.New("invalid user ID format")
}
```

### Debug Logging

Enable detailed logging for troubleshooting:

```go
type LoggingResolver struct {
    manager ResourceManager
    logger  Logger
}

func (l *LoggingResolver) Resolve(ctx context.Context, name string, opts *ResolveOptions) (*ResolvedResource, error) {
    l.logger.Debugf("Resolving resource: %s", name)
    l.logger.Debugf("Options: %+v", opts)
    
    result, err := l.manager.DefinitionResolver().Resolve(ctx, name, opts)
    
    if err != nil {
        l.logger.Errorf("Resolution failed: %s - %v", name, err)
        return nil, err
    }
    
    l.logger.Debugf("Resolved URL: %s", result.URL)
    l.logger.Debugf("Provider: %s", result.Provider)
    
    return result, nil
}
```

## Quick Reference

### Essential Patterns

```go
// 1. Basic resolution
result, _ := manager.DefinitionResolver().Resolve(ctx, "resource-name",
    ResolveOptions{}.WithValues(map[ParameterName]string{
        "param": "value",
    }))

// 2. Signed URL
result, _ := manager.DefinitionResolver().Resolve(ctx, "resource-name",
    ResolveOptions{}.
        WithValues(params).
        WithURLOptions(&provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: time.Hour,
        }))

// 3. Specific provider
result, _ := manager.DefinitionResolver().Resolve(ctx, "resource-name",
    ResolveOptions{}.
        WithValues(params).
        WithProvider(provider.ProviderGCS))

// 4. URL type selection
result, _ := manager.DefinitionResolver().Resolve(ctx, "resource-name",
    ResolveOptions{}.
        WithValues(params).
        WithURLType(URLTypeOperation))

// 5. Scope override
result, _ := manager.DefinitionResolver().Resolve(ctx, "resource-name",
    ResolveOptions{}.
        WithValues(params).
        WithScope(ScopeGlobal))

// 6. Complete example
result, _ := manager.DefinitionResolver().Resolve(ctx, "resource-name",
    ResolveOptions{}.
        WithProvider(provider.ProviderGCS).
        WithURLType(URLTypeOperation).
        WithScope(ScopeApp).
        WithURLOptions(&provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: 30 * time.Minute,
        }).
        WithValues(map[ParameterName]string{
            "app":         "bike",
            "resource_id": "config_123",
        }))
```

### Decision Matrix

| Use Case | URL Type | Provider | Signed | Typical Expiry |
|----------|----------|----------|--------|----------------|
| Public assets | Content | CDN | No | - |
| User avatars | Content | CDN | Yes | 24h |
| Temp downloads | Content | CDN | Yes | 15m |
| Admin uploads | Operation | GCS | Yes | 15m |
| Backups | Operation | GCS | Yes | 1h |
| Premium content | Content | CDN | Yes | 7d |

### Common Parameter Names

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

This guide provides a complete path from basic usage to advanced production patterns. Start with the basics and gradually incorporate more advanced features as your needs grow.