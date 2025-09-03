# Documentation: Resource Package

## Overview

The `resource` package provides a unified system for managing and resolving resource paths in the Aviron game server ecosystem. It enables flexible resource path definitions with multi-provider support, parameterized templates, and scope-based resolution. The package handles URL generation for different providers (CDN, Google Cloud Storage) while supporting signed URLs, parameter resolution from context, and secure template processing.

The system is designed around path definitions that can be resolved to actual URLs based on context (scope, provider preferences) and runtime parameters, making it ideal for managing game assets, user data, and configuration files across different environments and storage backends.

## Core Architecture

### Resource Resolution Flow
1. **Path Definition** → Contains patterns and metadata for resource paths
2. **Scope Selection** → Determines the appropriate scope (Global, App, ClientApp) based on context
3. **Parameter Resolution** → Resolves template parameters from various sources (context, values, fallbacks)
4. **Provider Selection** → Chooses storage provider based on URL type and preferences
5. **URL Generation** → Creates final URLs with optional signing and expiry

## API Reference

### Core Types

#### ResourceManager
Central interface for managing resource definitions and resolution.

```go
type ResourceManager interface {
    PathURLResolver() PathURLResolver
    PathDefinitionResolver() PathDefinitionResolver
    
    GetDefinition(name string) (*PathDefinition, error)
    AddDefinition(definition PathDefinition) error
    RemoveDefinition(defName string) error
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

#### PathDefinition
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

#### ResolveOptions
Configuration for resource resolution.

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

#### ResolvedResource
Result of resource resolution.

```go
type ResolvedResource struct {
    URL          string                 // Generated URL
    ResolvedPath string                 // Resolved path template
    Provider     provider.ProviderName  // Selected provider
}
```

### Scope Management

#### ScopeType Constants
```go
const (
    ScopeGlobal    ScopeType = "G"   // Global resources (no context required)
    ScopeApp       ScopeType = "A"   // App-scoped resources (requires app context)
    ScopeClientApp ScopeType = "CA"  // Client app-scoped resources (requires client app context)
)
```

#### ScopeSelector
Determines appropriate scope based on context.

```go
type ScopeSelector interface {
    Select(ctx context.Context, definition *PathDefinition) (ScopeType, error)
    IsAllowed(ctx context.Context, scope ScopeType, definition *PathDefinition) (bool, error)
}

func NewScopeSelector() ScopeSelector
```

### Parameter Resolution

#### ParameterResolver
Interface for resolving template parameters.

```go
type ParameterResolver interface {
    Resolve(ctx context.Context, paramName ParameterName) (string, error)
}

// Built-in parameter names
const (
    ClientApp      ParameterName = "client_app"
    App            ParameterName = "app"
    ResourceID     ParameterName = "resource_id"
    Version        ParameterName = "version"
    Timestamp      ParameterName = "timestamp"
    ResourceFormat ParameterName = "format"
)
```

#### Parameter Resolver Constructors

```go
// Chain multiple resolvers
func NewParameterResolvers(resolvers ...ParameterResolver) (ParameterResolver, error)

// Static values
func NewValuesParameterResolver(params map[ParameterName]string) ParameterResolver

// Context-based resolvers
func NewClientAppParameterResolver(clientAppName map[int16]string) *clientAppParameterResolver
func NewAppParameterResolver(appNames map[int16]string) *appParameterResolver

// Individual value resolvers
func NewValueParameterResolver(paramName ParameterName, value string) *valueParameterResolver
func NewIDParameterResolver(resourceID string) *idParameterResolver
func NewVersionParameterResolver(version string) *versionParameterResolver
func NewTimestampParameterResolver(timestamp time.Time, format string) *timestampParameterResolver

// Routing resolver
func NewRoutingParameterResolver(resolvers map[ParameterName]ParameterResolver) *routingParameterResolver

