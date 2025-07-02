package api

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/snowplow-incubator/avalanche-api/pkg/config"
	"github.com/snowplow-incubator/avalanche-api/pkg/repositories"
	"github.com/snowplow-incubator/avalanche-api/pkg/services"
)

type Server struct {
	app            *fiber.App
	config         *config.Config
	profileService services.ProfileServiceInterface
}

func NewServer(cfg *config.Config) *Server {
	app := fiber.New()

	ctx := context.Background()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Failed to parse database URI: %v", err)
	}

	poolSize := cfg.DatabasePoolSize
	if poolSize == 0 {
		poolSize = config.DefaultDatabasePoolSize
	}

	poolConfig.MaxConns = int32(poolSize)
	poolConfig.MinConns = int32(poolSize / 4)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}

	attributeRepo := repositories.NewAttributesRepository(pool)
	profileRepo := repositories.NewProfilesRepository(pool)

	serviceConfig := services.ServiceConfig{
		MaxWorkers: cfg.MaxWorkers,
		BatchSize:  cfg.BatchSize,
	}

	profileService := services.NewProfileService(attributeRepo, profileRepo, cfg.ExtractKeys, serviceConfig)

	server := &Server{
		app:            app,
		config:         cfg,
		profileService: profileService,
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	s.app.Get("/health", s.healthHandler)
	s.app.Post("/events", s.eventsHandler)
}

func (s *Server) healthHandler(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func (s *Server) eventsHandler(c *fiber.Ctx) error {
	var events []map[string]string

	if err := c.BodyParser(&events); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	maxRequestSize := s.config.MaxRequestSize
	if maxRequestSize == 0 {
		maxRequestSize = config.DefaultMaxRequestSize
	}

	if len(events) > maxRequestSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Request too large: %d events (max %d)", len(events), maxRequestSize),
		})
	}

	if len(events) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"results": []any{},
		})
	}

	ctx := context.Background()
	results, err := s.profileService.ProcessEvents(ctx, events)
	if err != nil {
		log.Printf("Failed to process events: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process events",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"results": results,
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	log.Printf("Starting server on %s", addr)
	return s.app.Listen(addr)
}
