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
	actionKey      = attribute.Key(labels.ActionKey)
	msgLabelKey    = attribute.Key(labels.MsgLabelKey)
	partnerKey     = attribute.Key(labels.PartnerKey)
	currentRoleKey = attribute.Key(labels.CurrentRoleKey)
	actionSend     = "Send"
	actionRecv     = "Recv"
)

type Message struct {
	Label string
	Value interface{}
}

type EndPoint interface {
	Name() string
	Tracer() trace.Tracer
	Send(ctx context.Context, partner string, message Message) error
	RecvSync(ctx context.Context, partner string) (Message, error)
	// a pointer is a lay-person's option type (LOL)
	RecvAsync(ctx context.Context, partner string) (*Message, error)
	Connect(other EndPoint) error
	Run(group *sync.WaitGroup)
}

type P2PEndPoint struct {
	name     string
	run      func(self EndPoint, group *sync.WaitGroup)
	partners map[string]*P2PEndPoint
	buffer   map[string]chan Message
	tracer   trace.Tracer
}

func MakeP2PEndPoint(name string, runFunc func(self EndPoint, group *sync.WaitGroup)) EndPoint {
	return &P2PEndPoint{
		name:     name,
		run:      runFunc,
		partners: make(map[string]*P2PEndPoint),
		buffer:   make(map[string]chan Message),
		tracer:   otel.Tracer(name),
	}
}

func (e *P2PEndPoint) Run(group *sync.WaitGroup) {
	e.run(e, group)
}

func (e *P2PEndPoint) Name() string {
	return e.name
}

func (e *P2PEndPoint) Tracer() trace.Tracer {
	return e.tracer
}

func connectP2PEndpoints(ep1 *P2PEndPoint, ep2 *P2PEndPoint) {
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

func (e *P2PEndPoint) Connect(other EndPoint) error {
	connectP2PEndpoints(e, other.(*P2PEndPoint))
	return nil
}

func (e *P2PEndPoint) Send(ctx context.Context, partner string, message Message) error {
	var span trace.Span
	_, span = e.tracer.Start(ctx, "Send")
	defer span.End()
	span.SetAttributes(
		msgLabelKey.String(message.Label),
		partnerKey.String(partner),
		actionKey.String(actionSend),
		currentRoleKey.String(e.Name()),
	)
	if _, exists := e.partners[partner]; !exists {
		log.Panicf("%s is trying to send a message to an unconnected endpoint %s", e.name, partner)
	}
	e.partners[partner].buffer[e.name] <- message
	return nil
}

func (e *P2PEndPoint) RecvSync(ctx context.Context, partner string) (Message, error) {
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
		currentRoleKey.String(e.Name()),
	)
	return message, nil
}
func (e *P2PEndPoint) RecvAsync(ctx context.Context, partner string) (*Message, error) {
	var span trace.Span
	_, span = e.tracer.Start(ctx, "Recv (async)")
	defer span.End()
	if _, exists := e.partners[partner]; !exists {
		log.Panicf("%s is trying to send a message to an unconnected endpoint %s", e.name, partner)
	}
	select {
	case message := <-e.buffer[partner]:
		span.SetAttributes(
			msgLabelKey.String(message.Label),
			partnerKey.String(partner),
			actionKey.String(actionRecv),
			currentRoleKey.String(e.Name()),
		)
		return &message, nil
	default:
		return nil, nil
	}
}
