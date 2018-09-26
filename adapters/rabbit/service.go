package rabbit

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/matroskin13/babex"
	"github.com/streadway/amqp"
)

var (
	ErrorQueueIsNotInitialize = errors.New("queue is not initialized")
)

type Adapter struct {
	Channel *amqp.Channel
	Queue   *amqp.Queue

	ch  chan *babex.Message
	err chan error

	options Options
}

type Options struct {
	Name             string // name of your service, and for declare queue
	Address          string // addr for rabbit, example amqp://guest:guest@localhost:5672
	IsSingle         bool   // if true, service create uniq queue (example - test.adska1231k)
	SkipDeclareQueue bool
	AutoAck          bool
}

func NewAdapter(options Options) (*Adapter, error) {
	qName := options.Name

	if options.IsSingle {
		hash := md5.New()
		hash.Write([]byte(qName))

		qName = options.Name + "." + hex.EncodeToString(hash.Sum(nil))
	}

	conn, err := amqp.Dial(options.Address)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	adapter := Adapter{
		Channel: ch,
		ch:      make(chan *babex.Message),
		err:     make(chan error),
		options: options,
	}

	if options.SkipDeclareQueue == false {
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

		adapter.Queue = &q
	}

	return &adapter, nil
}

func (a Adapter) GetMessages() (<-chan *babex.Message, error) {
	msgs, err := a.Channel.Consume(
		a.Queue.Name,
		"",
		a.options.AutoAck,
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
			m, err := NewMessage(&msg)
			if err != nil {
				msg.Ack(false)
				continue
			}

			a.ch <- m
		}

		close(a.ch)
	}()

	return a.ch, nil
}

// Get channel for fatal errors
func (a *Adapter) GetErrors() chan error {
	return a.err
}

func (a *Adapter) PublishMessage(exchange string, key string, chain []babex.ChainItem, data interface{}, meta map[string]string, config json.RawMessage) error {
	bData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	b, err := json.Marshal(babex.InitialMessage{
		Data:   bData,
		Chain:  chain,
		Config: config,
		Meta:   meta,
	})
	if err != nil {
		return err
	}

	return a.Channel.Publish(
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			Body: b,
		},
	)
}

func (a *Adapter) Publish(exchange string, key string, message babex.InitialMessage) error {
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return a.Channel.Publish(
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			Body: b,
		},
	)
}

func (a *Adapter) BindToExchange(exchange string, key string) error {
	if a.Queue == nil {
		return ErrorQueueIsNotInitialize
	}

	return a.Channel.QueueBind(
		a.Queue.Name,
		key,
		exchange,
		false,
		nil,
	)
}

func (a *Adapter) Close() error {
	return a.Channel.Close()
}
