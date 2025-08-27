package main

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/openlearnnitj/openlearn-monitoring/internal/config"
	"github.com/openlearnnitj/openlearn-monitoring/internal/status"
	"github.com/openlearnnitj/openlearn-monitoring/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize DynamoDB client
	dynamoClient, err := storage.NewDynamoDBClient(cfg.AWSRegion)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB client: %v", err)
	}

	// Initialize status service
	statusService := status.NewStatusService(dynamoClient.GetClient(), cfg.DynamoDBTableName)

	// Initialize template engine with custom functions
	engine := html.New("./web/templates", ".html")
	engine.AddFunc("lower", func(s string) string {
		return strings.ToLower(s)
	})

	// Create Fiber app
	app := fiber.New(fiber.Config{
		Views:   engine,
		AppName: "OpenLearn Status Page",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			log.Printf("Error: %v", err)
			return c.Status(code).SendString("Internal Server Error")
		},
	})

	// Middleware
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		systemStatus, err := statusService.GetCurrentStatus(c.Context())
		if err != nil {
			log.Printf("Failed to get status: %v", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to load status")
		}

		return c.Render("status", systemStatus)
	})

	// API endpoint for JSON status (for external integrations)
	app.Get("/api/status", func(c *fiber.Ctx) error {
		systemStatus, err := statusService.GetCurrentStatus(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to load status",
			})
		}

		return c.JSON(systemStatus)
	})

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "openlearn-status-page",
		})
	})

	// Static files (if any)
	app.Static("/static", "./web/static")

	// Get port from environment or use default
	port := "8080"
	if cfg.Port != "" {
		port = cfg.Port
	}

	log.Printf("Starting OpenLearn Status Page on port %s", port)
	log.Printf("Visit: http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
