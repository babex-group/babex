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

// Create Babex service
func NewService(adapter Adapter) *Service {
	service := Service{
		adapter: adapter,
	}

	return &service
}

func (s *Service) PublishMessage(exchange string, key string, chain []*ChainItem, data interface{}, headers map[string]interface{}, config json.RawMessage) error {
	return s.adapter.PublishMessage(exchange, key, chain, data, headers, config)
}

func (s Service) Next(msg *Message, data interface{}, headers map[string]interface{}) error {
	err := msg.RawMessage.Ack(true)
	if err != nil {
		return err
	}

	if headers == nil {
		headers = msg.Headers
	}

	if msg.Chain == nil {
		return ErrorChainIsEmpty
	}

	currIndex, currElement := getCurrentItem(msg.Chain)

	currElement.Successful = true

	if len(msg.Chain) <= currIndex+1 {
		return ErrorNextIsNotDefined
	}

	nextElement := msg.Chain[currIndex+1]

	if nextElement.IsMultiple {
		val := reflect.ValueOf(data)

		if val.Kind() != reflect.Slice {
			return ErrorDataIsNotArray
		}

		for i := 0; i < val.Len(); i++ {
			item := val.Index(i).Interface()

			err := s.adapter.PublishMessage(
				nextElement.Exchange,
				nextElement.Key,
				msg.Chain,
				item,
				headers,
				msg.Config,
			)
			if err != nil {
				return err
			}
		}
	} else {
		err := s.adapter.PublishMessage(
			nextElement.Exchange,
			nextElement.Key,
			msg.Chain,
			data,
			headers,
			msg.Config,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// Get channel for receive messages
func (s *Service) GetMessages() (<-chan *Message, error) {
	return s.adapter.GetMessages()
}

// Get channel for fatal errors
func (s *Service) GetErrors() chan error {
	return s.adapter.GetErrors()
}

func getCurrentItem(chain []*ChainItem) (int, *ChainItem) {
	for i, item := range chain {
		if item.Successful != true {
			return i, item
		}
	}

	return -1, nil
}
