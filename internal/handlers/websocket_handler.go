package handlers

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
)

type WebSocketHandler struct {
	messageService *service.MessageService
	clients        map[uint]*websocket.Conn
	clientsMux     sync.RWMutex
}

func NewWebSocketHandler(messageService *service.MessageService) *WebSocketHandler {
	return &WebSocketHandler{
		messageService: messageService,
		clients:        make(map[uint]*websocket.Conn),
	}
}

type WSMessage struct {
	Type        string      `json:"type"` // "message", "typing", "read", "delivered"
	RecipientID *uint       `json:"recipient_id,omitempty"`
	Content     string      `json:"content,omitempty"`
	MessageID   uint        `json:"message_id,omitempty"`
}

func (h *WebSocketHandler) HandleWebSocket(c *websocket.Conn) {
	userID := c.Locals("userID").(uint)
	
	// Register client
	h.clientsMux.Lock()
	h.clients[userID] = c
	h.clientsMux.Unlock()

	log.Printf("User %d connected via WebSocket", userID)

	// Cleanup on disconnect
	defer func() {
		h.clientsMux.Lock()
		delete(h.clients, userID)
		h.clientsMux.Unlock()
		c.Close()
		log.Printf("User %d disconnected from WebSocket", userID)
	}()

	// Handle incoming messages
	for {
		var msg WSMessage
		if err := c.ReadJSON(&msg); err != nil {
			log.Printf("Error reading message from user %d: %v", userID, err)
			break
		}

		switch msg.Type {
		case "message":
			h.handleNewMessage(userID, msg)
		case "typing":
			h.handleTyping(userID, msg)
		case "read":
			h.handleReadReceipt(msg)
		case "delivered":
			h.handleDeliveredReceipt(msg)
		}
	}
}

func (h *WebSocketHandler) handleNewMessage(senderID uint, msg WSMessage) {
	// Save message to database
	input := service.SendMessageInput{
		RecipientID: msg.RecipientID,
		Content:     msg.Content,
		MessageType: "text",
	}

	message, err := h.messageService.SendMessage(senderID, input)
	if err != nil {
		log.Printf("Error saving message: %v", err)
		return
	}

	// Send to recipient if online
	if msg.RecipientID != nil {
		h.sendToUser(*msg.RecipientID, map[string]interface{}{
			"type":    "message",
			"message": message.ToResponse(),
		})
	}
}

func (h *WebSocketHandler) handleTyping(senderID uint, msg WSMessage) {
	if msg.RecipientID != nil {
		h.sendToUser(*msg.RecipientID, map[string]interface{}{
			"type":      "typing",
			"sender_id": senderID,
		})
	}
}

func (h *WebSocketHandler) handleReadReceipt(msg WSMessage) {
	if err := h.messageService.MarkAsRead(msg.MessageID); err != nil {
		log.Printf("Error marking message as read: %v", err)
	}
}

func (h *WebSocketHandler) handleDeliveredReceipt(msg WSMessage) {
	if err := h.messageService.MarkAsDelivered(msg.MessageID); err != nil {
		log.Printf("Error marking message as delivered: %v", err)
	}
}

func (h *WebSocketHandler) sendToUser(userID uint, data interface{}) {
	h.clientsMux.RLock()
	conn, exists := h.clients[userID]
	h.clientsMux.RUnlock()

	if !exists {
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling data: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		log.Printf("Error sending message to user %d: %v", userID, err)
	}
}
