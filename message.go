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
