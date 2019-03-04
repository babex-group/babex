package babex

type MiddlewareDone func(err error)

type Middleware interface {
	Use(s *Service, m *Message) (MiddlewareDone, error)
}
