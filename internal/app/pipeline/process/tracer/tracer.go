package tracer

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/banzaicloud/pipeline/internal/app/pipeline/process/client"
	processClient "github.com/banzaicloud/pipeline/internal/app/pipeline/process/client"
)

const (
	workflowTag = "cadenceWorkflowID"

	runTag = "cadenceRunID"
)

type processTracer struct {
	client *processClient.Client
}

var _ opentracing.Tracer = processTracer{}

func NewProcessTracer(address string) (opentracing.Tracer, error) {
	client, err := processClient.NewClient(processClient.Config{Address: address})
	if err != nil {
		return nil, err
	}

	return &processTracer{client: client}, nil
}

func (t processTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	span := processSpan{
		tracer: &t,
	}

	options := opentracing.StartSpanOptions{}
	for _, opt := range opts {
		opt.Apply(&options)
	}

	span.activity = isActivity(options)

	// do we have a parent?
	if len(options.References) > 0 {
		fmt.Printf("------ yes we have a parent: %+v\n", options.References)
		reference := options.References[0]
		if reference.Type == opentracing.FollowsFromRef {
			parentContext := reference.ReferencedContext.(processSpanContext)
			span.process.ParentID = parentContext.span.process.ID
		}
	}

	if !span.activity {
		span.process.ID = options.Tags[workflowTag].(string)
		span.process.Name = operationName
		span.process.StartedAt = options.StartTime
		span.process.Status = client.Running
		span.process.ResourceType = client.Cluster
		span.process.OrgID = 1 // TODO

		err := t.client.LogProcess(context.Background(), span.process)
		if err != nil {
			println("----------- failed to start span:", err.Error())
		}

		fmt.Printf("------------ started span: %+v\n", span.process)
	} else {
		span.event.ProcessID = options.Tags[workflowTag].(string)
		span.event.Name = operationName
		span.event.Log = operationName + " has started"
		span.event.Timestamp = options.StartTime

		err := t.client.LogEvent(context.Background(), span.event)
		if err != nil {
			println("----------- failed to start span:", err.Error())
		}

		fmt.Printf("------------ started span: %+v\n", span.event)

	}

	return &span
}

func isActivity(options opentracing.StartSpanOptions) bool {
	_, ok := options.Tags[runTag]
	return ok
}

func (t processTracer) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	fmt.Printf("----------- processTracer inject format: %+v carrier: %+v\n", format, carrier)
	return nil
}

type tracingReader interface {
	ForeachKey(handler func(key, val string) error) error
}

func (t processTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	fmt.Printf("----------- processTracer extract format: %+v carrier: %+v\n", format, carrier)

	if format.(opentracing.BuiltinFormat) == opentracing.HTTPHeaders {
		if reader, ok := carrier.(tracingReader); ok {
			err := reader.ForeachKey(func(key, val string) error {
				println("Extract")
				println(key, " = ", val)
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			println("nem nyert a cast")
		}
	} else if format.(opentracing.BuiltinFormat) == opentracing.TextMap {
		if reader, ok := carrier.(tracingReader); ok {
			err := reader.ForeachKey(func(key, val string) error {
				println("Extract")
				println(key, " = ", val)
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			println("nem nyert a cast")
		}
	}

	return nil, nil
}

type processSpanContext struct {
	span *processSpan
}

func (n processSpanContext) ForeachBaggageItem(handler func(k, v string) bool) {}

type processSpan struct {
	tracer   *processTracer
	process  processClient.ProcessEntry
	event    processClient.ProcessEvent
	activity bool
}

func (n processSpan) Context() opentracing.SpanContext                      { return processSpanContext{span: &n} }
func (n processSpan) SetBaggageItem(key, val string) opentracing.Span       { return n }
func (n processSpan) BaggageItem(key string) string                         { return "" }
func (n processSpan) SetTag(key string, value interface{}) opentracing.Span { return n }
func (n processSpan) LogFields(fields ...log.Field)                         {}
func (n processSpan) LogKV(keyVals ...interface{})                          {}
func (n processSpan) Finish() {
	finishedAt := time.Now()
	if !n.activity {
		n.process.FinishedAt = &finishedAt
		n.process.Status = client.Finished // TODO
		err := n.tracer.client.LogProcess(context.Background(), n.process)
		if err != nil {
			println("----------- failed to finish span:", err.Error())
		}
	} else {
		n.event.Timestamp = finishedAt
		n.event.Log = n.event.Name + " has finished"
		err := n.tracer.client.LogEvent(context.Background(), n.event)
		if err != nil {
			println("----------- failed to finish span:", err.Error())
		}
	}
}
func (n processSpan) FinishWithOptions(opts opentracing.FinishOptions)       {}
func (n processSpan) SetOperationName(operationName string) opentracing.Span { return n }
func (n processSpan) Tracer() opentracing.Tracer                             { return n.tracer }
func (n processSpan) LogEvent(event string)                                  {}
func (n processSpan) LogEventWithPayload(event string, payload interface{})  {}
func (n processSpan) Log(data opentracing.LogData)                           {}