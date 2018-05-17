package babex

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

type Message struct {
	Key     string
	Chain   []*ChainItem
	Data    json.RawMessage
	Headers map[string]interface{}
	Config  []byte

	msg *amqp.Delivery
}

type InitialMessage struct {
	Chain  []*ChainItem    `json:"chain"`
	Data   json.RawMessage `json:"data"`
	Config json.RawMessage `json:"config"`
}

func NewMessage(msg *amqp.Delivery) (*Message, error) {
	var initialMessage InitialMessage

	if err := json.Unmarshal(msg.Body, &initialMessage); err != nil {
		return nil, err
	}

	message := Message{
		Key:     msg.RoutingKey,
		Chain:   initialMessage.Chain,
		Data:    initialMessage.Data,
		msg:     msg,
		Headers: msg.Headers,
		Config:  initialMessage.Config,
	}

	return &message, nil
}

func (m Message) Ack(status bool) error {
	return m.msg.Ack(status)
}

func (m Message) Nack(status bool) error {
	return m.msg.Nack(false, true)
}

func (m *Message) SetChain(chain []ChainItem) error {
	return nil
}

func (m *Message) SetConfig(config []byte) {
	m.Config = json.RawMessage(config)
}
