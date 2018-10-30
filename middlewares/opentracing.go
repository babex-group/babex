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

	//if no error after context extraction - use it
	if err == nil {
		msg.Span = m.Tracer.StartSpan(m.handlerName, opentracing.ChildOf(ctx))
	} else {
		msg.Span = m.Tracer.StartSpan(m.handlerName)
	}

	return func(err error) {
		m.addCarrier(msg)
		msg.Span.Finish()
	}, nil
}

func (m OpentracingMiddleware) addCarrier(msg *babex.Message) {
	carrier := opentracing.TextMapCarrier{}
	m.Tracer.Inject(msg.Span.Context(), opentracing.TextMap, carrier)
	for k, v := range carrier {
		// fill into meta for the Next method
		msg.Meta[k] = v

		// fill into intial message meta for the Catch method
		msg.InitialMessage.Meta[k] = v
	}
}
