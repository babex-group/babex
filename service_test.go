package babex

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestService_Receive(t *testing.T) {
	ch := make(chan *Message)

	stub := StubAdapter{OutCh: ch}
	s := NewServiceListener(&stub)

	expectedMsg := &Message{Exchange: "test", Key: "test-key"}

	ch <- expectedMsg

	select {
	case msg := <-s.GetMessages():
		assert.Equal(t, expectedMsg, msg)
	case <-time.After(time.Microsecond * 100):
		panic("cannot read message")
	}

	expectedMsg1 := &Message{Exchange: "test1", Key: "test-key1"}

	ch <- expectedMsg1

	select {
	case msg := <-s.GetMessages():
		assert.Equal(t, expectedMsg1, msg)
	case <-time.After(time.Microsecond * 1):
		panic("cannot read message")
	}
}

func TestService_ApplyMiddleware(t *testing.T) {
	var isUse bool
	var isFinish bool

	ch := make(chan *Message)

	middleware := StubMiddleware{
		OnUse: func(s *Service, msg *Message) (MiddlewareDone, error) {
			isUse = true

			return func(err error) {
				isFinish = true
			}, nil
		},
	}

	stub := StubAdapter{OutCh: ch}
	s := NewServiceListener(&stub, middleware)

	expectedMsg := &Message{
		Exchange:   "test",
		Key:        "test-key",
		RawMessage: StubMessage{},
	}

	ch <- expectedMsg

	select {
	case msg := <-s.GetMessages():
		assert.Equal(t, expectedMsg, msg)

		s.Next(msg, nil, nil)
	case <-time.After(time.Microsecond * 1):
		panic("cannot read message")
	}

	assert.True(t, isUse)
	assert.True(t, isFinish)
}

func TestService_Handler(t *testing.T) {
	ch := make(chan *Message)
	done := make(chan *Message)

	stub := StubAdapter{OutCh: ch}
	s := NewService(&stub)

	s.Handler("test", "test-key", func(msg *Message) error {
		done <- msg
		return nil
	})

	go s.Listen()

	expectedMsg := &Message{Exchange: "test", Key: "test-key"}

	ch <- expectedMsg

	select {
	case msg := <-done:
		assert.Equal(t, expectedMsg, msg)
	case <-time.After(time.Microsecond * 1):
		panic("cannot read message")
	}
}

func TestService_Publish(t *testing.T) {
	chain := Chain{
		{
			Exchange:   "first",
			Key:        "first-key",
			Successful: true,
		},
		{
			Exchange: "second",
			Key:      "second-key",
		},
		{
			Exchange: "third",
			Key:      "third-key",
		},
	}

	stub := StubAdapter{
		onPublish: func(exchange string, key string, message InitialMessage) error {
			assert.Equal(t, "second", exchange)
			assert.Equal(t, "second-key", key)
			return nil
		},
	}

	s := NewServiceListener(&stub)

	s.Publish(InitialMessage{
		Chain: chain,
	})
}

func TestService_PublishWithWhen(t *testing.T) {
	chain := Chain{
		{
			Exchange:   "first",
			Key:        "first-key",
			Successful: true,
		},
		{
			Exchange: "second",
			Key:      "second-key",
			When: When{
				"$meta.advertStatus": []string{"1", "2", "3"},
			},
		},
		{
			Exchange: "third",
			Key:      "third-key",
		},
	}

	stub := StubAdapter{
		onPublish: func(exchange string, key string, message InitialMessage) error {
			assert.Equal(t, "third", exchange)
			assert.Equal(t, "third-key", key)
			return nil
		},
	}

	s := NewServiceListener(&stub)

	s.Publish(InitialMessage{
		Chain: chain,
		Meta:  Meta{"advertStatus": "4"},
	})
}