// Definition-based resolver with validation
func NewDefinitionParameterResolver(definition *ParameterDefinition, resolver ParameterResolver) (ParameterResolver, error)
```

### Template Processing

#### TemplateResolver
Handles template parameter substitution with security validation.

```go
type TemplateResolver interface {
    Resolve(ctx context.Context, template string, params map[ParameterName]string) (string, error)
    ResolveWith(ctx context.Context, template string, resolver ParameterResolver) (string, error)
    Validate(ctx context.Context, template string) error
}

// Constructors
func NewBracesTemplateResolver(opts ...BracesTemplateResolverOption) (TemplateResolver, error)
func NewNamedGroupsTemplateResolver(opts ...RegexTemplateResolverOption) (TemplateResolver, error)
```

#### Template Configuration Options

```go
// Braces template options
func WithBracesTemplateOpeningBrace(openingBrace string) BracesTemplateResolverOption
func WithBracesTemplateClosingBrace(closingBrace string) BracesTemplateResolverOption

// Regex template options
func WithSafetyPatterns(patterns []string) RegexTemplateResolverOption
func WithCacheInterval(interval time.Duration) RegexTemplateResolverOption
```

### Provider System

#### Provider Interface
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

#### URL Options
```go
type URLOptions struct {
    SignedURL    bool          // Generate signed URL
    SignedExpiry time.Duration // Expiry duration for signed URLs
}
```

## Usage Examples

### Basic Resource Definition and Resolution

```go
// Define a resource path
userAvatarDef := PathDefinition{
    Name:          "user-avatars",
    DisplayName:   "User Avatars", 
    Description:   "User avatar images",
    AllowedScopes: []ScopeType{ScopeGlobal, ScopeApp},
    Patterns: map[provider.ProviderName]PathPatterns{
        provider.ProviderCDN: {
            Patterns: map[ScopeType]string{
                ScopeGlobal: "shared/global/users/avatars/{user_id}.png",
                ScopeApp:    "shared/{app}/users/avatars/{user_id}.png",
            },
            URLType: URLTypeContent,
        },
        provider.ProviderGCS: {
            Patterns: map[ScopeType]string{
                ScopeGlobal: "global-avatars/{user_id}.png",
                ScopeApp:    "app-avatars/{app}/{user_id}.png",
            },
            URLType: URLTypeOperation,
        },
    },
    Parameters: []*ParameterDefinition{
        {Name: "user_id", Required: true, Description: "User identifier"},
    },
}

// Create resource manager
registry := NewPathDefinitionRegistry()
registry.Add(userAvatarDef)

providerRegistry := provider.NewRegistry()
// Register providers...

pathResolver := NewPathResolver(templateResolver, fallbackResolver)
scopeSelector := NewScopeSelector()

manager := NewResourceManager(pathResolver, scopeSelector, registry, providerRegistry)

// Resolve resource URL
ctx := context.Background()
opts := ResolveOptions{}.WithValues(map[ParameterName]string{
    "user_id": "123",
})

result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &opts)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("URL: %s\n", result.URL)
fmt.Printf("Path: %s\n", result.ResolvedPath)
fmt.Printf("Provider: %s\n", result.Provider)
```

### Advanced Parameter Resolution

```go
// Create layered parameter resolvers
contextResolver := NewParameterResolvers(
    NewClientAppParameterResolver(map[int16]string{1: "mobile", 2: "web"}),
    NewAppParameterResolver(map[int16]string{1: "bike", 2: "rower"}),
    NewValuesParameterResolver(map[ParameterName]string{
        "user_id": "456",
        "format":  "jpg",
    }),
)

opts := ResolveOptions{
    ParamResolver: contextResolver,
    URLType:       &URLTypeContent,
    URLOptions: &provider.URLOptions{
        SignedURL:    true,
        SignedExpiry: 1 * time.Hour,
    },
}

result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &opts)
```

### Template Processing

```go
// Create template resolver with custom braces
templateResolver, err := NewBracesTemplateResolver(
    WithBracesTemplateOpeningBrace("{{"),
    WithBracesTemplateClosingBrace("}}"),
)
if err != nil {
    log.Fatal(err)
}

