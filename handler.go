package babex

type Handlers map[string]Handler

type Handler func(msg *Message) error
