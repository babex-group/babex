package kafka

import (
	"encoding/json"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/matroskin13/babex"
	"fmt"
)

type Adapter struct {
	Consumer *cluster.Consumer
	Producer sarama.SyncProducer

	options Options
	ch      chan *babex.Message
	err     chan error
}

type Options struct {
	Name   string
	Addrs  []string
	Topics []string

	consumerConfig *cluster.Config
}

func NewAdapter(options Options) (*Adapter, error) {
	if options.consumerConfig == nil {
		options.consumerConfig = cluster.NewConfig()
	}

	consumer, err := cluster.NewConsumer(
		options.Addrs,
		options.Name,
		options.Topics,
		options.consumerConfig,
	)
	if err != nil {
		return nil, err
	}

	producerConfig := sarama.NewConfig()
	producerConfig.Producer.RequiredAcks = sarama.WaitForAll
	producerConfig.Producer.Retry.Max = 5
	producerConfig.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(options.Addrs, producerConfig)
	if err != nil {
		return nil, err
	}

	adapter := Adapter{
		Consumer: consumer,
		options:  options,
		Producer: producer,
		ch:       make(chan *babex.Message),
		err:      make(chan error),
	}

	go func() {
	MainLoop:
		for {
			select {
			case msg, ok := <-adapter.Consumer.Messages():
				if !ok {
					break MainLoop
				}

				m, err := NewMessage(adapter.Consumer, msg)
				if err != nil {
					adapter.err <- err
					adapter.Consumer.MarkOffset(msg, "")
					continue
				}

				adapter.ch <- m
			case err, ok := <-adapter.Consumer.Errors():
				if ok {
					adapter.err <- fmt.Errorf("kafka consumer error: %s", err)
				}
			}
		}

		close(adapter.ch)
	}()

	return &adapter, nil
}

func (a Adapter) GetMessages() (<-chan *babex.Message, error) {
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

	_, _, err = a.Producer.SendMessage(&sarama.ProducerMessage{
		Topic: exchange,
		Value: sarama.ByteEncoder(b),
	})

	return err
}

func (a *Adapter) Publish(exchange string, key string, message babex.InitialMessage) error {
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, _, err = a.Producer.SendMessage(&sarama.ProducerMessage{
		Topic: exchange,
		Value: sarama.ByteEncoder(b),
	})

	return err
}
