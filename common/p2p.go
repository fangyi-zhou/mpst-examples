package common

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"sync"
)

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

func connectP2PEndpoints(ep1 *P2PEndPoint, ep2 *P2PEndPoint) error {
	if ep1.name == ep2.name {
		return fmt.Errorf("cannot connect two endpoints with same name %s", ep1.name)
	}
	if _, exists := ep1.partners[ep2.name]; exists {
		return fmt.Errorf("cannot connect two connected endpoints, %s and %s", ep1.name, ep2.name)
	}
	if _, exists := ep2.partners[ep1.name]; exists {
		return fmt.Errorf("cannot connect two connected endpoints, %s and %s", ep1.name, ep2.name)
	}
	ep1.partners[ep2.name] = ep2
	ep1.buffer[ep2.name] = make(chan Message, 1)
	ep2.partners[ep1.name] = ep1
	ep2.buffer[ep1.name] = make(chan Message, 1)
	return nil
}

func (e *P2PEndPoint) Connect(other EndPoint) error {
	other_, ok := other.(*P2PEndPoint)
	if ok {
		return connectP2PEndpoints(e, other_)
	}
	return fmt.Errorf("cannot connect P2P endpoint to non-P2P endpoint")
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
		return fmt.Errorf("%s is trying to send a message to an unconnected endpoint %s", e.name, partner)
	}
	e.partners[partner].buffer[e.name] <- message
	return nil
}

func (e *P2PEndPoint) RecvSync(ctx context.Context, partner string) (Message, error) {
	var span trace.Span
	_, span = e.tracer.Start(ctx, "Recv")
	defer span.End()
	if _, exists := e.partners[partner]; !exists {
		return Message{}, fmt.Errorf("%s is trying to send a message to an unconnected endpoint %s", e.name, partner)
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
		return nil, fmt.Errorf("%s is trying to send a message to an unconnected endpoint %s", e.name, partner)
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
