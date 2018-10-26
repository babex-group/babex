## Protocol

The babex message has JSON format. The message includes the following properties:

### Chain

The chain is an array of objects including exchange and key properties.

For the Kafka exchange is a *topic*. For the Rabbit exchange is a *exchange*.

In runtime elements of chain follow each other via call the `service.Next(...)` method. Every service in chain doesn't know about next element. It's very flexible for the building pipeline of services.

```json
{
  "chain": [
    {"exchange": "first-topic"},
    {"exchange": "second-topic"}
  ]
}
```

In the above example, the following actions will be performed:

1. The message will be sent to the `first-topic`.
2. The service listening the `first-topic`, will receive the message, and process it. After,  the service calls the `service.Next(...)` method, and send message to `second-topic`.
3. The service listening the `second-topic`, will receive the message, and process it. After, the service will not get the next element, and it will not do anything.

### Data

Data is any json property. Each service of chain can modify the data, and pass to next service.

```go
var data struct{
  Count int
}

if err := json.Unmarshal(msg.Data, &data); err != nil {
  msg.Ack()
  return err
}

data.Count += 1

fmt.Printf("count = %v\r\n", data.Count)

return s.Next(msg, data, nil)
```

Chain for example:

```json
{
  "chain": [
    {"exchange": "first-topic"},
    {"exchange": "second-topic"}
  ],
  "data": {
    "count": 1
  }
}
```

### Config

Config is any JSON property too. But the config is used in each service of chain.

```go
var data struct{
  Count int
}

if err := json.Unmarshal(msg.Data, &data); err != nil {
  msg.Ack()
  return err
}

var cfg struct{
  Step int
}

if err := json.Unmarshal(msg.Config, &cfg); err != nil {
  msg.Ack()
  return err
}

data.Count += cfg.Step

fmt.Printf("count = %v\r\n", data.Count)

return s.Next(msg, data, nil)
```

Chain for example:

```json
{
  "chain": [
    {"exchange": "first-topic"},
    {"exchange": "second-topic"}
  ],
  "data": {
    "count": 1
  },
  "config": {
    "step": 1
  }
}
```

### Meta

Meta is `map[string]string` property. It's like config, but config for users, meta for libraries. For example the Babex uses meta for the opentracing `TextMapCarrier`.

