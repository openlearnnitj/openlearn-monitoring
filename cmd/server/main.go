package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/openlearnnitj/openlearn-monitoring/internal/config"
	"github.com/openlearnnitj/openlearn-monitoring/internal/handler"
	"github.com/openlearnnitj/openlearn-monitoring/internal/monitoring"
	"github.com/openlearnnitj/openlearn-monitoring/internal/storage"
)

func main() {
	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage client
	storageClient, err := storage.NewDynamoDBClient(cfg.AWSRegion)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB client: %v", err)
	}

	// Initialize monitoring service
	monitoringService := monitoring.NewService(cfg.MonitoringAPIURL, cfg.MonitoringAPISecret)

	// Initialize storage service
	storageService := storage.NewService(storageClient, cfg.DynamoDBTableName)

	// Initialize handler
	h := handler.NewHandler(monitoringService, storageService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "OpenLearn Monitoring Service",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			log.Printf("Error: %v", err)
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"service": "openlearn-monitoring",
		})
	})

	// Trigger monitoring endpoint
	app.Post("/monitor", func(c *fiber.Ctx) error {
		if err := h.Handle(c.Context()); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(fiber.Map{
			"message": "Monitoring completed successfully",
		})
	})

	// Scheduled monitoring endpoint (for external schedulers like cron)
	app.Get("/monitor", func(c *fiber.Ctx) error {
		if err := h.Handle(c.Context()); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(fiber.Map{
			"message": "Monitoring completed successfully",
		})
	})

	// Start server
	port := "3000"
	if envPort := cfg.Port; envPort != "" {
		port = envPort
	}
	
	log.Printf("Starting Fiber server on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
