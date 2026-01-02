package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

// Hub manages all active WebSocket connections
type Hub struct {
	clients    map[uint]*websocket.Conn
	clientsMux sync.RWMutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients: make(map[uint]*websocket.Conn),
	}
}

// Register adds a client connection
func (h *Hub) Register(userID uint, conn *websocket.Conn) {
	h.clientsMux.Lock()
	h.clients[userID] = conn
	h.clientsMux.Unlock()
	log.Printf("User %d connected to hub (total: %d)", userID, len(h.clients))
}

// Unregister removes a client connection
func (h *Hub) Unregister(userID uint) {
	h.clientsMux.Lock()
	delete(h.clients, userID)
	count := len(h.clients)
	h.clientsMux.Unlock()
	log.Printf("User %d disconnected from hub (total: %d)", userID, count)
}

// IsOnline checks if a user is connected
func (h *Hub) IsOnline(userID uint) bool {
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()
	_, exists := h.clients[userID]
	return exists
}

// SendToUser sends data to a specific user
func (h *Hub) SendToUser(userID uint, data interface{}) error {
	h.clientsMux.RLock()
	conn, exists := h.clients[userID]
	h.clientsMux.RUnlock()

	if !exists {
		return nil // User offline, message should be queued
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling data for user %d: %v", userID, err)
		return err
	}

	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		log.Printf("Error sending message to user %d: %v", userID, err)
		// Connection may be dead, unregister
		h.Unregister(userID)
		return err
	}

	return nil
}

// Broadcast sends data to all connected users
func (h *Hub) Broadcast(data interface{}) {
	h.clientsMux.RLock()
	clients := make(map[uint]*websocket.Conn, len(h.clients))
	for id, conn := range h.clients {
		clients[id] = conn
	}
	h.clientsMux.RUnlock()

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling broadcast data: %v", err)
		return
	}

	for userID, conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			log.Printf("Error broadcasting to user %d: %v", userID, err)
			h.Unregister(userID)
		}
	}
}

// BroadcastToUsers sends data to specific users
func (h *Hub) BroadcastToUsers(userIDs []uint, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling data: %v", err)
		return
	}

	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()

	for _, userID := range userIDs {
		if conn, exists := h.clients[userID]; exists {
			if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
				log.Printf("Error sending to user %d: %v", userID, err)
			}
		}
	}
}

// GetOnlineUsers returns list of currently connected user IDs
func (h *Hub) GetOnlineUsers() []uint {
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()

	users := make([]uint, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// Count returns the number of connected clients
func (h *Hub) Count() int {
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()
	return len(h.clients)
}
