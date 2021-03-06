package babex

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrorNextIsNotDefined = errors.New("next is not defined")
	ErrorNextNoCount      = errors.New("next does not contain count chain")
	ErrorDataIsNotArray   = errors.New("data is not array. next item of chain has isMultiple flag")
	ErrorChainIsEmpty     = errors.New("chain is empty")
	ErrorCloseConsumer    = errors.New("close consumer")
)

type Service struct {
	adapter    Adapter
	middleware []Middleware
	in         chan *Message
	channels   chan *Channel
	handlers   Handlers
	onclose    chan error
	logger     Logger
}

func NewServiceListener(adapter Adapter, middleware ...Middleware) *Service {
	service := NewService(adapter, middleware...)

	go service.listen(true)

	return service
}

// Create Babex service via the adapter interface
func NewService(adapter Adapter, middleware ...Middleware) *Service {
	service := Service{
		adapter:    adapter,
		middleware: middleware,
		in:         make(chan *Message),
		channels:   make(chan *Channel),
		handlers:   Handlers{},
		onclose:    make(chan error, 1),
		logger:     &StubLogger{},
	}

	return &service
}

func (s *Service) SetLogger(logger Logger) {
	s.logger = logger
}

func (s *Service) listen(receiveChannels bool) error {
	apply := func(msg *Message) error {
		for _, m := range s.middleware {
			finish, err := m.Use(s, msg)
			if err != nil {
				return err
			}

			msg.done = append(msg.done, finish)
		}

		return nil
	}

	channels := s.adapter.Channels()
	if channels != nil {
		for {
			select {
			case ch, ok := <-channels:
				if !ok {
					s.onclose <- nil
					return nil
				}

				s.logger.Log(fmt.Sprintf("debug_babex: receive channel. channel_info: %v", ch.Info))

				go func(ch *Channel) {
					messageChannel := make(chan *Message)

					if receiveChannels {
						sch := Channel{
							ch: messageChannel,
						}

						s.channels <- &sch

						s.logger.Log(fmt.Sprintf("debug_babex: success publish to GetChannels(). channel_info: %v", ch.Info))
					}

					for msg := range ch.GetMessages() {
						s.logger.Log(fmt.Sprintf("debug_babex: receive message from channel.GetMessages(). channel_info: %v", ch.Info))

						apply(msg)

						if h, ok := s.handlers[msg.Exchange+":"+msg.Key]; ok {
							err := h(msg)

							s.Done(msg, err)

							if err != nil {
								msg.Nack()
							}
						} else if receiveChannels {
							messageChannel <- msg
						}

						s.logger.Log(fmt.Sprintf("debug_babex: success publish message to channel. channel_info: %s", ch.Info))
					}

					s.logger.Log(fmt.Sprintf("debug_babex: close channel. channel_info: %s", ch.Info))

					close(messageChannel)
				}(ch)
			}
		}
	} else {
		msgs, _ := s.adapter.GetMessages()
		for msg := range msgs {
			apply(msg)

			if h, ok := s.handlers[msg.Exchange+":"+msg.Key]; ok {
				err := h(msg)

				s.Done(msg, err)

				if err != nil {
					msg.Nack()
				}
			} else if receiveChannels {
				s.in <- msg
			}
		}

		close(s.in)
		s.onclose <- nil
	}

	return nil
}

// Use the method for done middleware.
func (s *Service) Done(msg *Message, err error) {
	for _, done := range msg.done {
		done(err)
	}
}

// Publish message
func (s *Service) Publish(message InitialMessage) error {
	meta := Meta{}
	meta.Merge(message.Meta)

	var nextElement ChainItem

	for {
		nextIndex := getCurrentChainIndex(message.Chain)
		if nextIndex == -1 {
			return nil
		}

		next := message.Chain[nextIndex]

		ok, err := ApplyWhen(next.When, meta)
		if err != nil {
			return err
		}

		if ok {
			nextElement = next
			break
		}

		message.Chain = SetCurrentItemSuccess(message.Chain)
	}

	meta.Merge(nextElement.Meta)

	message.Meta = meta

	return s.adapter.Publish(nextElement.Exchange, nextElement.Key, message)
}

