package babex

type Logger interface {
	Log(message string)
}

type StubLogger struct{}

func (s *StubLogger) Log(message string) {}
