package common

import (
	"context"
	"log"
	"sync"

	"github.com/fangyi-zhou/mpst-tracing/labels"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	actionKey   = attribute.Key(labels.ActionKey)
	msgLabelKey = attribute.Key(labels.MsgLabelKey)
	partnerKey  = attribute.Key(labels.PartnerKey)
	actionSend  = "Send"
	actionRecv  = "Recv"
)

type Message struct {
	Label string
	Value interface{}
}

type EndPoint interface {
	Name() string
	Tracer() trace.Tracer
	Send(ctx context.Context, partner string, message Message) error
	Recv(ctx context.Context, partner string) (Message, error)
	Connect(other EndPoint) error
	Run(group *sync.WaitGroup)
}

type QueueEndPoint struct {
	name     string
	run      func(self EndPoint, group *sync.WaitGroup)
	partners map[string]*QueueEndPoint
	buffer   map[string]chan Message
	tracer   trace.Tracer
}

func MakeQueueEndPoint(name string, runFunc func(self EndPoint, group *sync.WaitGroup)) EndPoint {
	return &QueueEndPoint{
		name:     name,
		run:      runFunc,
		partners: make(map[string]*QueueEndPoint),
		buffer:   make(map[string]chan Message),
		tracer:   otel.Tracer(name),
	}
}

func (e *QueueEndPoint) Run(group *sync.WaitGroup) {
	e.run(e, group)
}

func (e *QueueEndPoint) Name() string {
	return e.name
}

func (e *QueueEndPoint) Tracer() trace.Tracer {
	return e.tracer
}

func connectQueueEndpoints(ep1 *QueueEndPoint, ep2 *QueueEndPoint) {
	if ep1.name == ep2.name {
		log.Panicf("Cannot connect two endpoints with same name %s", ep1.name)
	}
	if _, exists := ep1.partners[ep2.name]; exists {
		log.Panicf("Cannot connect two connected endpoints, %s and %s", ep1.name, ep2.name)
	}
	if _, exists := ep2.partners[ep1.name]; exists {
		log.Panicf("Cannot connect two connected endpoints, %s and %s", ep1.name, ep2.name)
	}
	ep1.partners[ep2.name] = ep2
	ep1.buffer[ep2.name] = make(chan Message, 1)
	ep2.partners[ep1.name] = ep1
	ep2.buffer[ep1.name] = make(chan Message, 1)
}

func (e *QueueEndPoint) Connect(other EndPoint) error {
	connectQueueEndpoints(e, other.(*QueueEndPoint))
	return nil
}

func (e *QueueEndPoint) Send(ctx context.Context, partner string, message Message) error {
	var span trace.Span
	_, span = e.tracer.Start(ctx, "Send")
	defer span.End()
	span.SetAttributes(
		msgLabelKey.String(message.Label),
		partnerKey.String(partner),
		actionKey.String(actionSend),
	)
	if _, exists := e.partners[partner]; !exists {
		log.Panicf("%s is trying to send a message to an unconnected endpoint %s", e.name, partner)
	}
	e.partners[partner].buffer[e.name] <- message
	return nil
}

func (e *QueueEndPoint) Recv(ctx context.Context, partner string) (Message, error) {
	var span trace.Span
	_, span = e.tracer.Start(ctx, "Recv")
	defer span.End()
	if _, exists := e.partners[partner]; !exists {
		log.Panicf("%s is trying to send a message to an unconnected endpoint %s", e.name, partner)
	}
	message := <-e.buffer[partner]
	span.SetAttributes(
		msgLabelKey.String(message.Label),
		partnerKey.String(partner),
		actionKey.String(actionRecv),
	)
	return message, nil
}
