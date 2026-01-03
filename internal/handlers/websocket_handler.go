package handlers

import (
	"log"

	"github.com/gofiber/websocket/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/cache"
	"github.com/noteduco342/OMMessenger-backend/internal/handlers/ws"
	"github.com/noteduco342/OMMessenger-backend/internal/repository"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
)

type WebSocketHandler struct {
	messageService *service.MessageService
	userService    *service.UserService
	groupService   *service.GroupService
	hub            *ws.Hub
	userCache      *cache.UserCache
	messageCache   *cache.MessageCache
}

func NewWebSocketHandler(messageService *service.MessageService, userService *service.UserService, groupService *service.GroupService, pendingRepo repository.PendingMessageRepositoryInterface, userCache *cache.UserCache, messageCache *cache.MessageCache) *WebSocketHandler {
	return &WebSocketHandler{
		messageService: messageService,
		userService:    userService,
		groupService:   groupService,
		hub:            ws.NewHub(pendingRepo),
		userCache:      userCache,
		messageCache:   messageCache,
	}
}

// GetHub returns the hub instance (useful for sending messages from other handlers)
func (h *WebSocketHandler) GetHub() *ws.Hub {
	return h.hub
}

func (h *WebSocketHandler) HandleWebSocket(c *websocket.Conn) {
	userID := c.Locals("userID").(uint)

	// Check if client supports gzip compression (via query param or header)
	supportsGzip := c.Query("gzip") == "1" || c.Headers("X-Supports-Gzip") == "1"

	// Register client in hub
	h.hub.Register(userID, c, supportsGzip)

	// Update user status to online
	go func() {
		if h.userCache != nil {
			if err := h.userCache.SetUserOnline(userID); err != nil {
				log.Printf("Failed to set user %d online in cache: %v", userID, err)
			}
		}
		if err := h.userService.SetUserOnline(userID); err != nil {
			log.Printf("Failed to set user %d online in DB: %v", userID, err)
		}
	}()

	// Flush pending messages after successful connection
	go func() {
		if err := h.hub.FlushPendingMessages(userID); err != nil {
			log.Printf("Failed to flush pending messages for user %d: %v", userID, err)
		}
	}()

	defer func() {
		h.hub.Unregister(userID)
		// Update user status to offline
		go func() {
			if h.userCache != nil {
				if err := h.userCache.SetUserOffline(userID); err != nil {
					log.Printf("Failed to set user %d offline in cache: %v", userID, err)
				}
			}
			if err := h.userService.SetUserOffline(userID); err != nil {
				log.Printf("Failed to set user %d offline in DB: %v", userID, err)
			}
		}()
	}()

	log.Printf("User %d connected via WebSocket", userID)

	// Create message context
	ctx := &ws.MessageContext{
		UserID:         userID,
		Conn:           c,
		Hub:            h.hub,
		MessageService: h.messageService,
		UserService:    h.userService,
		GroupService:   h.groupService,
		MessageCache:   h.messageCache,
		UserCache:      h.userCache,
	}

	// Handle incoming messages
	for {
		_, messageBytes, err := c.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from user %d: %v", userID, err)
			break
		}

		// Deserialize message
		msg, err := ws.Deserialize(messageBytes)
		if err != nil {
			log.Printf("Error deserializing message from user %d: %v", userID, err)
			ws.SendError(c, "invalid_message", "Invalid message format", err.Error())
			continue
		}

		// Process message
		if err := msg.Process(ctx); err != nil {
			log.Printf("Error processing message %s from user %d: %v", msg.GetType(), userID, err)
			ws.SendError(c, "processing_failed", "Failed to process message", err.Error())
		}
	}

	log.Printf("User %d disconnected from WebSocket", userID)
}
