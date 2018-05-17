package babex

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"reflect"

	"github.com/streadway/amqp"
)

var (
	ErrorNextIsNotDefined     = errors.New("next is not defined")
	ErrorDataIsNotArray       = errors.New("data is not array. next item of chain has isMultiple flag")
	ErrorChainIsEmpty         = errors.New("chain is empty")
	ErrorQueueIsNotInitialize = errors.New("queue is not initialized")
)

type ServiceConfig struct {
	Name             string
	Address          string
	IsSingle         bool
	SkipDeclareQueue bool
}

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

type Service struct {
	Channel *amqp.Channel
	Queue   *amqp.Queue
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
	err := msg.Ack(true)

	if headers == nil {
		headers = msg.Headers
	}

	if err != nil {
		return err
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

	log.Printf("Publish message to %s:%s\r\n", nextElement.Exchange, nextElement.Key)

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

type Handler func(message *Message) error

func (s Service) ListenMessages(msgs <-chan amqp.Delivery, h Handler) error {
	for msg := range msgs {
		m, err := NewMessage(&msg)

		if err != nil {
			log.Println(err)
			continue
		}

		currIndex, _ := getCurrentItem(m.Chain)

		log.Printf("Recieve message. Routing key: %s\r\n", msg.RoutingKey)

		if currIndex > 0 {
			prevItem := m.Chain[currIndex-1]
			log.Printf("Previous chain item is %s:%s\r\n", prevItem.Exchange, prevItem.Key)
		}

		err = h(m)

		if err != nil {
			log.Println(err)
			continue
		}
	}

	return nil
}

func (s Service) Listen(h Handler) error {
	msgs, err := s.Channel.Consume(
		s.Queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	log.Println("service listen queue", s.Queue.Name)

	return s.ListenMessages(msgs, h)
}

func getCurrentItem(chain []*ChainItem) (int, *ChainItem) {
	for i, item := range chain {
		if item.Successful != true {
			return i, item
		}
	}

	return -1, nil
}
