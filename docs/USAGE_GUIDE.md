# ResourceManager Usage Guide

This guide demonstrates how to use ResourceManager with various ResolveOptions to handle different resource resolution scenarios in production applications.

## Table of Contents

1. [Basic Setup](#basic-setup)
2. [Basic Resource Resolution](#basic-resource-resolution)
3. [Provider Selection](#provider-selection)
4. [URL Type-Based Resolution](#url-type-based-resolution)
5. [Signed URL Generation](#signed-url-generation)
6. [Scope Management](#scope-management)
7. [Complex Scenarios](#complex-scenarios)
8. [Fluent API Usage](#fluent-api-usage)
9. [Error Handling](#error-handling)
10. [Performance Best Practices](#performance-best-practices)
11. [Production Patterns](#production-patterns)

## Basic Setup

First, create and configure your ResourceManager:

```go
// Setup ResourceManager
func setupResourceManager() ResourceManager {
    // Create template resolver for path interpolation
    templateResolver, err := NewBracesTemplateResolver()
    if err != nil {
        log.Fatal(err)
    }

    // Create scope selector
    scopeSelector := NewDefaultScopeSelector()

    // Create path resolver
    pathResolver := NewPathResolver(templateResolver, nil)

    // Create registries
    definitionRegistry := NewPathDefinitionRegistry()
    providerRegistry := provider.NewRegistry()

    // Register providers
    cdnProvider := provider.NewCDNProvider(provider.CDNConfig{
        BaseURL:    "https://cdn.example.com",
        SigningKey: "your-signing-key",
        Expiry:     1 * time.Hour,
    })
    providerRegistry.Register(provider.ProviderCDN, cdnProvider)

    gcsProvider, _ := provider.NewGCSProvider(ctx, provider.GCSConfig{
        BucketName: "your-bucket",
        Expiry:     1 * time.Hour,
    })
    providerRegistry.Register(provider.ProviderGCS, gcsProvider)

    return NewResourceManager(pathResolver, scopeSelector, definitionRegistry, providerRegistry)
}

// Add your path definitions
manager.AddDefinition(PathDefinition{
    Name:          "user-avatars",
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
        {Name: "user_id", Required: true},
    },
})
```

## Basic Resource Resolution

### Simple Resolution (Default Provider)

```go
// Resolve user avatar with default provider (CDN for content)
result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &ResolveOptions{
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
        "user_id": "12345",
    }),
})
if err != nil {
    return err
}

fmt.Printf("Avatar URL: %s\n", result.URL)
// Output: Avatar URL: https://cdn.example.com/shared/global/users/avatars/12345.png
```

## Provider Selection

### Explicit Provider Selection

```go
// Force use of GCS provider instead of default CDN
gcsProvider := provider.ProviderGCS
result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &ResolveOptions{
    Provider: &gcsProvider,
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
        "user_id": "12345",
    }),
})

fmt.Printf("GCS URL: %s\n", result.URL)
// Output: GCS URL: https://storage.googleapis.com/global-avatars/12345.png
```

## URL Type-Based Resolution

### Content vs Operation URLs

```go
// Content URL (optimized for delivery - uses CDN)
contentURLType := URLTypeContent
contentResult, err := manager.PathDefinitionResolver().Resolve(ctx, "achievements", &ResolveOptions{
    URLType: &contentURLType,
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
        "achievement_id": "gold_medal",
        "format":         "png",
    }),
})

// Operation URL (for management - uses GCS)
operationURLType := URLTypeOperation
operationResult, err := manager.PathDefinitionResolver().Resolve(ctx, "achievements", &ResolveOptions{
    URLType: &operationURLType,
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
        "achievement_id": "gold_medal",
        "format":         "png",
    }),
})

fmt.Printf("Content URL: %s\n", contentResult.URL)
fmt.Printf("Operation URL: %s\n", operationResult.URL)
```

## Signed URL Generation

### Basic Signed URLs

```go
// Generate signed URL with default expiry
result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &ResolveOptions{
    URLOptions: &provider.URLOptions{
        SignedURL: true,
    },
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
        "user_id": "secure_user",
    }),
})

fmt.Printf("Signed URL: %s\n", result.URL)
```

### Custom Expiry Signed URLs

```go
// Generate signed URL with custom expiry
result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &ResolveOptions{
    URLOptions: &provider.URLOptions{
        SignedURL:    true,
        SignedExpiry: 2 * time.Hour, // Expires in 2 hours
    },
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
        "user_id": "temp_user",
    }),
})

fmt.Printf("2-Hour Signed URL: %s\n", result.URL)
```

### Different Expiry Times for Different Use Cases

```go
// Short expiry for admin operations (5 minutes)
adminResult, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &ResolveOptions{
    URLOptions: &provider.URLOptions{
        SignedURL:    true,
        SignedExpiry: 5 * time.Minute,
    },
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
        "user_id": "admin_user",
    }),
})

// Long expiry for premium users (24 hours)
premiumResult, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &ResolveOptions{
    URLOptions: &provider.URLOptions{
        SignedURL:    true,
        SignedExpiry: 24 * time.Hour,
    },
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
        "user_id": "premium_user",
    }),
})

// Medium expiry for sharing (2 hours)
shareResult, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &ResolveOptions{
    URLOptions: &provider.URLOptions{
        SignedURL:    true,
        SignedExpiry: 2 * time.Hour,
    },
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
        "user_id": "shared_user",
    }),
})
```

## Scope Management

### Global vs App-Specific Resources

```go
// Global scope (shared across all apps)
globalScope := ScopeGlobal
globalResult, err := manager.PathDefinitionResolver().Resolve(ctx, "achievements", &ResolveOptions{
    Scope: &globalScope,
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
        "achievement_id": "grand_master",
        "format":         "svg",
    }),
})

// App-specific scope
appScope := ScopeApp
bikeResult, err := manager.PathDefinitionResolver().Resolve(ctx, "achievements", &ResolveOptions{
    Scope: &appScope,
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
        "app":            "bike",
        "achievement_id": "century_ride",
        "format":         "png",
    }),
})

fmt.Printf("Global Achievement: %s\n", globalResult.URL)
fmt.Printf("Bike Achievement: %s\n", bikeResult.URL)
```

## Complex Scenarios

### Combining All Options

```go
// Complex scenario: GCS + operation URL + signed URL + app scope
gcsProvider := provider.ProviderGCS
operationURLType := URLTypeOperation
appScope := ScopeApp

result, err := manager.PathDefinitionResolver().Resolve(ctx, "achievements", &ResolveOptions{
    Provider: &gcsProvider,
    URLType:  &operationURLType,
    Scope:    &appScope,
    URLOptions: &provider.URLOptions{
        SignedURL:    true,
        SignedExpiry: 30 * time.Minute,
    },
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
        "app":            "rower",
        "achievement_id": "platinum_medal",
        "format":         "webp",
    }),
})

fmt.Printf("Complex URL: %s\n", result.URL)
// This gives you a signed GCS operation URL for a rower app achievement
```

## Fluent API Usage

### Simple Fluent API

```go
// Using WithValues helper
result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars",
    ResolveOptions{}.WithValues(map[ParameterName]string{
        "user_id": "fluent_user",
    }))
```

### Chained Fluent API

```go
// Chaining multiple options
result, err := manager.PathDefinitionResolver().Resolve(ctx, "achievements",
    ResolveOptions{}.
        WithValues(map[ParameterName]string{
            "achievement_id": "diamond_medal",
            "format":         "png",
        }).
        WithProvider(provider.ProviderGCS).
        WithURLOptions(&provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: 1 * time.Hour,
        }).
        WithURLType(URLTypeOperation))
```

### Building Reusable Options

```go
// Create base options for common patterns
baseSecureOptions := ResolveOptions{}.
    WithURLOptions(&provider.URLOptions{
        SignedURL:    true,
        SignedExpiry: 1 * time.Hour,
    })

// Use base options with specific parameters
userResult, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars",
    baseSecureOptions.WithValues(map[ParameterName]string{
        "user_id": "secure_user_1",
    }))

adminResult, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars",
    baseSecureOptions.
        WithProvider(provider.ProviderGCS).
        WithValues(map[ParameterName]string{
            "user_id": "admin_user",
        }))
```

## Error Handling

### Common Error Scenarios

```go
// Handle missing required parameters
result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &ResolveOptions{
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{}), // Empty
})
if err != nil {
    if errors.Is(err, ErrTemplateResolutionFailed) {
        log.Printf("Missing required parameters: %v", err)
        // Handle missing parameters
    }
}

// Handle invalid scope
invalidScope := ScopeApp // user-avatars only supports global
result, err = manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &ResolveOptions{
    Scope: &invalidScope,
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
        "user_id": "test_user",
    }),
})
if err != nil {
    if errors.Is(err, ErrScopeNotAllowed) {
        log.Printf("Invalid scope requested: %v", err)
        // Handle scope error
    }
}

// Handle non-existent resource
result, err = manager.PathDefinitionResolver().Resolve(ctx, "non-existent", &ResolveOptions{
    ParamResolver: NewValuesParameterResolver(map[ParameterName]string{
        "param": "value",
    }),
})
if err != nil {
    if errors.Is(err, ErrDefinitionNotFound) {
        log.Printf("Resource not found: %v", err)
        // Handle missing resource
    }
}
```

### Graceful Degradation

```go
// Try specific provider, fallback to default
func getResourceURL(ctx context.Context, manager ResourceManager, resourceName PathDefinitionName, params map[ParameterName]string) (string, error) {
    // First try with preferred provider
    gcsProvider := provider.ProviderGCS
    result, err := manager.PathDefinitionResolver().Resolve(ctx, resourceName, &ResolveOptions{
        Provider: &gcsProvider,
        ParamResolver: NewValuesParameterResolver(params),
    })
    if err == nil {
        return result.URL, nil
    }

    // Fallback to default provider
    result, err = manager.PathDefinitionResolver().Resolve(ctx, resourceName, &ResolveOptions{
        ParamResolver: NewValuesParameterResolver(params),
    })
    if err != nil {
        return "", fmt.Errorf("failed to resolve resource with any provider: %w", err)
    }

    return result.URL, nil
}
```

## Performance Best Practices

### Reuse ResolveOptions Structures

```go
// Don't create new options for every request
// ❌ Bad
func getUserAvatar(userID string) string {
    result, _ := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &ResolveOptions{
        URLOptions: &provider.URLOptions{SignedURL: true, SignedExpiry: 1 * time.Hour},
        ParamResolver: NewValuesParameterResolver(map[ParameterName]string{"user_id": userID}),
    })
    return result.URL
}

// ✅ Good - Reuse base options
var baseSecureOptions = &ResolveOptions{
    URLOptions: &provider.URLOptions{
        SignedURL:    true,
        SignedExpiry: 1 * time.Hour,
    },
}

func getUserAvatar(userID string) string {
    opts := *baseSecureOptions // Copy struct
    opts.ParamResolver = NewValuesParameterResolver(map[ParameterName]string{"user_id": userID})
    
    result, _ := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &opts)
    return result.URL
}
```

### Batch Processing

```go
// Process multiple resources efficiently
func processUserBatch(userIDs []string) []string {
    baseOptions := &ResolveOptions{
        URLOptions: &provider.URLOptions{
            SignedURL:    true,
            SignedExpiry: 2 * time.Hour,
        },
    }

    urls := make([]string, 0, len(userIDs))
    for _, userID := range userIDs {
        opts := *baseOptions
        opts.ParamResolver = NewValuesParameterResolver(map[ParameterName]string{
            "user_id": userID,
        })

        result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &opts)
        if err != nil {
            log.Printf("Failed to process user %s: %v", userID, err)
            continue
        }
        urls = append(urls, result.URL)
    }
    
    return urls
}
```

## Production Patterns

### Service-Specific Configurations

```go
// User Service
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
                SignedExpiry: 24 * time.Hour, // Long expiry for user avatars
            },
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

// Admin Service
type AdminService struct {
    resourceManager ResourceManager
    adminOptions    *ResolveOptions
}

func NewAdminService(manager ResourceManager) *AdminService {
    gcsProvider := provider.ProviderGCS
    operationURLType := URLTypeOperation
    
    return &AdminService{
        resourceManager: manager,
        adminOptions: &ResolveOptions{
            Provider: &gcsProvider,
            URLType:  &operationURLType,
            URLOptions: &provider.URLOptions{
                SignedURL:    true,
                SignedExpiry: 15 * time.Minute, // Short expiry for admin operations
            },
        },
    }
}

// Content Delivery Service
type CDNService struct {
    resourceManager ResourceManager
    cdnOptions      *ResolveOptions
}

func NewCDNService(manager ResourceManager) *CDNService {
    contentURLType := URLTypeContent
    
    return &CDNService{
        resourceManager: manager,
        cdnOptions: &ResolveOptions{
            URLType: &contentURLType,
            // No signed URLs - public content
        },
    }
}
```

### Environment-Specific Configurations

```go
// Development environment
func setupDevelopmentResourceManager() ResourceManager {
    // Use shorter expiry times for development
    // Enable more logging
    // Use test buckets/CDNs
}

// Production environment  
func setupProductionResourceManager() ResourceManager {
    // Use longer expiry times
    // Enable performance optimizations
    // Use production buckets/CDNs
}

// Staging environment
func setupStagingResourceManager() ResourceManager {
    // Mix of development and production settings
    // Use staging buckets/CDNs
}
```

This usage guide provides comprehensive examples of how to use ResourceManager with ResolveOptions in real-world scenarios, covering everything from basic usage to advanced production patterns.