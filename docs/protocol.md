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

- chain - 
- data 
- config
- meta
- catch
