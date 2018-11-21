package babex

type MiddlewareDone func(err error)

type Middleware interface {
	Use(m *Message) (MiddlewareDone, error)
}
