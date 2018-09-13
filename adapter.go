package babex

import "encoding/json"

type Adapter interface {
	GetMessages() (<-chan *Message, error)
	GetErrors() chan error
	PublishMessage(exchange string, key string, chain []ChainItem, data interface{}, meta map[string]string, config json.RawMessage) error
	Publish(exchange string, key string, message InitialMessage) error
}
