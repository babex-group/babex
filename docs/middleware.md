## Middleware

You can use middleware for your service. It is very convenient for logging, modifying message, and etc.

### Default middlewares:

- [Opentracing](../middlewares/opentracing.go)

### Custom middleware:

For example we may to implement logger:

```go
type loggerMiddleware struct{}

func (l loggerMiddleware) Use(msg *babex.Message) (babex.MiddlewareDone, error) {
	fmt.Printf("logger: receive message to %s\r\n", msg.Exchange)

	start := time.Now()

	return func(err error) {
		fmt.Printf("logger: done message. time: %s, err: %v \r\n", time.Since(start), err)
	}, nil
}
```

And use it:

```go
s := babex.NewService(adapter, loggerMiddleware{})
```

Now when the service receives a message from the adapter, you will see the logs:

```
logger: receive message to babex-sandbox
logger: done message. time: 9.043µs, err: <nil>
```