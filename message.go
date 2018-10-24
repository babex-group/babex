package babex

import (
	"encoding/json"
)

type RawMessage interface {
	Ack() error
	Nack() error
}

type Message struct {
	Exchange string
	Key      string
	Chain    Chain
	Data     json.RawMessage
	Headers  map[string]interface{}
	Config   []byte
	Meta     map[string]string

	InitialMessage *InitialMessage
	RawMessage     RawMessage
}

func NewMessage(initialMessage *InitialMessage, exchange, key string) *Message {
	return &Message{
		Exchange:       exchange,
		Key:            key,
		Chain:          initialMessage.Chain,
		Data:           initialMessage.Data,
		Config:         initialMessage.Config,
		Meta:           initialMessage.Meta,
		InitialMessage: initialMessage,
	}
}

func (m Message) Ack() error {
	return m.RawMessage.Ack()
}

func (m Message) Nack() error {
	return m.RawMessage.Nack()
}

type InitialMessage struct {
	Chain  Chain             `json:"chain"`
	Data   json.RawMessage   `json:"data"`
	Config json.RawMessage   `json:"config"`
	Meta   map[string]string `json:"meta"`
	Catch  Chain             `json:"catch"`
}

// DataAll is container for total number of multiple-sent messages and shoud be placed in `data` key of a message.
type DataAll struct {
	All int `json:"all"`
}
