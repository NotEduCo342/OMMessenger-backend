package ws

import "github.com/gofiber/websocket/v2"

const (
	MsgStatus = "Status"
)

type MessageStatus struct {
	List map[string]string `json:"list"`
}

func (msg *MessageStatus) GetType() string {
	return MsgStatus
}

func (msg *MessageStatus) Process(ws *websocket.Conn) {

}
