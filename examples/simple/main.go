package main

import (
	"log"
	"encoding/json"

	"github.com/matroskin13/babex"
)

type Config struct {
	Name string `json:"name"`
}

func main() {
	service, err := babex.NewService(&babex.ServiceConfig{
		Address:  "amqp://guest:guest@localhost:5672/",
		Name:     "example-service",
		IsSingle: true,
	})

	if err != nil {
		log.Fatal(err)
	}

	err = service.BindToExchange("x.import", "example")

	if err != nil {
		log.Fatal(err)
	}

	service.Listen(func(message *babex.Message) error {
		var config Config

		data := string(message.Data)

		if err := json.Unmarshal(message.Config, &config); err != nil {
			log.Println("bad json -> ", string(message.Config))
			return err
		}

		log.Printf("key: %s, config.name: %s, data: %s\r\n", message.Key, config.Name, string(message.Data))

		err := service.Next(message, []string{data, data}, nil)

		if err == babex.ErrorNextIsNotDefined {
			log.Println("finish!")
			return nil
		}

		return err
	})
}
