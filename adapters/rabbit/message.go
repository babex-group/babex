package rabbit

import (
	"encoding/json"

	"github.com/matroskin13/babex"
	"github.com/streadway/amqp"
)

func NewMessage(msg *amqp.Delivery) (*babex.Message, error) {
	var initialMessage babex.InitialMessage

	if err := json.Unmarshal(msg.Body, &initialMessage); err != nil {
		return nil, err
	}

	message := babex.Message{
		Key:        msg.RoutingKey,
		Chain:      initialMessage.Chain,
		Data:       initialMessage.Data,
		Headers:    msg.Headers,
		Config:     initialMessage.Config,
		RawMessage: RabbitMessage{msg: msg},
	}

	return &message, nil
}

type RabbitMessage struct {
	msg *amqp.Delivery
}

func (m RabbitMessage) Ack(multiple bool) error {
	return m.msg.Ack(multiple)
}

func (m RabbitMessage) Nack(multiple bool) error {
	return m.msg.Nack(multiple, true)
}
