package ws

import (
	"reflect"
)

var typeRegistry = map[string]reflect.Type{}

func init() {
	// Register all message types
	RegisterType(&MessageSync{})
	RegisterType(&MessageChat{})
	RegisterType(&MessageAck{})
	RegisterType(&MessageTyping{})
	RegisterType(&MessageRead{})
	RegisterType(&MessageDelivery{})
	RegisterType(&MessageGroupRead{})
	RegisterType(&MessagePing{})
	RegisterType(&MessagePong{})
}

func RegisterType(msg Message) {
	typeRegistry[msg.GetType()] = reflect.TypeOf(msg).Elem()
}

// GetTypeRegistry returns the type registry for testing
func GetTypeRegistry() map[string]reflect.Type {
	return typeRegistry
}
