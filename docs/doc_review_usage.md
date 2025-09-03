# Resource Package: Rich Usage Examples

## Overview

The Resource package is a comprehensive resource management system for the Aviron Game Server that handles dynamic path resolution, URL generation, parameter injection, and multi-provider storage support. This document provides detailed usage examples covering all aspects of the system.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Core Components Setup](#core-components-setup)
3. [Basic Resource Resolution](#basic-resource-resolution)
4. [Parameter Handling](#parameter-handling)
5. [Signed URL Generation](#signed-url-generation)
6. [Multi-Provider Support](#multi-provider-support)
7. [Scope-Based Resolution](#scope-based-resolution)
8. [Path Definition Management](#path-definition-management)
9. [Template Resolution](#template-resolution)
10. [Error Handling](#error-handling)
11. [Advanced Patterns](#advanced-patterns)
12. [Integration Examples](#integration-examples)
13. [Testing Strategies](#testing-strategies)
14. [Performance Optimization](#performance-optimization)

## Quick Start

### Basic Resource Manager Setup

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
    templateResolver := resource.NewTemplateResolver()
    
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
    
    err := manager.AddDefinition(definition)
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

## Core Components Setup

### 1. Template Resolver Configuration

```go
// Basic template resolver
templateResolver := resource.NewTemplateResolver()

// Template resolver with custom security settings
templateResolver := resource.NewTemplateResolverWithConfig(&resource.TemplateConfig{
    MaxPathLength:      2000,
    MaxParameterLength: 500,
    SafetyPatterns: []string{
        `\.\./`,        // Path traversal
        `/etc/`,        // System directories
        `/proc/`,       // Process directories
        `\\windows\\`,  // Windows system paths
    },
})
```

### 2. Parameter Resolver Setup

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

// Timestamp resolver
timestampResolver := resource.NewCurrentTimestampParameterResolver("2006-01-02T15:04:05Z")

// Chain multiple resolvers
chainedResolver, err := resource.NewParameterResolvers(
    valuesResolver,
    clientAppResolver,
    appResolver,
    timestampResolver,
)
if err != nil {
    log.Fatal(err)
}
```

### 3. Provider Registry Configuration

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
    BucketName:     "aviron-resources",
    ProjectID:      "aviron-prod",
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

## Basic Resource Resolution

### Simple Resource Resolution

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

### Resolution with Multiple Parameters

```go
func resolveWithMultipleParams(manager resource.ResourceManager) {
    ctx := context.Background()
    
    opts := resource.ResolveOptions{}.WithValues(map[resource.ParameterName]string{
        resource.ResourceID:     "workout_456",
        resource.Version:        "1.5.0",
        resource.ResourceFormat: "mp4",
        resource.Timestamp:      time.Now().Format(time.RFC3339),
        "workout_type":          "strength",
        "difficulty":            "intermediate",
    })
    
    result, err := manager.DefinitionResolver().Resolve(ctx, "workout-video", &opts)
    if err != nil {
        log.Printf("Resolution failed: %v", err)
        return
    }
    
    fmt.Printf("Video URL: %s\n", result.URL)
}
```

## Parameter Handling

### Custom Parameter Resolvers

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

// Usage
userResolver := NewUserParameterResolver("user123", map[string]string{
    "tier":         "premium",
    "subscription": "pro",
})

opts := resource.ResolveOptions{
    ParamResolver: userResolver,
}
```

### Parameter Validation

```go
// Custom validator for resource IDs
func validateResourceID(ctx context.Context, paramName resource.ParameterName, value string) error {
    if paramName != resource.ResourceID {
        return nil // Not our parameter
    }
    
    if len(value) < 3 {
        return fmt.Errorf("resource ID too short")
    }
    
    if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, value); !matched {
        return fmt.Errorf("resource ID contains invalid characters")
    }
    
    return nil
}

// Create validating resolver
validatingResolver, err := resource.NewValidatingParameterResolver(
    baseResolver,
    validateResourceID,
    resource.WithEmptyParameterValidation,
)
if err != nil {
    log.Fatal(err)
}
```

### Definition-Based Parameter Resolution

```go
// Define parameter with validation rules
resourceIDDef := &resource.ParameterDefinition{
    Name:         resource.ResourceID,
    Required:     true,
    DefaultValue: "",
    Description:  "Unique resource identifier (alphanumeric, min 3 chars)",
}

// Create resolver that enforces the definition
defResolver, err := resource.NewDefinitionParameterResolver(
    resourceIDDef,
    baseParameterResolver,
)
if err != nil {
    log.Fatal(err)
}

// Use in routing resolver
routingResolver := resource.NewRoutingParameterResolver(
    map[resource.ParameterName]resource.ParameterResolver{
        resource.ResourceID: defResolver,
    },
)
```

## Signed URL Generation

### Basic Signed URLs

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
        "access_level":      func() string {
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

### Custom Expiry Based on Content Type

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

## Multi-Provider Support

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

### URL Type-Based Provider Selection

```go
func resolveByURLType(manager resource.ResourceManager) {
    ctx := context.Background()
    
    // Content access (likely CDN)
    contentOpts := resource.ResolveOptions{}.
        WithURLType(resource.URLTypeContent).
        WithValues(map[resource.ParameterName]string{
            resource.ResourceID: "image_123",
        })
    
    contentResult, err := manager.DefinitionResolver().Resolve(ctx, "user-image", &contentOpts)
    if err != nil {
        log.Printf("Content resolution failed: %v", err)
        return
    }
    
    // Management operations (likely GCS)
    operationOpts := resource.ResolveOptions{}.
        WithURLType(resource.URLTypeOperation).
        WithValues(map[resource.ParameterName]string{
            resource.ResourceID: "image_123",
        })
    
    operationResult, err := manager.DefinitionResolver().Resolve(ctx, "user-image", &operationOpts)
    if err != nil {
        log.Printf("Operation resolution failed: %v", err)
        return
    }
    
    fmt.Printf("Content URL (%s): %s\n", contentResult.Provider, contentResult.URL)
    fmt.Printf("Operation URL (%s): %s\n", operationResult.Provider, operationResult.URL)
}
```

### Fallback Provider Logic

```go
func resolveWithFallback(manager resource.ResourceManager) {
    ctx := context.Background()
    
    // Try primary provider first
    primaryOpts := resource.ResolveOptions{}.
        WithProvider(provider.ProviderCDN).
        WithValues(map[resource.ParameterName]string{
            resource.ResourceID: "critical_asset_123",
        })
    
    result, err := manager.DefinitionResolver().Resolve(ctx, "critical-asset", &primaryOpts)
    if err != nil {
        log.Printf("Primary provider failed, trying fallback: %v", err)
        
        // Fallback to GCS
        fallbackOpts := resource.ResolveOptions{}.
            WithProvider(provider.ProviderGCS).
            WithValues(map[resource.ParameterName]string{
                resource.ResourceID: "critical_asset_123",
            })
        
        result, err = manager.DefinitionResolver().Resolve(ctx, "critical-asset", &fallbackOpts)
        if err != nil {
            log.Printf("Fallback also failed: %v", err)
            return
        }
        
        fmt.Printf("Using fallback provider: %s\n", result.Provider)
    }
    
    fmt.Printf("Final URL: %s\n", result.URL)
}
```

## Scope-Based Resolution

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
    
    // Force App scope for app-specific resources
    appScope := resource.ScopeApp
    appOpts := resource.ResolveOptions{
        Scope: &appScope,
    }.WithValues(map[resource.ParameterName]string{
        resource.ResourceID: "bike_specific_config",
        resource.App:        "bike",
    })
    
    appResult, err := manager.DefinitionResolver().Resolve(ctx, "app-config", &appOpts)
    if err != nil {
        log.Printf("App scope resolution failed: %v", err)
        return
    }
    
    fmt.Printf("Global resource: %s\n", globalResult.URL)
    fmt.Printf("App-specific resource: %s\n", appResult.URL)
}
```

### Scope Validation

```go
func validateScopeAccess(manager resource.ResourceManager) {
    ctx := context.Background()
    
    // Try to access ClientApp scope without proper context
    clientAppScope := resource.ScopeClientApp
    opts := resource.ResolveOptions{
        Scope: &clientAppScope,
    }.WithValues(map[resource.ParameterName]string{
        resource.ResourceID: "user_preferences",
    })
    
    result, err := manager.DefinitionResolver().Resolve(ctx, "user-preferences", &opts)
    if err != nil {
        // Should fail because no ClientApp context is available
        fmt.Printf("Expected error for invalid scope: %v\n", err)
        
        // Now try with proper context
        ctx = useragent.WithUserAgent(ctx, &useragent.UserAgent{
            ClientAppID: 1,
            Version:     "2.0.0",
        })
        
        result, err = manager.DefinitionResolver().Resolve(ctx, "user-preferences", &opts)
        if err != nil {
            log.Printf("Still failed with context: %v", err)
            return
        }
        
        fmt.Printf("Success with proper context: %s\n", result.URL)
    }
}
```

## Path Definition Management

### Dynamic Path Definition Registration

```go
func registerDynamicDefinitions(manager resource.ResourceManager) {
    // Define a workout video resource
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
        },
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
    
    // Define a user progress backup resource
    userProgressDef := resource.PathDefinition{
        Name:        "user-progress-backup",
        DisplayName: "User Progress Backup",
        Description: "Encrypted user progress data backups",
        AllowedScopes: []resource.ScopeType{
            resource.ScopeClientApp, // Only client-app specific
        },
        Patterns: map[provider.ProviderName]resource.PathPatterns{
            provider.ProviderGCS: {
                Patterns: map[resource.ScopeType]string{
                    resource.ScopeClientApp: "backups/{client_app}/users/{resource_id}/{timestamp}.encrypted",
                },
                URLType: resource.URLTypeOperation,
            },
        },
        Parameters: []*resource.ParameterDefinition{
            {Name: resource.ResourceID, Required: true, Description: "User ID"},
            {Name: resource.Timestamp, Required: true, Description: "Backup timestamp"},
            {Name: resource.Version, DefaultValue: "1.0.0", Description: "Backup format version"},
        },
        URLOptions: &provider.URLOptions{
            SignedURL:    true,          // Always signed for backups
            SignedExpiry: 24 * time.Hour, // 24-hour access
        },
    }
    
    err = manager.AddDefinition(userProgressDef)
    if err != nil {
        log.Printf("Failed to register user progress definition: %v", err)
        return
    }
    
    fmt.Printf("Successfully registered %d definitions\n", len(manager.GetAllDefinitions()))
}
```

### Definition Updates and Removal

```go
func manageDefinitionLifecycle(manager resource.ResourceManager) {
    // Get existing definition
    definition, err := manager.GetDefinition("user-avatar")
    if err != nil {
        log.Printf("Definition not found: %v", err)
        return
    }
    
    fmt.Printf("Current definition: %s - %s\n", definition.Name, definition.Description)
    
    // Remove old definition
    err = manager.RemoveDefinition("user-avatar")
    if err != nil {
        log.Printf("Failed to remove definition: %v", err)
        return
    }
    
    // Add updated definition
    updatedDefinition := *definition // Copy
    updatedDefinition.Description = "Updated user avatar with WebP support"
    updatedDefinition.Parameters = append(updatedDefinition.Parameters, 
        &resource.ParameterDefinition{
            Name:         "compression",
            DefaultValue: "high",
            Description:  "Image compression level",
            Required:     false,
        },
    )
    
    err = manager.AddDefinition(updatedDefinition)
    if err != nil {
        log.Printf("Failed to add updated definition: %v", err)
        return
    }
    
    fmt.Println("Definition successfully updated")
}
```

### Batch Definition Operations

```go
func batchDefinitionOperations(manager resource.ResourceManager) {
    // Get all current definitions
    allDefinitions := manager.GetAllDefinitions()
    fmt.Printf("Current definitions: %d\n", len(allDefinitions))
    
    // Filter definitions by criteria
    var cdnOnlyDefinitions []*resource.PathDefinition
    var gcsDefinitions []*resource.PathDefinition
    
    for _, def := range allDefinitions {
        hasCDN := false
        hasGCS := false
        
        for providerName := range def.Patterns {
            switch providerName {
            case provider.ProviderCDN:
                hasCDN = true
            case provider.ProviderGCS:
                hasGCS = true
            }
        }
        
        if hasCDN && !hasGCS {
            cdnOnlyDefinitions = append(cdnOnlyDefinitions, def)
        }
        if hasGCS {
            gcsDefinitions = append(gcsDefinitions, def)
        }
    }
    
    fmt.Printf("CDN-only definitions: %d\n", len(cdnOnlyDefinitions))
    fmt.Printf("GCS definitions: %d\n", len(gcsDefinitions))
    
    // Update GCS definitions to add signed URL support
    for _, def := range gcsDefinitions {
        if def.URLOptions == nil || !def.URLOptions.SignedURL {
            // Remove and re-add with signed URL support
            err := manager.RemoveDefinition(def.Name)
            if err != nil {
                log.Printf("Failed to remove %s: %v", def.Name, err)
                continue
            }
            
            updatedDef := *def // Copy
            updatedDef.URLOptions = &provider.URLOptions{
                SignedURL:    true,
                SignedExpiry: 2 * time.Hour,
            }
            
            err = manager.AddDefinition(updatedDef)
            if err != nil {
                log.Printf("Failed to re-add %s: %v", def.Name, err)
                continue
            }
            
            fmt.Printf("Updated %s with signed URL support\n", def.Name)
        }
    }
}
```

## Template Resolution

### Basic Template Patterns

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
    
    result, err := templateResolver.Resolve(ctx, bracesTemplate, params)
    if err != nil {
        log.Printf("Braces template failed: %v", err)
        return
    }
    fmt.Printf("Braces result: %s\n", result)
    
    // Named groups template (regex-style)
    namedTemplate := "/assets/:client_app/:resource_id.:format"
    namedParams := map[resource.ParameterName]string{
        "client_app":   "android",
        "resource_id":  "button_456",
        "format":       "svg",
    }
    
    result, err = templateResolver.Resolve(ctx, namedTemplate, namedParams)
    if err != nil {
        log.Printf("Named template failed: %v", err)
        return
    }
    fmt.Printf("Named result: %s\n", result)
}
```

### Advanced Template Validation

```go
func validateTemplatesSafety(templateResolver resource.TemplateResolver) {
    ctx := context.Background()
    
    // Test path traversal protection
    dangerousTemplates := []string{
        "/files/{resource_id}/../../../etc/passwd",
        "/data/{path}/../../sensitive",
        "/uploads/{filename}\\..\\..\\windows\\system32",
    }
    
    for _, template := range dangerousTemplates {
        err := templateResolver.Validate(ctx, template)
        if err != nil {
            fmt.Printf("Template correctly rejected: %s - %v\n", template, err)
        } else {
            fmt.Printf("WARNING: Dangerous template allowed: %s\n", template)
        }
    }
    
    // Test parameter value validation
    params := map[resource.ParameterName]string{
        "resource_id": "../secret",      // Path traversal in parameter
        "filename":    "test\x00.txt",   // Null byte injection
        "path":        "/etc/passwd",    // Absolute path
    }
    
    safeTemplate := "/files/{resource_id}/{filename}"
    result, err := templateResolver.Resolve(ctx, safeTemplate, params)
    if err != nil {
        fmt.Printf("Parameters correctly rejected: %v\n", err)
    } else {
        fmt.Printf("Resolved with potentially dangerous params: %s\n", result)
    }
}
```

### Custom Template Functions

```go
// Custom parameter resolver with template-level logic
type SmartParameterResolver struct {
    baseResolver resource.ParameterResolver
}

func (r *SmartParameterResolver) Resolve(ctx context.Context, paramName resource.ParameterName) (string, error) {
    // Get base value
    value, err := r.baseResolver.Resolve(ctx, paramName)
    if err != nil {
        return "", err
    }
    
    // Apply template-specific transformations
    switch paramName {
    case "timestamp":
        // Convert to filename-safe format
        if t, err := time.Parse(time.RFC3339, value); err == nil {
            return t.Format("20060102_150405"), nil
        }
    case "resource_id":
        // Sanitize resource IDs for filesystem safety
        sanitized := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(value, "_")
        return sanitized, nil
    case "format":
        // Normalize format extensions
        value = strings.TrimPrefix(value, ".")
        value = strings.ToLower(value)
        return value, nil
    }
    
    return value, nil
}

func useSmartTemplateResolver(manager resource.ResourceManager) {
    ctx := context.Background()
    
    baseResolver := resource.NewValuesParameterResolver(map[resource.ParameterName]string{
        "resource_id": "user@email.com", // Contains unsafe characters
        "timestamp":   time.Now().Format(time.RFC3339),
        "format":      ".PNG", // Wrong case and leading dot
    })
    
    smartResolver := &SmartParameterResolver{baseResolver: baseResolver}
    
    opts := resource.ResolveOptions{
        ParamResolver: smartResolver,
    }
    
    result, err := manager.DefinitionResolver().Resolve(ctx, "user-export", &opts)
    if err != nil {
        log.Printf("Smart resolution failed: %v", err)
        return
    }
    
    fmt.Printf("Sanitized path: %s\n", result.ResolvedPath)
}
```

## Error Handling

### Comprehensive Error Handling

```go
func handleResolutionErrors(manager resource.ResourceManager) {
    ctx := context.Background()
    
    // Test various error conditions
    testCases := []struct {
        name       string
        defName    string
        opts       *resource.ResolveOptions
        expectErr  bool
        errType    error
    }{
        {
            name:      "Definition not found",
            defName:   "nonexistent-definition",
            opts:      &resource.ResolveOptions{},
            expectErr: true,
            errType:   resource.ErrDefinitionNotFound,
        },
        {
            name:    "Missing required parameter",
            defName: "user-avatar",
            opts: resource.ResolveOptions{}.WithValues(map[resource.ParameterName]string{
                // Missing required resource_id
                "format": "png",
            }),
            expectErr: true,
        },
        {
            name:    "Invalid scope",
            defName: "user-avatar",
            opts: func() *resource.ResolveOptions {
                invalidScope := resource.ScopeType("INVALID")
                return &resource.ResolveOptions{
                    Scope: &invalidScope,
                }
            }(),
            expectErr: true,
            errType:   resource.ErrScopeNotAllowed,
        },
        {
            name:    "Provider not supported",
            defName: "user-avatar",
            opts: resource.ResolveOptions{}.
                WithProvider("unsupported-provider").
                WithValues(map[resource.ParameterName]string{
                    resource.ResourceID: "test123",
                }),
            expectErr: true,
        },
    }
    
    for _, tc := range testCases {
        fmt.Printf("Testing: %s\n", tc.name)
        
        result, err := manager.DefinitionResolver().Resolve(ctx, tc.defName, tc.opts)
        
        if tc.expectErr {
            if err == nil {
                fmt.Printf("  ERROR: Expected error but got none\n")
                continue
            }
            
            if tc.errType != nil && !errors.Is(err, tc.errType) {
                fmt.Printf("  ERROR: Expected %v, got %v\n", tc.errType, err)
                continue
            }
            
            fmt.Printf("  SUCCESS: Got expected error: %v\n", err)
        } else {
            if err != nil {
                fmt.Printf("  ERROR: Unexpected error: %v\n", err)
                continue
            }
            
            fmt.Printf("  SUCCESS: Resolved to %s\n", result.URL)
        }
    }
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
    
    // Strategy 2: Scope fallback
    scopes := []*resource.ScopeType{
        func() *resource.ScopeType { s := resource.ScopeClientApp; return &s }(),
        func() *resource.ScopeType { s := resource.ScopeApp; return &s }(),
        func() *resource.ScopeType { s := resource.ScopeGlobal; return &s }(),
    }
    
    for _, scope := range scopes {
        opts := resource.ResolveOptions{
            Scope: scope,
        }.WithValues(map[resource.ParameterName]string{
            resource.ResourceID: resourceID,
        })
        
        result, err := manager.DefinitionResolver().Resolve(ctx, "flexible-asset", &opts)
        if err == nil {
            fmt.Printf("Success with scope %s: %s\n", *scope, result.URL)
            return
        }
        
        fmt.Printf("Scope %s failed: %v\n", *scope, err)
    }
    
    // Strategy 3: Parameter defaults
    fallbackParams := map[resource.ParameterName]string{
        resource.ResourceID:     "fallback_asset",
        resource.ResourceFormat: "jpg",
        resource.Version:        "0.0.1",
    }
    
    opts := resource.ResolveOptions{}.WithValues(fallbackParams)
    result, err := manager.DefinitionResolver().Resolve(ctx, "default-asset", &opts)
    if err == nil {
        fmt.Printf("Fallback success: %s\n", result.URL)
    } else {
        fmt.Printf("All recovery strategies failed: %v\n", err)
    }
}
```

### Structured Error Logging

```go
func logResourceErrors(manager resource.ResourceManager, logger *log.Logger) {
    ctx := context.Background()
    
    opts := resource.ResolveOptions{}.WithValues(map[resource.ParameterName]string{
        // Intentionally invalid parameters
        resource.ResourceID: "",
    })
    
    result, err := manager.DefinitionResolver().Resolve(ctx, "user-avatar", &opts)
    if err != nil {
        // Log structured error information
        var anerr *anerror.Error
        if errors.As(err, &anerr) {
            logger.Printf("Resource resolution failed: code=%d, category=%s, type=%s, message=%s, context=%v",
                anerr.Code(),
                anerr.Category(),
                anerr.Type(),
                anerr.Message(),
                anerr.Context(),
            )
        } else {
            logger.Printf("Resource resolution failed with unexpected error: %v", err)
        }
        return
    }
    
    logger.Printf("Resource resolved successfully: %s", result.URL)
}
```

## Advanced Patterns

### Caching Layer Implementation

```go
type CachedResourceManager struct {
    manager resource.ResourceManager
    cache   map[string]*resource.ResolvedResource
    mutex   sync.RWMutex
    ttl     time.Duration
}

func NewCachedResourceManager(manager resource.ResourceManager, ttl time.Duration) *CachedResourceManager {
    return &CachedResourceManager{
        manager: manager,
        cache:   make(map[string]*resource.ResolvedResource),
        ttl:     ttl,
    }
}

func (c *CachedResourceManager) Resolve(ctx context.Context, defName string, opts *resource.ResolveOptions) (*resource.ResolvedResource, error) {
    // Create cache key
    cacheKey := c.createCacheKey(defName, opts)
    
    // Check cache first
    c.mutex.RLock()
    if cached, exists := c.cache[cacheKey]; exists {
        c.mutex.RUnlock()
        return cached, nil
    }
    c.mutex.RUnlock()
    
    // Resolve using underlying manager
    result, err := c.manager.DefinitionResolver().Resolve(ctx, defName, opts)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    c.mutex.Lock()
    c.cache[cacheKey] = result
    c.mutex.Unlock()
    
    // Schedule cache cleanup
    go func() {
        time.Sleep(c.ttl)
        c.mutex.Lock()
        delete(c.cache, cacheKey)
        c.mutex.Unlock()
    }()
    
    return result, nil
}

func (c *CachedResourceManager) createCacheKey(defName string, opts *resource.ResolveOptions) string {
    // Simple cache key implementation
    parts := []string{defName}
    
    if opts.Provider != nil {
        parts = append(parts, string(*opts.Provider))
    }
    
    if opts.Scope != nil {
        parts = append(parts, string(*opts.Scope))
    }
    
    // Add parameter values
    if opts.ParamResolver != nil {
        // This is simplified - in real implementation, you'd need
        // to extract all parameter values for the cache key
        parts = append(parts, "with_params")
    }
    
    return strings.Join(parts, ":")
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

// Metrics middleware
type MetricsMiddleware struct {
    next    resource.DefinitionResolver
    metrics map[string]int
    mutex   sync.RWMutex
}

func NewMetricsMiddleware() ResolutionMiddleware {
    return func(next resource.DefinitionResolver) resource.DefinitionResolver {
        return &MetricsMiddleware{
            next:    next,
            metrics: make(map[string]int),
        }
    }
}

func (m *MetricsMiddleware) Resolve(ctx context.Context, defName string, opts *resource.ResolveOptions) (*resource.ResolvedResource, error) {
    result, err := m.next.Resolve(ctx, defName, opts)
    
    m.mutex.Lock()
    if err != nil {
        m.metrics[defName+"_errors"]++
    } else {
        m.metrics[defName+"_success"]++
    }
    m.mutex.Unlock()
    
    return result, err
}

func (m *MetricsMiddleware) GetMetrics() map[string]int {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    result := make(map[string]int)
    for k, v := range m.metrics {
        result[k] = v
    }
    return result
}

// Usage
func useMiddleware(baseResolver resource.DefinitionResolver, logger *log.Logger) resource.DefinitionResolver {
    // Chain middleware
    resolver := baseResolver
    resolver = NewMetricsMiddleware()(resolver)
    resolver = NewLoggingMiddleware(logger)(resolver)
    
    return resolver
}
```

### Batch Resource Resolution

```go
func batchResolveResources(manager resource.ResourceManager) {
    ctx := context.Background()
    
    // Define batch requests
    requests := []struct {
        Name   string
        DefName string
        Opts   *resource.ResolveOptions
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
            Name   string
            DefName string
            Opts   *resource.ResolveOptions
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

// Usage
func main() {
    manager := setupResourceManager() // Setup function
    
    http.HandleFunc("/api/resources", createResourceHandler(manager))
    
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
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

### gRPC Service Integration

```go
type ResourceService struct {
    manager resource.ResourceManager
}

func (s *ResourceService) ResolveResource(ctx context.Context, req *pb.ResolveResourceRequest) (*pb.ResolveResourceResponse, error) {
    // Convert parameters
    params := make(map[resource.ParameterName]string)
    for k, v := range req.Parameters {
        params[resource.ParameterName(k)] = v
    }
    
    // Create options
    opts := resource.ResolveOptions{}.WithValues(params)
    
    if req.Provider != "" {
        opts = opts.WithProvider(provider.ProviderName(req.Provider))
    }
    
    if req.UrlOptions != nil {
        opts = opts.WithURLOptions(&provider.URLOptions{
            SignedURL:    req.UrlOptions.SignedUrl,
            SignedExpiry: time.Duration(req.UrlOptions.SignedExpirySeconds) * time.Second,
        })
    }
    
    // Resolve
    result, err := s.manager.DefinitionResolver().Resolve(ctx, req.Definition, &opts)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "Failed to resolve resource: %v", err)
    }
    
    return &pb.ResolveResourceResponse{
        Url:      result.URL,
        Path:     result.ResolvedPath,
        Provider: string(result.Provider),
    }, nil
}
```

## Testing Strategies

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
            name:       "Missing required parameter",
            definition: "user-avatar",
            params:     map[resource.ParameterName]string{},
            expectError: true,
        },
        {
            name:       "Custom format",
            definition: "user-avatar",
            params: map[resource.ParameterName]string{
                resource.ResourceID:     "test456",
                resource.ResourceFormat: "webp",
            },
            expectError: false,
            expectURL:   "https://cdn.example.com/avatars/global/test456.webp",
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

func setupTestResourceManager() resource.ResourceManager {
    // Create mock components
    templateResolver := resource.NewTemplateResolver()
    fallbackResolver := resource.NewValuesParameterResolver(map[resource.ParameterName]string{})
    pathResolver := resource.NewPathResolver(templateResolver, fallbackResolver)
    scopeSelector := resource.NewScopeSelector()
    definitionRegistry := resource.NewPathDefinitionRegistry()
    
    // Create mock provider registry
    providerRegistry := provider.NewRegistry()
    mockCDN := &mockCDNProvider{baseURL: "https://cdn.example.com"}
    providerRegistry.Register(mockCDN)
    providerRegistry.SetDefault(provider.ProviderCDN)
    
    // Create manager
    manager := resource.NewResourceManager(
        pathResolver,
        scopeSelector,
        definitionRegistry,
        providerRegistry,
    )
    
    // Add test definitions
    definition := resource.PathDefinition{
        Name:          "user-avatar",
        AllowedScopes: []resource.ScopeType{resource.ScopeGlobal},
        Patterns: map[provider.ProviderName]resource.PathPatterns{
            provider.ProviderCDN: {
                Patterns: map[resource.ScopeType]string{
                    resource.ScopeGlobal: "/avatars/global/{resource_id}.{format}",
                },
                URLType: resource.URLTypeContent,
            },
        },
        Parameters: []*resource.ParameterDefinition{
            {Name: resource.ResourceID, Required: true},
            {Name: resource.ResourceFormat, DefaultValue: "jpg"},
        },
    }
    
    manager.AddDefinition(definition)
    
    return manager
}

type mockCDNProvider struct {
    baseURL string
}

func (m *mockCDNProvider) Name() provider.ProviderName {
    return provider.ProviderCDN
}

func (m *mockCDNProvider) GenerateURL(path string, opts *provider.URLOptions) (string, error) {
    return m.baseURL + path, nil
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
    
    t.Run("Signed URL Expiry", func(t *testing.T) {
        ctx := context.Background()
        opts := resource.ResolveOptions{
            URLOptions: &provider.URLOptions{
                SignedURL:    true,
                SignedExpiry: 1 * time.Second, // Very short for testing
            },
        }.WithValues(map[resource.ParameterName]string{
            resource.ResourceID: "temp_test_file",
        })
        
        result, err := manager.DefinitionResolver().Resolve(ctx, "temp-file", &opts)
        require.NoError(t, err)
        
        // URL should work immediately
        resp, err := http.Get(result.URL)
        require.NoError(t, err)
        resp.Body.Close()
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        // Wait for expiry
        time.Sleep(2 * time.Second)
        
        // URL should now be expired
        resp, err = http.Get(result.URL)
        require.NoError(t, err)
        resp.Body.Close()
        assert.Equal(t, http.StatusForbidden, resp.StatusCode)
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

func BenchmarkConcurrentResolution(b *testing.B) {
    manager := setupBenchmarkManager()
    ctx := context.Background()
    
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            opts := resource.ResolveOptions{}.WithValues(map[resource.ParameterName]string{
                resource.ResourceID: fmt.Sprintf("resource_%d", i),
            })
            
            _, err := manager.DefinitionResolver().Resolve(ctx, "benchmark-resource", &opts)
            if err != nil {
                b.Fatal(err)
            }
            i++
        }
    })
}
```

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

func (m *OptimizedResourceManager) returnPooledProvider(name provider.ProviderName, p provider.URLProvider) {
    if pool, exists := m.providerPool[name]; exists {
        pool.Put(p)
    }
}
```

### Memory Optimization

```go
type CompactParameterResolver struct {
    values []string
    keys   []resource.ParameterName
}

func NewCompactParameterResolver(params map[resource.ParameterName]string) *CompactParameterResolver {
    resolver := &CompactParameterResolver{
        values: make([]string, 0, len(params)),
        keys:   make([]resource.ParameterName, 0, len(params)),
    }
    
    for k, v := range params {
        resolver.keys = append(resolver.keys, k)
        resolver.values = append(resolver.values, v)
    }
    
    return resolver
}

func (r *CompactParameterResolver) Resolve(ctx context.Context, paramName resource.ParameterName) (string, error) {
    for i, key := range r.keys {
        if key == paramName {
            return r.values[i], nil
        }
    }
    return "", nil
}

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

### Lazy Loading

```go
type LazyPathDefinitionRegistry struct {
    loader func() []*resource.PathDefinition
    cache  []*resource.PathDefinition
    loaded bool
    mutex  sync.RWMutex
}

func NewLazyPathDefinitionRegistry(loader func() []*resource.PathDefinition) *LazyPathDefinitionRegistry {
    return &LazyPathDefinitionRegistry{
        loader: loader,
    }
}

func (r *LazyPathDefinitionRegistry) ensureLoaded() {
    if r.loaded {
        return
    }
    
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    if r.loaded {
        return
    }
    
    r.cache = r.loader()
    r.loaded = true
}

func (r *LazyPathDefinitionRegistry) GetAll() []*resource.PathDefinition {
    r.ensureLoaded()
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    return r.cache
}
```

## Conclusion

This comprehensive guide demonstrates the rich capabilities of the Resource package through practical examples covering:

- **Basic Setup**: Quick start patterns for common use cases
- **Parameter Handling**: Advanced parameter resolution strategies
- **Multi-Provider Support**: Flexible provider selection and fallback
- **Security**: Signed URL generation and template safety
- **Error Handling**: Robust error management and recovery
- **Performance**: Optimization techniques for production use
- **Integration**: Real-world integration patterns
- **Testing**: Comprehensive testing strategies

The Resource package provides a powerful and flexible foundation for resource management in the Aviron Game Server, supporting complex scenarios while maintaining security and performance.