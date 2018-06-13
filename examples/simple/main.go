package main

import (
	"encoding/json"
	"log"

	"github.com/matroskin13/babex"
)

func main() {
	service, err := babex.NewService(&babex.ServiceConfig{
		Address:  "amqp://guest:guest@localhost:5672/",
		Name:     "inc-service",
	})
	if err != nil {
		log.Fatal(err)
	}

	err = service.BindToExchange("example", "inc")
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

			service.Next(msg, data, nil)
		case err := <-errChan:
			log.Fatal("err", err)
		}
	}
}
