package babex

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/streadway/amqp"
)

var (
	ErrorNextIsNotDefined     = errors.New("next is not defined")
	ErrorDataIsNotArray       = errors.New("data is not array. next item of chain has isMultiple flag")
	ErrorChainIsEmpty         = errors.New("chain is empty")
	ErrorQueueIsNotInitialize = errors.New("queue is not initialized")
	ErrorCloseConsumer        = errors.New("close consumer")
)

type ServiceConfig struct {
	Name             string // name of your service, and for declare queue
	Address          string // addr for rabbit, example amqp://guest:guest@localhost:5672
	IsSingle         bool   // if true, service create uniq queue (example - test.adska1231k)
	SkipDeclareQueue bool
	AutoAck          bool
}

type Service struct {
	Channel *amqp.Channel
	Queue   *amqp.Queue

	ch     chan *Message
	err    chan error
	config *ServiceConfig
}

// Create Babex service
func NewService(config *ServiceConfig) (*Service, error) {
	qName := config.Name

	if config.IsSingle {
		hash := md5.New()
		hash.Write([]byte(qName))

		qName = config.Name + "." + hex.EncodeToString(hash.Sum(nil))
	}

	conn, err := amqp.Dial(config.Address)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	service := Service{
		Channel: ch,
		ch:      make(chan *Message),
		err:     make(chan error),
		config:  config,
	}

	if config.SkipDeclareQueue == false {
		q, err := ch.QueueDeclare(
			qName,
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return nil, err
		}

		service.Queue = &q
	}

	return &service, nil
}

func (s *Service) BindToExchange(exchange string, key string) error {
	if s.Queue == nil {
		return ErrorQueueIsNotInitialize
	}

	return s.Channel.QueueBind(
		s.Queue.Name,
		key,
		exchange,
		false,
		nil,
	)
}

func (s *Service) PublishMessage(exchange string, key string, chain []*ChainItem, data interface{}, headers map[string]interface{}, config json.RawMessage) error {
	bData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	b, err := json.Marshal(InitialMessage{
		Data:   bData,
		Chain:  chain,
		Config: config,
	})
	if err != nil {
		return err
	}

	return s.Channel.Publish(
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			Body:    b,
			Headers: headers,
		},
	)
}

func (s Service) Next(msg *Message, data interface{}, headers map[string]interface{}) error {
	err := msg.Ack(false)
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

			err = s.PublishMessage(
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
		err = s.PublishMessage(
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
	msgs, err := s.Channel.Consume(
		s.Queue.Name,
		"",
		s.config.AutoAck,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	go func() {
		for msg := range msgs {
			m, err := NewMessage(msg)
			if err != nil {
				msg.Ack(false)
				continue
			}

			s.ch <- m
		}

		s.err <- ErrorCloseConsumer
	}()

	return s.ch, nil
}

// Get channel for fatal errors
func (s *Service) GetErrors() chan error {
	return s.err
}

func getCurrentItem(chain []*ChainItem) (int, *ChainItem) {
	for i, item := range chain {
		if item.Successful != true {
			return i, item
		}
	}

	return -1, nil
}
