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
