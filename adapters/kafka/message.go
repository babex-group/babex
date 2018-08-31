package kafka

import (
	"encoding/json"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/matroskin13/babex"
)

func NewMessage(consumer *cluster.Consumer, msg *sarama.ConsumerMessage) (*babex.Message, error) {
	var initialMessage babex.InitialMessage

	if err := json.Unmarshal(msg.Value, &initialMessage); err != nil {
		return nil, err
	}

	message := babex.Message{
		Key:        string(msg.Key),
		Chain:      initialMessage.Chain,
		Data:       initialMessage.Data,
		Config:     initialMessage.Config,
		RawMessage: KafkaMessage{msg: msg, consumer: consumer},
	}

	return &message, nil
}

type KafkaMessage struct {
	msg      *sarama.ConsumerMessage
	consumer *cluster.Consumer
}

func (m KafkaMessage) Ack(multiple bool) error {
	m.consumer.MarkOffset(m.msg, "")
	return nil
}

func (m KafkaMessage) Nack(multiple bool) error {
	return nil
}
