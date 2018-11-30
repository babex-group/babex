package babex

type Channels <-chan *Channel

type Channel struct {
	ch <-chan *Message

	Info map[string]interface{}
}

func NewChannel(ch <-chan *Message) *Channel {
	return &Channel{ch: ch}
}

func (c Channel) GetMessages() <-chan *Message {
	return c.ch
}
