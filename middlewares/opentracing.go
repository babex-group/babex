package middlewares

import (
	"github.com/babex-group/babex"

	"github.com/opentracing/opentracing-go"
)

const (
	spanHandlerName = "handle"
)

type OpentracingMiddleware struct {
	Tracer      opentracing.Tracer
	handlerName string
}

type OpentracingOptions struct {
	Name string
}

func NewOpentracing(tracer opentracing.Tracer) OpentracingMiddleware {
	return OpentracingMiddleware{
		Tracer:      tracer,
		handlerName: spanHandlerName,
	}
}

func NewOpentracingWithOpts(tracer opentracing.Tracer, opts OpentracingOptions) OpentracingMiddleware {
	return OpentracingMiddleware{
		Tracer:      tracer,
		handlerName: opts.Name,
	}
}

func (m OpentracingMiddleware) Use(msg *babex.Message) (babex.MiddlewareDone, error) {
	carrier := opentracing.TextMapCarrier(msg.Meta)
	ctx, err := m.Tracer.Extract(opentracing.TextMap, carrier)
	var opt opentracing.StartSpanOption

	//if no error after context extraction - use it
	if err == nil {
		opt = opentracing.ChildOf(ctx)
	}

	msg.Span = m.Tracer.StartSpan(m.handlerName, opt)

	return func(err error) {
		msg.Span.Finish()
	}, nil
}