// Resolve template
template := "assets/{{client_app}}/{{app}}/images/{{resource_id}}.{{format}}"
params := map[ParameterName]string{
    "client_app":  "mobile",
    "app":         "bike", 
    "resource_id": "avatar_123",
    "format":      "png",
}

resolvedPath, err := templateResolver.Resolve(ctx, template, params)
// Result: "assets/mobile/bike/images/avatar_123.png"
```

### Provider Selection and URL Generation

```go
// Resolve with specific provider
opts := ResolveOptions{}.
    WithProvider(provider.ProviderGCS).
    WithURLType(URLTypeOperation).
    WithValues(map[ParameterName]string{"user_id": "789"})

result, err := manager.PathDefinitionResolver().Resolve(ctx, "user-avatars", &opts)

// Direct URL resolution
pathResolver := manager.PathURLResolver() 
resolvedPath := "users/avatars/789.png"

urlResult, err := pathResolver.Resolve(ctx, resolvedPath, &ResolveOptions{
    Provider: &provider.ProviderCDN,
    URLOptions: &provider.URLOptions{
        SignedURL:    true,
        SignedExpiry: 30 * time.Minute,
    },
})
```

## Error Handling

The package defines specific error types for different failure scenarios:

```go
var (
    ErrDefinitionNotFound       = anerror.NewNotFoundError(8001, ...)
    ErrDefinitionExists         = anerror.NewInvalidArgumentError(8002, ...)
    ErrInvalidDefinition        = anerror.NewInvalidArgumentError(8003, ...)
    ErrTemplateResolutionFailed = anerror.NewInternalError(8004, ...)
    ErrConditionNotMet          = anerror.NewInvalidArgumentError(8005, ...)
    ErrScopeNotAllowed          = anerror.NewInvalidArgumentError(8006, ...)
    ErrInvalidTemplate          = anerror.NewInvalidArgumentError(8008, ...)
)
```

Use `errors.Is()` or `anerror.As()` for error type checking:

```go
if errors.Is(err, ErrDefinitionNotFound) {
    // Handle missing definition
}

if anerror.As(err).IsInvalidArgument() {
    // Handle validation errors
}
```

## Security Considerations

The package includes multiple security features:

1. **Path Traversal Protection**: Templates are validated against directory traversal patterns (`../`, etc.)
2. **Parameter Validation**: Parameter values are checked for dangerous characters (`/`, `\`, null bytes)
3. **Template Safety**: Built-in patterns prevent access to system directories (`/etc/`, `/proc/`, etc.)
4. **Length Limits**: Maximum lengths enforced for templates (2000 chars) and parameter values (500 chars)
5. **Path Normalization**: Final paths are cleaned and must be absolute

## Performance Considerations

- **Template Caching**: Compiled templates are cached with configurable intervals
- **Registry Optimization**: Path definitions are cached for O(1) lookup by name
- **Concurrent Safety**: All components are thread-safe with read-write mutexes
- **Provider Selection**: Cached provider lookups by URL type for O(1) access
- **Parameter Resolution**: Chained resolvers with short-circuiting for efficiency

## Testing

The package includes comprehensive test coverage:

- **Unit Tests**: `*_test.go` files test individual components
- **Integration Tests**: `*_integration_test.go` test complete workflows
- **Test Fixtures**: `tests/path_definitions.go` provides example definitions
- **Mock Support**: Test helpers for mocking interfaces

Run tests with:
```bash
go test -v ./resource/...           # Unit tests
go test -v -integration ./resource/... # Integration tests
```

## Related Components

- **Provider Package** (`resource/provider`): Storage and URL generation backends
- **Common/ANerrror**: Structured error handling
- **Common/Context**: Request context management
- **Common/App**: Application context extraction
- **Common/UserAgent**: Client application identification