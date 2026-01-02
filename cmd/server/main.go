package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
	"github.com/noteduco342/OMMessenger-backend/internal/handlers"
	"github.com/noteduco342/OMMessenger-backend/internal/middleware"
	"github.com/noteduco342/OMMessenger-backend/internal/repository"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:   "OM Messenger Backend",
		BodyLimit: 1 * 1024 * 1024, // 1MB
	})

	// Middleware
	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("ALLOWED_ORIGINS"),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-OM-CSRF",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Initialize database connection
	db, err := repository.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, refreshTokenRepo)
	userService := service.NewUserService(userRepo)
	messageService := service.NewMessageService(messageRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	messageHandler := handlers.NewMessageHandler(messageService)
	wsHandler := handlers.NewWebSocketHandler(messageService, userService)

	// Public routes
	api := app.Group("/api", middleware.OriginAllowed())
	auth := api.Group("/auth", limiter.New(limiter.Config{
		Max:        20,
		Expiration: time.Minute,
	}))
	auth.Get("/csrf", authHandler.CSRF)
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh) // No CSRF required - protected by HttpOnly refresh token
	auth.Post("/logout", middleware.CSRFRequired(), authHandler.Logout)
	api.Get("/users/check-username", userHandler.CheckUsername) // Public endpoint for username check

	// Protected routes
	protected := api.Group("/", middleware.AuthRequired(), middleware.CSRFRequired())
	protected.Get("/users/me", userHandler.GetCurrentUser)
	protected.Put("/users/me", userHandler.UpdateProfile)
	protected.Get("/users/search", userHandler.SearchUsers)
	protected.Get("/users/:username", userHandler.GetUserByUsername)
	protected.Get("/messages", messageHandler.GetMessages)
	protected.Post("/messages", messageHandler.SendMessage)

	// WebSocket route (websocket upgrade needs special handling)
	app.Use(
		"/ws",
		middleware.OriginAllowed(),
		middleware.AuthRequired(),
		func(c *fiber.Ctx) error {
			// Upgrade to WebSocket
			if websocket.IsWebSocketUpgrade(c) {
				return c.Next()
			}
			return fiber.ErrUpgradeRequired
		},
	)
	app.Get("/ws", websocket.New(wsHandler.HandleWebSocket))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "OM Messenger is running",
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
