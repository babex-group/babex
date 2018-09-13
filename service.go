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

func (s *Service) PublishMessage(exchange string, key string, chain []ChainItem, data interface{}, meta map[string]string, config json.RawMessage) error {
	return s.adapter.PublishMessage(exchange, key, chain, data, meta, config)
}

func (s *Service) Catch(msg *Message, err error) error {
	if len(msg.InitialMessage.Catch) == 0 {
		return nil
	}

	return s.adapter.PublishMessage(
		msg.InitialMessage.Catch[0].Exchange,
		msg.InitialMessage.Catch[0].Key,
		msg.InitialMessage.Catch,
		CatchData{Error: err, Exchange: msg.Exchange, Key: msg.Key},
		nil,
		msg.Config,
	)
}

// Publish the message to next elements of chain
func (s Service) Next(msg *Message, data interface{}, meta map[string]string) error {
	err := msg.RawMessage.Ack(true)
	if err != nil {
		return err
	}

	if meta == nil {
		meta = msg.Meta
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

	if nextElement.Meta != nil {
		for key, metaItem := range nextElement.Meta {
			meta[key] = metaItem
		}
	}

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
				chain,
				item,
				meta,
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
			chain,
			data,
			meta,
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

// Get channel for errors
func (s *Service) GetErrors() chan error {
	return s.adapter.GetErrors()
}
