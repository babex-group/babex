## Receiving messages

You can receive messages in two ways:

- Via the Handler method
- Via channels

The Handler way is preferring. Because the handler callback has the error tuple.

### Handler

```go
func main() {
	a, err := kafka.NewAdapter(kafka.Options{
		Name:   "babex-sandbox",
		Topics: []string{"example-topic"},
		Addrs:  []string{"localhost:29092"},
	})
	if err != nil {
		log.Fatal(err)
	}

	s := babex.NewService(a)

	defer s.Close()

	go func() {
		for err := range s.GetErrors() {
			fmt.Printf("receive babex error. err = %s", err)
		}
	}()

	s.Handler("example-topic", "", func(msg *babex.Message) error {
		fmt.Printf("receive message. exchange = %s", msg.Exchange)

		return s.Next(msg, nil, nil)
	})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	<-signals
}
```

### Channels

```go
func main() {
	a, err := kafka.NewAdapter(kafka.Options{
		Name:   "babex-sandbox",
		Topics: []string{"example-topic"},
		Addrs:  []string{"localhost:29092"},
	})
	if err != nil {
		log.Fatal(err)
	}

	s := babex.NewServiceListener(a)

	defer s.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for {
		select {
		case <-signals:
			return
		case msg := <-s.GetMessages():
			fmt.Printf("receive message. exchange = %s", msg.Exchange)

			s.Next(msg, nil, nil)
		case err := <-s.GetErrors():
			fmt.Printf("receive babex error. err = %s", err)
		}
	}
}
```