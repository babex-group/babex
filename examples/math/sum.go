package main

import (
	"log"

	"github.com/matroskin13/babex"
	"encoding/json"
)

func main() {
	service, err := babex.NewService(&babex.ServiceConfig{
		Address:  "amqp://guest:guest@localhost:5672/",
		Name:     "numbers",
	})
	if err != nil {
		log.Fatal(err)
	}

	err = service.BindToExchange("example", "get-numbers")
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
