package main

import (
	"log"

	"github.com/matroskin13/babex"
	"encoding/json"
	"github.com/matroskin13/babex/adapters/rabbit"
)

func main() {
	adapter, err := rabbit.NewAdapter(rabbit.Options{
		Address: "amqp://guest:guest@localhost:5672/",
		Name:    "inc-service",
	})
	if err != nil {
		log.Fatal(err)
	}

	service := babex.NewService(adapter)

	err = adapter.BindToExchange("example", "get-numbers")
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := service.GetMessages()
	if err != nil {
		log.Fatal(err)
	}

	errChan := service.GetErrors()

	for {
		select {
		case msg := <-msgs:
			data := struct {
				A int `json:"a"`
				B int `json:"b"`
			}{}

			if err := json.Unmarshal(msg.Data, &data); err != nil {
				msg.Ack(false)
				break
			}

			req := struct {
				Sum int `json:"sum"`
			}{}

			req.Sum = data.A + data.B

			log.Printf("A = %v, B = %v, Sum = %v \r\n", data.A, data.B, req.Sum)

			service.Next(msg, req, nil)
		case err := <-errChan:
			log.Fatal("err", err)
		}
	}
}
