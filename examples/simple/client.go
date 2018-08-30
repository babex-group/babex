package main

import (
	"encoding/json"
	"github.com/matroskin13/babex"
	"github.com/streadway/amqp"
	"log"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.Marshal(babex.InitialMessage{
		Chain: []*babex.ChainItem{
			{
				Exchange: "example",
				Key:      "inc",
			},
			{
				Exchange: "example",
				Key:      "inc",
			},
		},
		Data:   []byte(`{"count": 1}`),
		Config: []byte(`{"incStep": 1}`),
	})

	err = ch.Publish("example", "inc", false, false, amqp.Publishing{
		Body: b,
	})
	if err != nil {
		log.Fatal(err)
	}
}
