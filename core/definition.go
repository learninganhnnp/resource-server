package core

import (
	"context"
	"time"

	"avironactive.com/resource"
	"avironactive.com/resource/metadata"
	"avironactive.com/resource/provider"
	"avironactive.com/resource/resolver"
	"avironactive.com/resource/upload"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AllDefinitions() []*resolver.Definition {
	return []*resolver.Definition{
		AchievementsPath,
		WorkoutsPath,
	}
}

var (
	AchievementsPathName = resolver.DefinitionName("achievements")
	AchievementPathName  = resolver.DefinitionName("achievement")
	WorkoutsPathName     = resolver.DefinitionName("workouts")
	WorkoutPathName      = resolver.DefinitionName("workout")

	AchievementsPath = (&resolver.Definition{
		Name:          AchievementsPathName,
		DisplayName:   "Achievement Resources",
		Description:   "Achievement icons and metadata shared globally",
		AllowedScopes: []resolver.ScopeType{resolver.ScopeApp, resolver.ScopeGlobal},
		Patterns: map[provider.ProviderName]resolver.PathPatterns{
			provider.ProviderCDN: {
				Patterns: map[resolver.ScopeType]string{
					resolver.ScopeApp:    "/game/{env}/shared/{app}/achievements",
					resolver.ScopeGlobal: "/game/{env}/shared/global/achievements",
				},
				URLType: resolver.URLTypeDelivery,
			},
			provider.ProviderR2: {
				Patterns: map[resolver.ScopeType]string{
					resolver.ScopeApp:    "/aviron-game-assets/{env}/shared/{app}/achievements",
					resolver.ScopeGlobal: "/aviron-game-assets/{env}/shared/global/achievements",
				},
				URLType: resolver.URLTypeStorage,
			},
		},
		Parameters: []*resolver.ParameterDefinition{
			{Name: "app", Description: "Application name (bike, rower) - required for app scope"},
		},
	}).WithChildren(&resolver.Definition{
		Name:        AchievementPathName,
		DisplayName: "Achievement Resource",
		Description: "Resource for a specific achievement",
		Patterns: map[provider.ProviderName]resolver.PathPatterns{
			provider.ProviderCDN: {
				Patterns: map[resolver.ScopeType]string{
					resolver.ScopeApp:    "{achievement_id}.{format}",
					resolver.ScopeGlobal: "{achievement_id}.{format}",
				},
				URLType: resolver.URLTypeDelivery,
			},
			provider.ProviderR2: {
				Patterns: map[resolver.ScopeType]string{
					resolver.ScopeApp:    "{achievement_id}.{format}",
					resolver.ScopeGlobal: "{achievement_id}.{format}",
				},
				URLType: resolver.URLTypeStorage,
			},
		},
		Parameters: []*resolver.ParameterDefinition{
			{Name: "achievement_id", Rules: []validation.Rule{validation.Required}, Description: "Achievement identifier"},
			{Name: "format", DefaultValue: "png", Rules: []validation.Rule{validation.Required}, Description: "Image format (png, jpg, svg)"},
			{Name: "app", Description: "Application name (bike, rower) - required for app scope"},
		},
	})

	WorkoutsPath = (&resolver.Definition{
		Name:          WorkoutsPathName,
		DisplayName:   "Workout Resources",
		Description:   "Workout files shared globally",
		AllowedScopes: []resolver.ScopeType{resolver.ScopeApp, resolver.ScopeGlobal},
		Patterns: map[provider.ProviderName]resolver.PathPatterns{
			provider.ProviderCDN: {
				Patterns: map[resolver.ScopeType]string{
					resolver.ScopeApp:    "/game/{env}/shared/{app}/workouts",
					resolver.ScopeGlobal: "/game/{env}/shared/global/workouts",
				},
				URLType: resolver.URLTypeDelivery,
			},
			provider.ProviderR2: {
				Patterns: map[resolver.ScopeType]string{
					resolver.ScopeApp:    "/aviron-assets/{env}/shared/{app}/workouts",
					resolver.ScopeGlobal: "/aviron-assets/{env}/shared/global/workouts",
				},
				URLType: resolver.URLTypeStorage,
			},
		},
		Parameters: []*resolver.ParameterDefinition{
			{Name: "app", Description: "Application name (bike, rower) - required for app scope"},
		},
		DefaultStorageMetadata: &metadata.StorageMetadataConfig{
			CacheControl: metadata.CacheControlConfig{
				MaxAge:      86400 * 7, // 7 days
				AllowPublic: true,
				Default:     "public, max-age=86400",
			},
			RequiredChecksums: []metadata.ChecksumAlgorithm{metadata.ChecksumAlgorithmSHA256},
			CustomHeaders: map[string]string{
				"x-aviron-private": "true",
			},
		},
	}).WithChildren(&resolver.Definition{
		Name:        WorkoutPathName,
		DisplayName: "Workout Resource",
		Description: "Resource for a specific workout file",
		Patterns: map[provider.ProviderName]resolver.PathPatterns{
			provider.ProviderCDN: {
				Patterns: map[resolver.ScopeType]string{
					resolver.ScopeApp:    "{user_id}/{workout_id}.{format}",
					resolver.ScopeGlobal: "{user_id}/{workout_id}.{format}",
				},
				URLType: resolver.URLTypeDelivery,
			},
			provider.ProviderR2: {
				Patterns: map[resolver.ScopeType]string{
					resolver.ScopeApp:    "{user_id}/{workout_id}.{format}",
					resolver.ScopeGlobal: "{user_id}/{workout_id}.{format}",
				},
				URLType: resolver.URLTypeStorage,
			},
		},
		Parameters: []*resolver.ParameterDefinition{
			{Name: "workout_id", Rules: []validation.Rule{validation.Required}, Description: "Workout identifier"},
			{Name: "format", DefaultValue: "json", Rules: []validation.Rule{validation.Required}, Description: "File format (erg, zwo, etc)"},
			{Name: "user_id", DefaultValue: "anonymous", Description: "User ID or 'anonymous' for public workouts"},
		},
	})
)

func NewResourceManager(ctx context.Context, pgxConn *pgxpool.Pool) (resource.ResourceManager, error) {
	return resource.NewResourceManager(
		resource.WithProviders(newProviders(ctx)...),
		resource.WithDefinitions(AllDefinitions()...),
		resource.WithFallbackParameterResolver(resolver.DefaultFallbackParameterResolver(
			AllClientAppNames(),
			AllAppNames(),
			"dev",
			"v1.0.0",
		)),
		resource.WithUploadRepository(upload.NewRepository(pgxConn)),
	)
}

// TODO: read from config
func newProviders(ctx context.Context) []provider.Provider {
	cdnProvider := provider.NewCDNProvider(provider.CDNConfig{
		BaseURL:    "https://cdn-dev.avironactive.net/assets",
		SigningKey: "your-signing-key",
		Expiry:     time.Hour * 24, // 24 hours
	})

	gcpProvider, err := provider.NewGCSProvider(ctx, provider.GCSConfig{
		Expiry: time.Hour * 24, // 24 hours
	})
	if err != nil {
		panic("Failed to create GCS provider: " + err.Error())
	}

	r2Provider, err := provider.NewR2Provider(ctx, provider.R2Config{
		AccountID:   "6bca35be0dfe6967897b13de521653e4",
		AccessKeyID: "2c6e52887d5119ceb3711a6a30808348",
		SecretKey:   "ce7c5ec1f84254a6768c246da70d7ebdd0351fd498b96dc83b33dd512c6596b4",
		Expiry:      time.Hour * 24, // 24 hours
	})
	if err != nil {
		panic("Failed to create R2 provider: " + err.Error())
	}

	return []provider.Provider{
		cdnProvider,
		gcpProvider,
		r2Provider,
	}
}

var (
	clientAppNames = map[int16]string{
		1: "unity",
		2: "unity",
		3: "mobile",
		4: "admin",
	}
)

func AllClientAppNames() map[int16]string {
	results := make(map[int16]string, len(clientAppNames))
	for id, name := range clientAppNames {
		results[int16(id)] = name
	}
	return results
}

var (
	appNames = map[int16]string{
		1: "rower",
		2: "c2rower",
		3: "c2ski",
		4: "c2bike",
	}
)

func AllAppNames() map[int16]string {
	results := make(map[int16]string, len(appNames))
	for id, name := range appNames {
		results[int16(id)] = name
	}
	return results
}
