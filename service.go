package babex

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/opentracing/opentracing-go"
)

var (
	ErrorNextIsNotDefined = errors.New("next is not defined")
	ErrorNextNoCount      = errors.New("next does not contain count chain")
	ErrorDataIsNotArray   = errors.New("data is not array. next item of chain has isMultiple flag")
	ErrorChainIsEmpty     = errors.New("chain is empty")
	ErrorCloseConsumer    = errors.New("close consumer")
)

type Service struct {
	adapter Adapter
	ch      chan *Message
}

func (s *Service) SetTracer(tracer opentracing.Tracer) {
	opentracing.SetGlobalTracer(tracer)
}

// Create Babex service via the adapter interface
func NewService(adapter Adapter) (*Service, error) {
	service := Service{
		adapter: adapter,
		ch:      make(chan *Message),
	}

	adapterCh, err := service.adapter.GetMessages()
	if err != nil {
		return nil, err
	}

	go func(s *Service, adapterCh <-chan *Message) {
		for msg := range adapterCh {
			s.injectSpan(msg)
			s.ch <- msg
		}
		close(s.ch)
	}(&service, adapterCh)

	return &service, nil
}

func (s *Service) injectSpan(msg *Message) {
	carrier := opentracing.TextMapCarrier(msg.Meta)
	ctx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, carrier)
	if err != nil {
		//no context or error - make new empty span
		msg.Span = opentracing.GlobalTracer().StartSpan("handle")
		return
	}
	msg.Span = opentracing.GlobalTracer().StartSpan(
		"handle",
		opentracing.ChildOf(ctx))
}

// Publish message
func (s *Service) Publish(message InitialMessage) error {
	nextIndex := getCurrentChainIndex(message.Chain)
	if nextIndex == -1 {
		return ErrorNextIsNotDefined
	}

	nextElement := message.Chain[nextIndex]

	meta := Meta{}
	meta.Merge(message.Meta, nextElement.Meta)

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
	defer msg.FinishSpan()
	currentIndex := getCurrentChainIndex(msg.Chain)
	if currentIndex == -1 {
		return ErrorNextIsNotDefined
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

	carrier := opentracing.TextMapCarrier{}
	opentracing.GlobalTracer().Inject(msg.Span.Context(), opentracing.TextMap, carrier)
	for k, v := range carrier {
		m.Meta[k] = v
	}

	err = s.Publish(m)

	return msg.Ack(false)
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
		return nil, ChainItem{}, ErrorNextIsNotDefined
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
	defer msg.FinishSpan()
	err := msg.RawMessage.Ack(true)
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
	carrier := opentracing.TextMapCarrier{}
	opentracing.GlobalTracer().Inject(msg.Span.Context(), opentracing.TextMap, carrier)
	meta.Merge(Meta(carrier))

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

// Get channel for receive messages
func (s *Service) GetMessages() <-chan *Message {
	return s.ch
}

// Get channel for errors
func (s *Service) GetErrors() chan error {
	return s.adapter.GetErrors()
}

func (s *Service) Close() error {
	return s.adapter.Close()
}
