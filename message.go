package babex

import (
	"encoding/json"
)

type RawMessage interface {
	Ack(multiple bool) error
	Nack(multiple bool) error
}

type Message struct {
	Key     string
	Chain   []*ChainItem
	Data    json.RawMessage
	Headers map[string]interface{}
	Config  []byte

	RawMessage RawMessage
}

type InitialMessage struct {
	Chain  []*ChainItem    `json:"chain"`
	Data   json.RawMessage `json:"data"`
	Config json.RawMessage `json:"config"`
}

func (m Message) Ack(multiple bool) error {
	return m.RawMessage.Nack(multiple)
}

func (m Message) Nack(multiple bool) error {
	return m.RawMessage.Nack(multiple)
}
