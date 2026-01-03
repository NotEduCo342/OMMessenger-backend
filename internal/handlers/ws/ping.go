package ws

// MessagePing is a keepalive ping from client
type MessagePing struct {
}

func (msg *MessagePing) GetType() string {
	return "ping"
}

func (msg *MessagePing) Process(ctx *MessageContext) error {
	// Respond with pong
	return ctx.Conn.WriteJSON(map[string]string{
		"type": "pong",
	})
}

// MessagePong is a pong response (in case client wants to track latency)
type MessagePong struct {
}

func (msg *MessagePong) GetType() string {
	return "pong"
}

func (msg *MessagePong) Process(ctx *MessageContext) error {
	// No-op - just acknowledge
	return nil
}
