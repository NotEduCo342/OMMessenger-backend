package ws

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/repository"
)

// ClientConnection wraps a WebSocket connection with metadata
type ClientConnection struct {
	Conn         *websocket.Conn
	UserID       uint
	LastPong     time.Time
	SupportsGzip bool
	PingTicker   *time.Ticker
	CloseChan    chan struct{}
}

// Hub manages all active WebSocket connections
type Hub struct {
	clients            map[uint]*ClientConnection
	clientsMux         sync.RWMutex
	pendingMessageRepo repository.PendingMessageRepositoryInterface
	deliveryRetryQueue chan *DeliveryAttempt
	maxRetries         int
	baseRetryDelay     time.Duration
	pingInterval       time.Duration
	pongTimeout        time.Duration
}

// DeliveryAttempt represents a message delivery attempt
type DeliveryAttempt struct {
	UserID       uint
	PendingMsgID uint
	Payload      string
	Attempts     int
	Priority     int
}

// NewHub creates a new Hub instance
func NewHub(pendingRepo repository.PendingMessageRepositoryInterface) *Hub {
	hub := &Hub{
		clients:            make(map[uint]*ClientConnection),
		pendingMessageRepo: pendingRepo,
		deliveryRetryQueue: make(chan *DeliveryAttempt, 1000),
		maxRetries:         5,
		baseRetryDelay:     2 * time.Second,
		pingInterval:       30 * time.Second,
		pongTimeout:        90 * time.Second,
	}

	// Start background workers
	go hub.retryWorker()
	go hub.connectionHealthChecker()

	return hub
}

// Register adds a client connection with health monitoring
func (h *Hub) Register(userID uint, conn *websocket.Conn, supportsGzip bool) {
	clientConn := &ClientConnection{
		Conn:         conn,
		UserID:       userID,
		LastPong:     time.Now(),
		SupportsGzip: supportsGzip,
		PingTicker:   time.NewTicker(h.pingInterval),
		CloseChan:    make(chan struct{}),
	}

	// Setup pong handler
	conn.SetPongHandler(func(appData string) error {
		h.clientsMux.Lock()
		if client, exists := h.clients[userID]; exists {
			client.LastPong = time.Now()
		}
		h.clientsMux.Unlock()
		return nil
	})

	// Set read deadline for ping/pong
	conn.SetReadDeadline(time.Now().Add(h.pongTimeout))

	h.clientsMux.Lock()
	h.clients[userID] = clientConn
	h.clientsMux.Unlock()

	// Start ping routine
	go h.pingRoutine(clientConn)

	log.Printf("User %d connected to hub (total: %d, gzip: %v)", userID, len(h.clients), supportsGzip)
}

