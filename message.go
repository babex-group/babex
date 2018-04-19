package babex

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

type Message struct {
	Key   string
	Chain []*ChainItem
	Data  json.RawMessage

	msg *amqp.Delivery
}

type InitialMessage struct {
	Chain []*ChainItem    `json:"chain"`
	Data  json.RawMessage `json:"data"`
}

func NewMessage(msg *amqp.Delivery) (*Message, error) {
	var initialMessage InitialMessage

	if err := json.Unmarshal(msg.Body, &initialMessage); err != nil {
		return nil, err
	}

	message := Message{
		Key:   msg.RoutingKey,
		Chain: initialMessage.Chain,
		Data:  initialMessage.Data,
		msg:   msg,
	}

	return &message, nil
}

func (m Message) Ack(status bool) error {
	return m.msg.Ack(status)
}

func (m *Message) SetChain(chain []ChainItem) error {
	return nil
}
