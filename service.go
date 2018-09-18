package babex

import (
	"encoding/json"
	"errors"
	"reflect"
)

var (
	ErrorNextIsNotDefined = errors.New("next is not defined")
	ErrorDataIsNotArray   = errors.New("data is not array. next item of chain has isMultiple flag")
	ErrorChainIsEmpty     = errors.New("chain is empty")
	ErrorCloseConsumer    = errors.New("close consumer")
)

type Service struct {
	adapter Adapter
}

// Create Babex service via the adapter interface
func NewService(adapter Adapter) *Service {
	service := Service{
		adapter: adapter,
	}

	return &service
}

// Publish message
func (s *Service) Publish(message InitialMessage) error {
	nextIndex := getCurrentChainIndex(message.Chain)
	if nextIndex == -1 {
		return ErrorNextIsNotDefined
	}

	nextElement := message.Chain[nextIndex]

	return s.adapter.Publish(nextElement.Exchange, nextElement.Key, message)
}

// The catch method allows publish error to Catch chain.
// For example:
// {
//    "chain": [...],
//	  "catch": [{
// 		"exchange": "error-topic"
// 	  }]
// }
// If you have an exception you can publish error to the catch chain:
//
// if err != nil {
// 		msg.Ack()
//		service.Catch(msg, err)
// }
func (s *Service) Catch(msg *Message, err error) error {
	defer msg.Ack(false)

	if len(msg.InitialMessage.Catch) == 0 {
		return nil
	}

	b, err := json.Marshal(CatchData{Error: err.Error(), Exchange: msg.Exchange, Key: msg.Key})
	if err != nil {
		return err
	}

	m := InitialMessage{
		Config: msg.Config,
		Chain:  msg.InitialMessage.Chain,
		Data:   b,
		Meta:   msg.InitialMessage.Meta,
	}

	return s.Publish(m)
}

// Publish the message to next elements of chain
//
// The data argument is any GO type.
// If current element of chain has multiple flag, you can put the slice.
//
// Every message of chain has meta property (map[string]string]).
// You can use it instead amqp headers.
// If you put the useMeta argument, the babex merges the current meta with useMeta.
func (s Service) Next(msg *Message, data interface{}, useMeta map[string]string) error {
	err := msg.RawMessage.Ack(true)
	if err != nil {
		return err
	}

	if msg.Chain == nil {
		return ErrorChainIsEmpty
	}

	chain := SetCurrentItemSuccess(msg.Chain)

	nextIndex := getCurrentChainIndex(chain)
	if nextIndex == -1 {
		return ErrorNextIsNotDefined
	}

	nextElement := chain[nextIndex]

	meta := Meta{}
	meta.Merge(msg.Meta, nextElement.Meta, useMeta)

	var items []interface{}

	if nextElement.IsMultiple {
		val := reflect.ValueOf(data)

		if val.Kind() != reflect.Slice {
			return ErrorDataIsNotArray
		}

		for i := 0; i < val.Len(); i++ {
			items = append(items, val.Index(i).Interface())
		}
	} else {
		items = append(items, data)
	}

	for _, item := range items {
		b, err := json.Marshal(item)
		if err != nil {
			return err
		}

		m := InitialMessage{
			Catch:  msg.InitialMessage.Catch,
			Config: msg.Config,
			Meta:   meta,
			Chain:  chain,
			Data:   b,
		}

		if err := s.Publish(m); err != nil {
			return err
		}
	}

	return nil
}

// Get channel for receive messages
func (s *Service) GetMessages() (<-chan *Message, error) {
	return s.adapter.GetMessages()
}

// Get channel for errors
func (s *Service) GetErrors() chan error {
	return s.adapter.GetErrors()
}
