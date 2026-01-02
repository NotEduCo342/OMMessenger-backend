package ws

import (
	"reflect"
)

var typeRegistry = map[string]reflect.Type{}

func init() {
	RegisterType(&MessageStatus{})
}

func RegisterType(msg Message) {
	typeRegistry[msg.GetType()] = reflect.TypeOf(msg).Elem()
}
