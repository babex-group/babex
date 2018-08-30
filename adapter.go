package babex

import "encoding/json"

type Adapter interface {
	GetMessages() (<-chan *Message, error)
	GetErrors() chan error
	PublishMessage(exchange string, key string, chain []*ChainItem, data interface{}, headers map[string]interface{}, config json.RawMessage) error
}
