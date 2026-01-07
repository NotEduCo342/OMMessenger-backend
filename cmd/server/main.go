package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
	"github.com/noteduco342/OMMessenger-backend/internal/cache"
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

	// Initialize Redis cache
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if parsedDB, err := strconv.Atoi(dbStr); err == nil {
			redisDB = parsedDB
		}
	}

	redisCache := cache.NewRedisCache(redisAddr, redisPassword, redisDB)
	if err := redisCache.Ping(); err != nil {
		log.Printf("WARNING: Redis connection failed: %v. Running without cache.", err)
		redisCache = nil
	} else {
		log.Println("Redis cache connected successfully")
	}

	messageCache := cache.NewMessageCache(redisCache)
	userCache := cache.NewUserCache(redisCache)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	pendingMessageRepo := repository.NewPendingMessageRepository(db)
	versionRepo := repository.NewVersionRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, refreshTokenRepo)
	userService := service.NewUserService(userRepo)
	messageService := service.NewMessageService(messageRepo)
	groupService := service.NewGroupService(groupRepo)
	versionService := service.NewVersionService(versionRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	messageHandler := handlers.NewMessageHandler(messageService, messageCache)
	groupHandler := handlers.NewGroupHandler(groupService)
	versionHandler := handlers.NewVersionHandler(versionService)
	wsHandler := handlers.NewWebSocketHandler(messageService, userService, groupService, pendingMessageRepo, userCache, messageCache)

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

	// Version endpoint (public - no auth required for update checks)
	api.Get("/version", versionHandler.GetVersion)
	api.Get("/version/check", versionHandler.CheckUpdate)

	// Protected routes
	protected := api.Group("/", middleware.AuthRequired(), middleware.CSRFRequired())
	protected.Get("/users/me", userHandler.GetCurrentUser)
	protected.Put("/users/me", userHandler.UpdateProfile)
	protected.Get("/users/search", userHandler.SearchUsers)
	protected.Get("/users/:username", userHandler.GetUserByUsername)
	protected.Get("/messages", messageHandler.GetMessages)
	protected.Post("/messages", messageHandler.SendMessage)

	// Group routes
	protected.Post("/groups", groupHandler.CreateGroup)
	protected.Get("/groups", groupHandler.GetMyGroups)
	protected.Post("/groups/:id/join", groupHandler.JoinGroup)
	protected.Post("/groups/:id/leave", groupHandler.LeaveGroup)
	protected.Get("/groups/:id/members", groupHandler.GetGroupMembers)

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
