package ws

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gofiber/websocket/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
)

// MessageContext provides all dependencies needed for message processing
type MessageContext struct {
	UserID         uint
	Conn           *websocket.Conn
	Hub            *Hub
	MessageService *service.MessageService
	UserService    *service.UserService
}

// Message interface for all WebSocket message types
type Message interface {
	GetType() string
	Process(ctx *MessageContext) error
}

// SerializedMessage is the wire format wrapper
type SerializedMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ErrorResponse is sent when message processing fails
type ErrorResponse struct {
	Type    string `json:"type"`
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

func ToJson(msg Message) ([]byte, error) {
	return json.Marshal(msg)
}

func FromJson(jsonBytes []byte, msg Message) error {
	return json.Unmarshal(jsonBytes, msg)
}

func CreateMessage(msgType string, typeRegistry map[string]reflect.Type) (Message, error) {
	msgTypeReflect, ok := typeRegistry[msgType]
	if !ok {
		return nil, fmt.Errorf("unknown message type: %s", msgType)
	}

	instance := reflect.New(msgTypeReflect).Interface()
	return instance.(Message), nil
}

// SendError sends an error response to the client
func SendError(conn *websocket.Conn, code, message, details string) error {
	errResp := ErrorResponse{
		Type:    "error",
		Error:   message,
		Code:    code,
		Details: details,
	}
	return conn.WriteJSON(errResp)
}
