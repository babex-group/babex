package main

import (
	"log"

	"github.com/matroskin13/babex"
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

			data.A = 3
			data.B = 7

			log.Printf("A = %v, B = %v \r\n", data.A, data.B)

			service.Next(msg, data, nil)
		case err := <-errChan:
			log.Fatal("err", err)
		}
	}
}
