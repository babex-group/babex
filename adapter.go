package babex

type Adapter interface {
	GetMessages() (<-chan *Message, error)
	GetErrors() chan error
	Publish(exchange string, key string, message InitialMessage) error
	Close() error
}