// Unregister removes a client connection
func (h *Hub) Unregister(userID uint) {
	h.clientsMux.Lock()
	if client, exists := h.clients[userID]; exists {
		if client.PingTicker != nil {
			client.PingTicker.Stop()
		}
		close(client.CloseChan)
	}
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

// SendToUser sends data to a specific user with optional compression
func (h *Hub) SendToUser(userID uint, data interface{}) error {
	return h.SendToUserWithID(userID, 0, data)
}

// SendToUserWithID sends data with explicit message ID for queueing
func (h *Hub) SendToUserWithID(userID uint, messageID uint, data interface{}) error {
	h.clientsMux.RLock()
	clientConn, exists := h.clients[userID]
	h.clientsMux.RUnlock()

	if !exists {
		// User offline, queue message for later delivery
		return h.queueMessage(userID, messageID, data, 0)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling data for user %d: %v", userID, err)
		return err
	}

	// Compress if supported and beneficial (> 512 bytes)
	var finalData []byte
	frameType := websocket.TextMessage
	if clientConn.SupportsGzip && len(jsonData) > 512 {
		compressed, err := h.compressData(jsonData)
		if err == nil && len(compressed) < len(jsonData) {
			finalData = compressed
			frameType = websocket.BinaryMessage
		} else {
			finalData = jsonData
		}
	} else {
		finalData = jsonData
	}

	if err := clientConn.Conn.WriteMessage(frameType, finalData); err != nil {
		log.Printf("Error sending message to user %d: %v", userID, err)
		// Connection may be dead, unregister and queue message
		h.Unregister(userID)
		return h.queueMessage(userID, messageID, data, 0)
	}

	return nil
}

// queueMessage stores a message for offline or failed delivery
func (h *Hub) queueMessage(userID uint, messageID uint, data interface{}, priority int) error {
	if h.pendingMessageRepo == nil {
		return nil // No repository configured, skip queueing
	}

	// Don't queue ephemeral messages (typing, ping, etc)
	if dataMap, ok := data.(map[string]interface{}); ok {
		msgType, _ := dataMap["type"].(string)
		if msgType == "typing" || msgType == "ping" || msgType == "pong" {
			return nil // Skip queueing ephemeral messages
		}
	}

	// Skip if no valid message ID (can't satisfy foreign key constraint)
	if messageID == 0 {
		log.Printf("⚠️ Skipping queue for user %d: no valid message_id", userID)
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return h.pendingMessageRepo.Enqueue(userID, messageID, string(jsonData), priority)
}

// Broadcast sends data to all connected users
func (h *Hub) Broadcast(data interface{}) {
	h.clientsMux.RLock()
	clients := make(map[uint]*ClientConnection, len(h.clients))
	for id, conn := range h.clients {
		clients[id] = conn
	}
	h.clientsMux.RUnlock()

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling broadcast data: %v", err)
		return
	}

	for userID, clientConn := range clients {
		if err := clientConn.Conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
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
		if clientConn, exists := h.clients[userID]; exists {
			if err := clientConn.Conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
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

// FlushPendingMessages sends all queued messages to a newly connected user
func (h *Hub) FlushPendingMessages(userID uint) error {
	if h.pendingMessageRepo == nil {
		return nil
	}

	// Get connection
	h.clientsMux.RLock()
	clientConn, exists := h.clients[userID]
	h.clientsMux.RUnlock()

	if !exists {
		return nil // User disconnected already
	}

	// Fetch pending messages in batches
	batchSize := 50
	pending, err := h.pendingMessageRepo.GetPendingForUser(userID, batchSize)
	if err != nil {
		log.Printf("Error fetching pending messages for user %d: %v", userID, err)
		return err
	}

	if len(pending) == 0 {
		return nil
	}

	log.Printf("Flushing %d pending messages to user %d", len(pending), userID)

	// Send messages in batch
	batch := make([]interface{}, 0, len(pending))
	successIDs := make([]uint, 0, len(pending))

	for _, pm := range pending {
		var data interface{}
		if err := json.Unmarshal([]byte(pm.Payload), &data); err != nil {
			log.Printf("Error unmarshaling pending message %d: %v", pm.ID, err)
			continue
		}
		batch = append(batch, data)
		successIDs = append(successIDs, pm.ID)
	}

	// Send batch envelope
	batchMessage := map[string]interface{}{
		"type":     "batch",
		"messages": batch,
		"count":    len(batch),
	}

	if err := clientConn.Conn.WriteJSON(batchMessage); err != nil {
		log.Printf("Error sending batch to user %d: %v", userID, err)
		// Connection failed, messages stay in queue
		return err
	}

	// Successfully delivered, remove from queue
	if err := h.pendingMessageRepo.DeleteBatch(successIDs); err != nil {
		log.Printf("Error deleting delivered messages: %v", err)
	}

	// If there are more messages, recursively flush (rate-limited by batch size)
	if len(pending) == batchSize {
		// Small delay to avoid overwhelming the connection
		time.Sleep(100 * time.Millisecond)
		return h.FlushPendingMessages(userID)
	}

	return nil
}

// retryWorker processes failed deliveries with exponential backoff
func (h *Hub) retryWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if h.pendingMessageRepo == nil {
			continue
		}

		// Get messages ready for retry
		retryable, err := h.pendingMessageRepo.GetRetryable(100)
		if err != nil {
			log.Printf("Error fetching retryable messages: %v", err)
			continue
		}

		for _, pm := range retryable {
			// Check if user is now online
			h.clientsMux.RLock()
			clientConn, isOnline := h.clients[pm.UserID]
			h.clientsMux.RUnlock()

			if !isOnline {
				// Still offline, calculate next retry with exponential backoff
				attempts := pm.Attempts + 1
				if attempts >= h.maxRetries {
					// Max retries reached, keep in queue but don't retry for a while
					nextRetry := time.Now().Add(1 * time.Hour)
					h.pendingMessageRepo.MarkAttempted(pm.ID, attempts, &nextRetry)
					continue
				}

				// Exponential backoff: 2s, 4s, 8s, 16s, 32s
				delay := h.baseRetryDelay * time.Duration(1<<uint(attempts))
				nextRetry := time.Now().Add(delay)
				h.pendingMessageRepo.MarkAttempted(pm.ID, attempts, &nextRetry)
				continue
			}

			// User is online, attempt delivery
			var data interface{}
			if err := json.Unmarshal([]byte(pm.Payload), &data); err != nil {
				log.Printf("Error unmarshaling message for retry %d: %v", pm.ID, err)
				continue
			}

			jsonData, _ := json.Marshal(data)
			if err := clientConn.Conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
				log.Printf("Retry delivery failed for user %d: %v", pm.UserID, err)
				// Mark for next retry
				attempts := pm.Attempts + 1
				delay := h.baseRetryDelay * time.Duration(1<<uint(attempts))
				nextRetry := time.Now().Add(delay)
				h.pendingMessageRepo.MarkAttempted(pm.ID, attempts, &nextRetry)
			} else {
				// Successfully delivered, remove from queue
				log.Printf("Successfully delivered pending message %d to user %d", pm.ID, pm.UserID)
				h.pendingMessageRepo.Delete(pm.ID)
			}
		}
	}
}

// pingRoutine sends periodic ping messages to keep connection alive
func (h *Hub) pingRoutine(client *ClientConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Ping routine recovered from panic for user %d: %v", client.UserID, r)
		}
	}()

	for {
		select {
		case <-client.CloseChan:
			return
		case <-client.PingTicker.C:
			// Check if connection is still valid
			h.clientsMux.RLock()
			_, exists := h.clients[client.UserID]
			h.clientsMux.RUnlock()

			if !exists {
				return
			}

			if err := client.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				log.Printf("Ping failed for user %d: %v", client.UserID, err)
				h.Unregister(client.UserID)
				return
			}
		}
	}
}

// connectionHealthChecker monitors connection health and removes dead connections
func (h *Hub) connectionHealthChecker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.clientsMux.RLock()
		deadConnections := make([]uint, 0)
		now := time.Now()

		for userID, client := range h.clients {
			if now.Sub(client.LastPong) > h.pongTimeout {
				deadConnections = append(deadConnections, userID)
			}
		}
		h.clientsMux.RUnlock()

		// Unregister dead connections
		for _, userID := range deadConnections {
			log.Printf("Removing dead connection for user %d (no pong received)", userID)
			h.Unregister(userID)
		}
	}
}

// compressData compresses data using gzip
func (h *Hub) compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	if _, err := gzipWriter.Write(data); err != nil {
		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompressData decompresses gzip data
func (h *Hub) decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