// Catch method allows publish error to Catch chain.
// For example:
//  {
//     "chain": [],
//     "catch": [{
//       "exchange": "error-topic"
//     }]
//  }
// If you have an exception you can publish error to the catch chain:
//
//  if err != nil {
//       msg.Ack()
//       service.Catch(msg, err, nil)
//  }
//
// You can use body argument for pass the custom data, otherwise the data will have msg.Data
// How you can handle catch data?
// For example:
//
//  var catch babex.CatchData
//
//  if err := json.Unmarshal(msg.Data, &catch); err != nil {}
//
//  fmt.Println(catch.Error)
func (s *Service) Catch(msg *Message, catchErr error, body []byte) error {
	s.Done(msg, nil)

	currentIndex := getCurrentChainIndex(msg.Chain)
	if currentIndex == -1 {
		return nil
	}
	currentElement := msg.Chain[currentIndex]

	chain := msg.InitialMessage.Catch
	if len(currentElement.Catch) != 0 {
		chain = currentElement.Catch
	}
	if len(chain) == 0 {
		return ErrorNextNoCount
	}

	if body == nil {
		body = msg.Data
	}

	catch := CatchData{
		Error:    catchErr.Error(),
		Exchange: msg.Exchange,
		Key:      msg.Key,
		Data:     body,
	}

	b, err := json.Marshal(catch)
	if err != nil {
		return err
	}

	m := InitialMessage{
		Config: msg.Config,
		Chain:  chain,
		Data:   b,
		Meta:   msg.InitialMessage.Meta,
	}

	if err := s.Publish(m); err != nil {
		return err
	}

	return msg.Ack()
}

// Count starts set count chain. Initial data for chain is object with key `all` containing total count of elements
func (s *Service) Count(msg *Message, count int) error {
	_, nextElement, err := s.chainCursor(msg)
	if err != nil {
		return err
	}
	if len(nextElement.SetCount) == 0 {
		return ErrorNextNoCount
	}

	data := DataAll{
		All: count,
	}

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	m := InitialMessage{
		Config: msg.Config,
		Chain:  nextElement.SetCount,
		Data:   b,
		Meta:   msg.InitialMessage.Meta,
	}

	return s.Publish(m)
}

func (s *Service) chainCursor(msg *Message) (Chain, ChainItem, error) {
	chain := SetCurrentItemSuccess(msg.Chain)
	nextIndex := getCurrentChainIndex(chain)
	if nextIndex == -1 {
		return nil, ChainItem{}, nil
	}

	nextElement := chain[nextIndex]
	return chain, nextElement, nil
}

// Next publishes the message to next elements of chain
//
// The data argument is any GO type.
// If current element of chain has multiple flag, you can put the slice.
//
// Every message of chain has meta property (map[string]string]).
// You can use it instead amqp headers.
// If you put the useMeta argument, the babex merges the current meta with useMeta.
func (s Service) Next(msg *Message, data interface{}, useMeta map[string]string) error {
	s.Done(msg, nil)

	err := msg.RawMessage.Ack()
	if err != nil {
		return err
	}

	if msg.Chain == nil {
		return ErrorChainIsEmpty
	}

	chain, nextElement, err := s.chainCursor(msg)
	if err != nil {
		return err
	}

	meta := Meta{}
	meta.Merge(msg.Meta, useMeta)

	var items []interface{}

	if nextElement.IsMultiple {
		val := reflect.ValueOf(data)

		if val.Kind() != reflect.Slice {
			return ErrorDataIsNotArray
		}

		for i := 0; i < val.Len(); i++ {
			items = append(items, val.Index(i).Interface())
		}

		if nextElement.SetCount != nil {
			if err := s.Count(msg, val.Len()); err != nil {
				return err
			}
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

func (s *Service) GetChannels() Channels {
	return s.channels
}

// Get channel for receive messages
func (s *Service) GetMessages() <-chan *Message {
	return s.in
}

// Get channel for errors
func (s *Service) GetErrors() chan error {
	return s.adapter.GetErrors()
}

func (s *Service) Close() error {
	return s.adapter.Close()
}

// Handling messages for a specific exchange and a topic.
// If you use handler then GetMessages() or Channels()
// don't receive a message for specific exchange and key.
func (s *Service) Handler(exchange string, key string, handler Handler) {
	s.handlers[exchange+":"+key] = handler
}

func (s *Service) Listen() error {
	return s.listen(false)
}

func (s *Service) OnClose() <-chan error {
	return s.onclose
}
