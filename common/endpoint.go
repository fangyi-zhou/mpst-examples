package common

import (
	"go.opentelemetry.io/otel/trace"
	"log"
	"sync"
)

type RunFunc = func(group *sync.WaitGroup)

type Message struct {
	Label string
	Value interface{}
}

type EndPoint struct {
	Name     string
	Run      RunFunc
	partners map[string]*EndPoint
	buffer   map[string]chan Message
	tracer   trace.Tracer
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

func (e *EndPoint) Send(partner string, message Message) {
	if _, exists := e.partners[partner]; !exists {
		log.Panicf("%s is trying to send a message to an unconnected endpoint %s", e.Name, partner)
	}
	e.partners[partner].buffer[e.Name] <- message
}

func (e *EndPoint) Recv(partner string) Message {
	if _, exists := e.partners[partner]; !exists {
		log.Panicf("%s is trying to send a message to an unconnected endpoint %s", e.Name, partner)
	}
	return <-e.buffer[partner]
}
