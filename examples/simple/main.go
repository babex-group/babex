package main

import (
	"log"

	"github.com/matroskin13/babex"
)

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
		data := string(message.Data)

		log.Println(message.Key, string(message.Data))

		err := service.Next(message, []string{data, data})

		if err == babex.ErrorNextIsNotDefined {
			log.Println("finish!")
			return nil
		}

		return err
	})
}
