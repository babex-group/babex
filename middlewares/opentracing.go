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

func (m OpentracingMiddleware) Use(s *babex.Service, msg *babex.Message) (babex.MiddlewareDone, error) {
	carrier := opentracing.TextMapCarrier(msg.Meta)
	ctx, err := m.Tracer.Extract(opentracing.TextMap, carrier)

	//if no error after context extraction - use it
	if err == nil {
		msg.Span = m.Tracer.StartSpan(m.handlerName, opentracing.ChildOf(ctx))
	} else {
		msg.Span = m.Tracer.StartSpan(m.handlerName)
	}

	return func(err error) {
		m.Tracer.Inject(msg.Span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(msg.Meta))
		msg.Span.Finish()
	}, nil
}
