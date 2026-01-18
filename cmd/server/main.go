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
	"github.com/noteduco342/OMMessenger-backend/internal/httpx"
	"github.com/noteduco342/OMMessenger-backend/internal/middleware"
	"github.com/noteduco342/OMMessenger-backend/internal/repository"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
	"github.com/noteduco342/OMMessenger-backend/internal/storage"
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
		AppName: "OM Messenger Backend",
		// Support avatar uploads up to 5MB + overhead.
		BodyLimit: 8 * 1024 * 1024, // 8MB
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
	groupInviteRepo := repository.NewGroupInviteRepository(db)
	groupReadStateRepo := repository.NewGroupReadStateRepository(db)
	pendingMessageRepo := repository.NewPendingMessageRepository(db)
	versionRepo := repository.NewVersionRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, refreshTokenRepo, groupRepo)
	userService := service.NewUserService(userRepo, groupRepo)
	messageService := service.NewMessageService(messageRepo)
	groupService := service.NewGroupService(groupRepo, groupReadStateRepo, userRepo, groupInviteRepo)
	versionService := service.NewVersionService(versionRepo)

	// Initialize S3/MinIO storage (best-effort; feature endpoints return 503 if missing)
	var s3Store *storage.S3Storage
	if cfg, err := storage.LoadS3ConfigFromEnv(); err != nil {
		log.Printf("WARNING: S3 storage not configured: %v", err)
	} else if st, err := storage.NewS3Storage(cfg); err != nil {
		log.Printf("WARNING: Failed to initialize S3 storage: %v", err)
	} else {
		s3Store = st
		log.Printf("S3 storage initialized successfully (bucket=%s)", cfg.Bucket)
	}

	avatarService := service.NewAvatarService(userRepo, s3Store)

	// Initialize handlers
	wsHandler := handlers.NewWebSocketHandler(messageService, userService, groupService, pendingMessageRepo, userCache, messageCache)
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	avatarHandler := handlers.NewAvatarHandler(avatarService)
	mediaHandler := handlers.NewMediaHandler(s3Store)
	messageHandler := handlers.NewMessageHandler(messageService, groupService, messageCache, wsHandler.GetHub())
	groupHandler := handlers.NewGroupHandler(groupService)
	versionHandler := handlers.NewVersionHandler(versionService)

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
	api.Get("/join/:token", groupHandler.GetInvitePreview)

	// Protected routes
	protected := api.Group("/", middleware.AuthRequired(), middleware.CSRFRequired())
	protected.Get("/users/me", userHandler.GetCurrentUser)
	protected.Put("/users/me", userHandler.UpdateProfile)
	protected.Post(
		"/users/me/avatar",
		limiter.New(limiter.Config{
			Max:        10,
			Expiration: 10 * time.Minute,
			KeyGenerator: func(c *fiber.Ctx) string {
				if uid, err := httpx.LocalUint(c, "userID"); err == nil {
					return "avatar:" + strconv.FormatUint(uint64(uid), 10)
				}
				return c.IP()
			},
		}),
		avatarHandler.UploadMyAvatar,
	)
	protected.Delete("/users/me/avatar", avatarHandler.DeleteMyAvatar)
	protected.Get("/media/avatars/*", mediaHandler.GetAvatar)
	protected.Get("/users/search", userHandler.SearchUsers)
	protected.Get("/users/:identifier", userHandler.GetUser)
	protected.Get("/conversations", messageHandler.GetConversations)
	protected.Post("/conversations/:peer_id/read", messageHandler.MarkConversationRead)
	protected.Get("/messages", messageHandler.GetMessages)
	protected.Post("/messages", messageHandler.SendMessage)
	protected.Post("/messages/sync", messageHandler.SyncMessages)

	// Group routes
	protected.Post("/groups", groupHandler.CreateGroup)
	protected.Get("/groups", groupHandler.GetMyGroups)
	protected.Get("/groups/public/search", groupHandler.SearchPublicGroups)
	protected.Get("/groups/handle/:handle", groupHandler.GetPublicGroupByHandle)
	protected.Post("/groups/handle/:handle/join", groupHandler.JoinPublicGroupByHandle)
	protected.Post("/groups/:id/join", groupHandler.JoinGroup)
	protected.Post("/groups/:id/leave", groupHandler.LeaveGroup)
	protected.Get("/groups/:id/members", groupHandler.GetGroupMembers)
	protected.Post("/groups/:id/invite-links", groupHandler.CreateInviteLink)
	protected.Post("/join/:token", groupHandler.JoinByInviteLink)
	protected.Get("/groups/:id/messages", messageHandler.GetGroupMessages)
	protected.Post("/groups/:id/messages", messageHandler.SendGroupMessage)
	protected.Post("/groups/:id/read", messageHandler.MarkGroupRead)
	protected.Get("/groups/:id/read-state", messageHandler.GetGroupReadState)

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
