package http

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"avironactive.com/resource"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/anh-nguyen/resource-server/core"
	"github.com/anh-nguyen/resource-server/internal/app/usecases"
	"github.com/anh-nguyen/resource-server/internal/infrastructure/config"
	"github.com/anh-nguyen/resource-server/internal/interfaces/http/handlers"
	"github.com/anh-nguyen/resource-server/internal/interfaces/http/middleware"
)

type Server struct {
	app             *fiber.App
	config          *config.Config
	resourceManager resource.ResourceManager
}

func NewServer(cfg *config.Config) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		ErrorHandler: middleware.ErrorHandler,
	})

	app.Use(recover.New())

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: joinStrings(cfg.CORS.AllowedOrigins, ","),
		AllowMethods: joinStrings(cfg.CORS.AllowedMethods, ","),
		AllowHeaders: joinStrings(cfg.CORS.AllowedHeaders, ","),
	}))

	resourceManager, err := core.NewResourceManager(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize resource manager: %v", err)
	}
	return &Server{
		app:             app,
		config:          cfg,
		resourceManager: resourceManager,
	}
}

func (s *Server) SetupRoutes() {
	api := s.app.Group("/api/v1")

	// Health check
	healthHandler := handlers.NewHealthHandler()
	api.Get("/health", healthHandler.HealthCheck)

	resourceDefinitionUseCase := usecases.NewResourceDefinitionUseCase(s.resourceManager)
	resourceDefinitionHandler := handlers.NewResourceDefinitionHandler(resourceDefinitionUseCase)

	providerUseCase := usecases.NewProviderUseCase(s.resourceManager)
	providerHandler := handlers.NewProviderHandler(providerUseCase)

	fileOperationsUseCase := usecases.NewFileOperationsUseCase(s.resourceManager)
	fileOperationsHandler := handlers.NewFileOperationsHandler(fileOperationsUseCase)

	// Resources group
	resources := api.Group("/resources")

	// Resource definition routes
	resources.Get("/definitions", resourceDefinitionHandler.ListDefinitions)
	resources.Get("/definitions/:name", resourceDefinitionHandler.GetDefinition)

	// Provider routes
	resources.Get("/providers", providerHandler.ListProviders)
	resources.Get("/providers/:name", providerHandler.GetProvider)

	// File operation routes
	resources.Get("/:provider/:definition", fileOperationsHandler.ListFiles)
	resources.Post("/:provider/:definition/upload", fileOperationsHandler.GenerateUploadURL)
	resources.Post("/:provider/:definition/upload/multipart", fileOperationsHandler.GenerateMultipartUploadURLs)
	resources.Post("/:provider/*/download", fileOperationsHandler.GenerateDownloadURL)
	resources.Get("/:provider/*/metadata", fileOperationsHandler.GetFileMetadata)
	resources.Put("/:provider/*/metadata", fileOperationsHandler.UpdateFileMetadata)
	resources.Delete("/:provider/*", fileOperationsHandler.DeleteFile)
}

func (s *Server) Start() error {
	s.SetupRoutes()

	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)

	go func() {
		if err := s.app.Listen(addr); err != nil {
			log.Printf("Server failed to start: %v", err)
		}
	}()

	log.Printf("Server started on %s", addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := s.Shutdown(); err != nil {
		return err
	}

	log.Println("Server exited")
	return nil
}

func (s *Server) Shutdown() error {
	log.Println("Shutting down server...")

	if err := s.resourceManager.Close(); err != nil {
		return fmt.Errorf("failed to close resource manager: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	return nil
}

func joinStrings(slice []string, separator string) string {
	if len(slice) == 0 {
		return ""
	}

	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += separator + slice[i]
	}
	return result
}
