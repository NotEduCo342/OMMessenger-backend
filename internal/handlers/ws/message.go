package ws

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gofiber/websocket/v2"
)

type Message interface {
	GetType() string
	Process(*websocket.Conn)
}

type SerializedMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
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

// Remember to add new messages to the type registry too
