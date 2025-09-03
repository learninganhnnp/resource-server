# PathDefinition → ResourcePolicy Rename Strategy

## Overview
Rename `PathDefinition` to `ResourcePolicy` to better reflect its actual business purpose as a comprehensive resource access and storage policy rather than just path configuration.

## Business Justification
After analyzing the `core` package and existing `PathDefinition` usage, it's clear that these entities are comprehensive **business resource policies** that define:
- Access control (allowed scopes)
- Storage policies (metadata, cache control, checksums)
- Security policies (URL expiry, custom headers)
- Multi-provider routing strategies
- Business rules for resource management

The current name `PathDefinition` suggests only URL/path concerns, but the actual implementation handles comprehensive resource management policies.

## Impact Analysis
- **Total occurrences found:** 293 PathDefinition references across 16 files
- **DefinitionName occurrences:** 52 across 8 files  
- **DefinitionResolver occurrences:** 39 across 8 files
- **PathDefinitionRegistry occurrences:** 78 across 7 files

## Phase 1: Core Type Renaming

### 1.1 Primary Types (resolver/definition.go)
```go
// Before → After
PathDefinition → ResourcePolicy
DefinitionName → ResourcePolicyName (or PolicyName)
```

### 1.2 Interface Types
```go  
// Before → After
DefinitionResolver → ResourcePolicyResolver
PathDefinitionRegistry → ResourcePolicyRegistry
```

### 1.3 Variable/Field Names
```go
// Before → After
pathDefinition → resourcePolicy
DefinitionResolver → resourcePolicyResolver  
definitionRegistry → policyRegistry
```

## Phase 2: File-by-File Migration Plan

### High Priority Files (Core Domain)
1. **resolver/definition.go** - Core aggregate definition
2. **resolver/definition_registry.go** - Registry interface/implementation
3. **resolver/resolver.go** - Resolver interfaces and implementations
4. **resource.go** - ResourceManager interface

### Medium Priority Files (Services)
5. **upload/manager_base.go** - Upload manager base
6. **upload/manager_simple.go** - Simple upload manager
7. **upload/manager_multipart.go** - Multipart upload service  
9. **options.go** - ResourceManager options

### Low Priority Files (Models & Tests)
10. **upload/upload.go** - Upload aggregate
11. **upload/models.go** - Upload models
12. **resolver/*_test.go** - Test files

## Phase 3: External Package Updates

### Core Package (Business Configuration)
```go
// core/policies.go - Update business definitions
var (
    AchievementsPolicy = &resource.ResourcePolicy{  // renamed from AchievementsPath
        Name: resource.PolicyName("achievements"),   // renamed from DefinitionName
        DisplayName: "Achievement Resources",
        Description: "Achievement icons and metadata shared globally",
        AllowedScopes: []resource.ScopeType{resource.ScopeApp, resource.ScopeGlobal},
        // ... rest of policy definition
    }
    
    WorkoutsPolicy = &resource.ResourcePolicy{      // renamed from WorkoutsPath
        Name: resource.PolicyName("workouts"),
        // ... policy definition
    }
)

func AllResourcePolicies() []*resource.ResourcePolicy {  // renamed from AllDefinitions
    return []*resource.ResourcePolicy{
        AchievementsPolicy,
        WorkoutsPolicy,
    }
}

func NewResourceManager(ctx context.Context) (resource.ResourceManager, error) {
    return resource.NewResourceManager(
        resource.WithProviders(newProviders(ctx)...),
        resource.WithPolicies(AllResourcePolicies()...), // renamed from WithDefinitions
    )
}
```

## Phase 4: Documentation Updates

### 4.1 Comments and Documentation
- Update all code comments referencing "path definition"
- Update struct documentation to reflect policy nature
- Update README and implementation docs

### 4.2 Ubiquitous Language Updates
```go
// Before: "Path Definition defines URL patterns for resource paths"
// After: "Resource Policy defines access rules, storage policies, security constraints, and URL patterns for business resources"
```

### 4.3 Method Names and Function Updates
```go
// Before → After
GetDefinition() → GetPolicy()
AddDefinition() → AddPolicy()
RemoveDefinition() → RemovePolicy()
GetAllDefinitions() → GetAllPolicies()
WithDefinitions() → WithPolicies()
```

## Phase 5: Testing Strategy

### 5.1 Automated Verification
- Run `go build ./...` after each file migration
- Run `go test ./...` to verify all tests pass
- Check for any missed references with grep

### 5.2 Integration Testing
- Test ResourceManager creation with new types
- Verify upload services work with renamed types
- Test policy resolution workflows
- Verify external core package integration

## Implementation Order

### Step 1: Foundation (Single Commit)
```bash
# Files to update in order:
1. resolver/definition.go - Rename core types and methods
2. resolver/definition_registry.go - Update registry interface/impl
3. resolver/resolver.go - Update resolver interfaces
```

### Step 2: Services (Single Commit)
```bash
# Files to update:
1. resource.go - Update ResourceManager interface
2. options.go - Update configuration options
3. upload/service_base.go - Update service dependencies
4. upload/service_simple.go - Update service references
5. upload/service_multipart.go - Update service references
6. upload/service_facade.go - Update facade references
```

### Step 3: Models & Tests (Single Commit)
```bash
# Files to update:
1. upload/upload.go - Update upload aggregate references
2. upload/models.go - Update model references
3. All *_test.go files - Update test code
```

### Step 4: External Packages (Separate Commit)
```bash
# Update consuming packages:
1. core package - Update business policy definitions
2. Any other packages importing resource types
```

## Pre-Migration Checklist
- [ ] Create feature branch: `git checkout -b refactor/rename-pathdefinition-to-resourcepolicy`
- [ ] Run full test suite to ensure clean starting state
- [ ] Document current test coverage baseline
- [ ] Create this strategy document for reference

## Post-Migration Verification
- [ ] All tests pass: `go test ./...`
- [ ] All packages build: `go build ./...`
- [ ] No compilation errors or warnings
- [ ] Grep verification for missed references:
  ```bash
  grep -r "PathDefinition" . --exclude-dir=.git --exclude="*.md"
  grep -r "pathDefinition" . --exclude-dir=.git --exclude="*.md"
  ```
- [ ] Integration test with core package
- [ ] Verify business logic unchanged

## Rollback Strategy
- Each phase is a separate commit with descriptive commit messages
- Can rollback individual phases if issues arise: `git revert <commit-hash>`
- Maintain clear commit history for easy navigation
- Test after each phase to catch issues early

## Success Criteria
- [ ] All tests pass without modification of test logic
- [ ] No compilation errors or import issues
- [ ] Business functionality remains identical
- [ ] Clear, consistent naming throughout codebase
- [ ] Documentation reflects new domain language
- [ ] External packages (core) integrate properly
- [ ] Team can understand new naming convention

## Domain Impact
This rename better aligns the codebase with Domain-Driven Design principles:

- **Before**: Technical focus on "path definitions"
- **After**: Business focus on "resource policies"

The renamed `ResourcePolicy` clearly communicates that this aggregate handles:
1. **Business Access Rules** - Who can access resources
2. **Storage Policies** - How resources are stored and cached
3. **Security Policies** - URL expiry, encryption, custom headers
4. **Multi-Provider Strategy** - How resources route across CDN/storage
5. **URL Resolution** - Technical implementation detail

This creates clearer domain boundaries and makes the codebase more maintainable for future business requirements.