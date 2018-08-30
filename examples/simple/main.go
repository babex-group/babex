package main

import (
	"encoding/json"
	"log"

	"github.com/matroskin13/babex"
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

	err = adapter.BindToExchange("example", "inc")
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
				Count int `json:"count"`
			}{}

			if err := json.Unmarshal(msg.Data, &data); err != nil {
				log.Println(err)
				break
			}

			cfg := struct {
				IncStep int `json:"incStep"`
			}{}

			if err := json.Unmarshal(msg.Config, &cfg); err != nil {
				log.Println(err)
				break
			}

			data.Count += cfg.IncStep

			log.Printf("count = %v, incStep = %v \r\n", data.Count, cfg.IncStep)

			err := service.Next(msg, data, nil)
			if err == babex.ErrorNextIsNotDefined {
				log.Println("chain is over")
			} else if err != nil {
				log.Println("cannot publish", err)
			} else {
				log.Println("publish next")
			}
		case err := <-errChan:
			log.Fatal("err", err)
		}
	}
}
