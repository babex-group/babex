package middlewares

import (
	"github.com/babex-group/babex"

	"github.com/opentracing/opentracing-go"
)

const (
	spanHandlerName = "handle"
)

type OpentracingMiddleware struct {
	Tracer *opentracing.Tracer
}

func (m OpentracingMiddleware) Use(msg *babex.Message) (babex.MiddlewareDone, error) {
	carrier := opentracing.TextMapCarrier(msg.Meta)
	ctx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, carrier)
	if err != nil {
		//no context or error - make new empty span
		msg.Span = opentracing.GlobalTracer().StartSpan("handle")
		return func(err error) {}, nil
	}
	msg.Span = opentracing.GlobalTracer().StartSpan(
		spanHandlerName,
		opentracing.ChildOf(ctx))

	return func(err error) {}, nil
}
