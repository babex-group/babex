# Babex

[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/matroskin13/babex)

The Babex allows you to define a chain before the start of processing by all services. In other words, you service does not have a constant for the next element.

The way without the Babex (Pseudocode):

```go
msg := GetMessageFromTopic("first-topic")
// do something
SendMessageToTopic("second-topic", newMessage)
```

The way with the Babex:

Initialize initial chain

```json
{
  "chain":[
    {"topic": "first-topic"},
    {"topic": "second-topic"}
  ]
}
```

And create service:

```go
msg := GetMessageFromTopic("first-topic")
// do something
SendMesssageToNext(newMessage)
```

## Docs

- [About the Babex](docs/protocol.md)
- [Receiving messages](docs/receiving.md)
- [Middleware](docs/middleware.md)
- [Examples](https://github.com/babex-group/examples)

## Adapters

- [Kafka](https://github.com/babex-group/babex-kafka)
- [Rabbit](https://github.com/babex-group/babex-rabbit)

## Usage

For example, we create service which will add the number to counter. We will use the Kafka adapter.

```go
package main

import (
	"github.com/babex-group/babex"
	"github.com/babex-group/babex-kafka"
	"log"
	"os"
	"os/signal"
	"encoding/json"
)

func main() {
	a, err := kafka.NewAdapter(kafka.Options{
		Name:   "babex-sandbox"
		Topics: []string{"example-topic"},
		Addrs:  []string{"localhost:29092"},
	})
	if err != nil {
		log.Fatal(err)
	}

	s := babex.NewService(a)

	defer s.Close()

	s.Handler("example-topic", "", func(msg *babex.Message) error {
		var data struct{
			Count int `json:"count"`
		}

		if err := json.Unmarshal(msg.Data, &data); err != nil {
			msg.Ack() // The message has invalid format, skip it.
			return err
		}

		data.Count += 1

		fmt.Printf("count = %v\r\n", data.Count)

		return s.Next(msg, data, nil) // Next automatically use msg.Ack()
	})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	<- signals
}
```

And publish the message to the topic "babex-sandbox":

```json
{
  "data": {
    "count": 0
  },
  "chain": [
    {
      "exchange": "example-topic"
    }
  ]
}
```

Check logs:

```bash
$ count = 1
```

Excellent! Let's change the message:

```json
{
  "data": {
    "count": 0
  },
  "chain": [
    {
      "exchange": "example-topic"
    },
    {
      "exchange": "example-topic"
    }
  ]
}
```

Check logs:

```bash
$ count = 1
$ count = 2
```

The service receive two message via chain, and increment the count.
