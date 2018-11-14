package babex

type StubMessage struct{}

// Use stub message for Catch without ack/nack
func NewStubMessage(initialMessage *InitialMessage, exchange string, key string) *Message {
	msg := NewMessage(initialMessage, exchange, key)

	msg.RawMessage = StubMessage{}

	return msg
}

func (sm StubMessage) Ack() error {
	return nil
}

func (sm StubMessage) Nack() error {
	return nil
}

type StubAdapter struct {
	OutCh       chan *Message
	OutChannels Channels
	Err         chan error

	onPublish func(exchange string, key string, message InitialMessage) error
}

func (stub *StubAdapter) GetMessages() (<-chan *Message, error) {
	return stub.OutCh, nil
}

func (stub *StubAdapter) GetErrors() chan error {
	return stub.Err
}

func (stub *StubAdapter) Publish(exchange string, key string, message InitialMessage) error {
	return stub.onPublish(exchange, key, message)
}

func (stub *StubAdapter) Close() error {
	return nil
}

func (stub *StubAdapter) Channels() Channels {
	return stub.OutChannels
}

type StubMiddleware struct {
	OnUse func(msg *Message) (MiddlewareDone, error)
}

func (stubMiddleware StubMiddleware) Use(msg *Message) (MiddlewareDone, error) {
	return stubMiddleware.OnUse(msg)
}
