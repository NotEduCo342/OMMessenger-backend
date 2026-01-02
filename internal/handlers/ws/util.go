package ws

import "encoding/json"

func Serialize(msg Message) ([]byte, error) {
	wrapper := SerializedMessage{
		Type:    msg.GetType(),
		Payload: nil,
	}
	payload, err := ToJson(msg)
	if err != nil {
		return nil, err
	}
	wrapper.Payload = payload
	return json.Marshal(wrapper)
}

func Deserialize(jsonBytes []byte) (Message, error) {
	var wrapper SerializedMessage
	if err := json.Unmarshal(jsonBytes, &wrapper); err != nil {
		return nil, err
	}

	return DeserializeSerializedMessage(&wrapper)
}

func DeserializeSerializedMessage(wrapper *SerializedMessage) (Message, error) {
	msg, err := CreateMessage(wrapper.Type, typeRegistry)
	if err != nil {
		return nil, err
	}

	if err := FromJson(wrapper.Payload, msg); err != nil {
		return nil, err
	}

	return msg, nil
}
