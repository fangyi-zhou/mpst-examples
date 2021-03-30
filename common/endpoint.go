package common

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"log"
	"sync"
)

var (
	actionKey   = attribute.Key("mpst/action")
	msgLabelKey = attribute.Key("mpst/msgLabel")
	partnerKey  = attribute.Key("mpst/partner")
	actionSend  = "Send"
	actionRecv  = "Recv"
)

type Message struct {
	Label string
	Value interface{}
}

type EndPoint struct {
	Name     string
	run      func(self *EndPoint, group *sync.WaitGroup)
	partners map[string]*EndPoint
	buffer   map[string]chan Message
	Tracer   trace.Tracer
}

func MakeEndPoint(name string, runFunc func(self *EndPoint, group *sync.WaitGroup)) *EndPoint {
	return &EndPoint{
		Name:     name,
		run:      runFunc,
		partners: make(map[string]*EndPoint),
		buffer:   make(map[string]chan Message),
		Tracer:   otel.Tracer(name),
	}
}

func connectEndpoints(ep1 *EndPoint, ep2 *EndPoint) {
	if ep1.Name == ep2.Name {
		log.Panicf("Cannot connect two endpoints with same name %s", ep1.Name)
	}
	if _, exists := ep1.partners[ep2.Name]; exists {
		log.Panicf("Cannot connect two connected endpoints, %s and %s", ep1.Name, ep2.Name)
	}
	if _, exists := ep2.partners[ep1.Name]; exists {
		log.Panicf("Cannot connect two connected endpoints, %s and %s", ep1.Name, ep2.Name)
	}
	ep1.partners[ep2.Name] = ep2
	ep1.buffer[ep2.Name] = make(chan Message, 1)
	ep2.partners[ep1.Name] = ep1
	ep2.buffer[ep1.Name] = make(chan Message, 1)
}

func (e *EndPoint) Send(ctx context.Context, partner string, message Message) {
	var span trace.Span
	_, span = e.Tracer.Start(ctx, "Send")
	defer span.End()
	span.SetAttributes(msgLabelKey.String(message.Label), partnerKey.String(partner), actionKey.String(actionSend))
	if _, exists := e.partners[partner]; !exists {
		log.Panicf("%s is trying to send a message to an unconnected endpoint %s", e.Name, partner)
	}
	e.partners[partner].buffer[e.Name] <- message
}

func (e *EndPoint) Recv(ctx context.Context, partner string) Message {
	var span trace.Span
	_, span = e.Tracer.Start(ctx, "Recv")
	defer span.End()
	if _, exists := e.partners[partner]; !exists {
		log.Panicf("%s is trying to send a message to an unconnected endpoint %s", e.Name, partner)
	}
	message := <-e.buffer[partner]
	span.SetAttributes(msgLabelKey.String(message.Label), partnerKey.String(partner), actionKey.String(actionRecv))
	return message
}
