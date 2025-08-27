package core

import (
	"context"
	"time"

	"avironactive.com/resource"
	"avironactive.com/resource/provider"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func AllDefinitions() []*resource.PathDefinition {
	return []*resource.PathDefinition{
		AchievementsPath,
		WorkoutsPath,
	}
}

var (
	AchievementsPathName = resource.PathDefinitionName("achievements")
	AchievementPathName  = resource.PathDefinitionName("achievement")
	WorkoutsPathName     = resource.PathDefinitionName("workouts")
	WorkoutPathName      = resource.PathDefinitionName("workout")

	AchievementsPath = (&resource.PathDefinition{
		Name:          AchievementsPathName,
		DisplayName:   "Achievement Resources",
		Description:   "Achievement icons and metadata shared globally",
		AllowedScopes: []resource.ScopeType{resource.ScopeApp, resource.ScopeGlobal},
		Patterns: map[provider.ProviderName]resource.PathPatterns{
			provider.ProviderCDN: {
				Patterns: map[resource.ScopeType]string{
					resource.ScopeApp:    "/game/{env}/shared/{app}/achievements",
					resource.ScopeGlobal: "/game/{env}/shared/global/achievements",
				},
				URLType: resource.URLTypeContent,
			},
			provider.ProviderR2: {
				Patterns: map[resource.ScopeType]string{
					resource.ScopeApp:    "/aviron-game-assets/{env}/shared/{app}/achievements",
					resource.ScopeGlobal: "/aviron-game-assets/{env}/shared/global/achievements",
				},
				URLType: resource.URLTypeOperation,
			},
		},
		Parameters: []*resource.ParameterDefinition{
			{Name: "app", Description: "Application name (bike, rower) - required for app scope"},
		},
	}).WithChildren(&resource.PathDefinition{
		Name:        AchievementPathName,
		DisplayName: "Achievement Resource",
		Description: "Resource for a specific achievement",
		Patterns: map[provider.ProviderName]resource.PathPatterns{
			provider.ProviderCDN: {
				Patterns: map[resource.ScopeType]string{
					resource.ScopeApp:    "{achievement_id}.{format}",
					resource.ScopeGlobal: "{achievement_id}.{format}",
				},
				URLType: resource.URLTypeContent,
			},
			provider.ProviderR2: {
				Patterns: map[resource.ScopeType]string{
					resource.ScopeApp:    "{achievement_id}.{format}",
					resource.ScopeGlobal: "{achievement_id}.{format}",
				},
				URLType: resource.URLTypeOperation,
			},
		},
		Parameters: []*resource.ParameterDefinition{
			{Name: "achievement_id", Rules: []validation.Rule{validation.Required}, Description: "Achievement identifier"},
			{Name: "format", DefaultValue: "png", Rules: []validation.Rule{validation.Required}, Description: "Image format (png, jpg, svg)"},
			{Name: "app", Description: "Application name (bike, rower) - required for app scope"},
		},
	})

	WorkoutsPath = (&resource.PathDefinition{
		Name:          WorkoutsPathName,
		DisplayName:   "Workout Resources",
		Description:   "Workout files shared globally",
		AllowedScopes: []resource.ScopeType{resource.ScopeApp, resource.ScopeGlobal},
		Patterns: map[provider.ProviderName]resource.PathPatterns{
			provider.ProviderCDN: {
				Patterns: map[resource.ScopeType]string{
					resource.ScopeApp:    "/game/{env}/shared/{app}/workouts",
					resource.ScopeGlobal: "/game/{env}/shared/global/workouts",
				},
				URLType: resource.URLTypeContent,
			},
			provider.ProviderR2: {
				Patterns: map[resource.ScopeType]string{
					resource.ScopeApp:    "/aviron-assets/{env}/shared/{app}/workouts",
					resource.ScopeGlobal: "/aviron-assets/{env}/shared/global/workouts",
				},
				URLType: resource.URLTypeOperation,
			},
		},
		Parameters: []*resource.ParameterDefinition{
			{Name: "app", Description: "Application name (bike, rower) - required for app scope"},
		},
		DefaultStorageMetadata: &resource.StorageMetadataConfig{
			CacheControl: resource.CacheControlConfig{
				MaxAge:      86400 * 7, // 7 days
				AllowPublic: true,
				Default:     "public, max-age=86400",
			},
			RequiredChecksums: []provider.ChecksumAlgorithm{provider.ChecksumAlgorithmSHA256},
			CustomHeaders: map[string]string{
				"x-aviron-private": "true",
			},
		},
	}).WithChildren(&resource.PathDefinition{
		Name:        WorkoutPathName,
		DisplayName: "Workout Resource",
		Description: "Resource for a specific workout file",
		Patterns: map[provider.ProviderName]resource.PathPatterns{
			provider.ProviderCDN: {
				Patterns: map[resource.ScopeType]string{
					resource.ScopeApp:    "{user_id}/{workout_id}.{format}",
					resource.ScopeGlobal: "{user_id}/{workout_id}.{format}",
				},
				URLType: resource.URLTypeContent,
			},
			provider.ProviderR2: {
				Patterns: map[resource.ScopeType]string{
					resource.ScopeApp:    "{user_id}/{workout_id}.{format}",
					resource.ScopeGlobal: "{user_id}/{workout_id}.{format}",
				},
				URLType: resource.URLTypeOperation,
			},
		},
		Parameters: []*resource.ParameterDefinition{
			{Name: "workout_id", Rules: []validation.Rule{validation.Required}, Description: "Workout identifier"},
			{Name: "format", DefaultValue: "json", Rules: []validation.Rule{validation.Required}, Description: "File format (erg, zwo, etc)"},
			{Name: "user_id", DefaultValue: "anonymous", Description: "User ID or 'anonymous' for public workouts"},
		},
	})
)

func NewResourceManager(ctx context.Context) (resource.ResourceManager, error) {
	return resource.NewResourceManager(
		resource.WithProviders(newProviders(ctx)...),
		resource.WithDefinitions(AllDefinitions()...),
		resource.WithFallbackParameterResolver(resource.DefaultFallbackParameterResolver(
			AllClientAppNames(),
			AllAppNames(),
			"dev",
			"v1.0.0",
		)),
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
