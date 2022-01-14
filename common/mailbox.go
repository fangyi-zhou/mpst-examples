package common

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"sync"
)

type taggedMessage struct {
	Message
	origin string
}

type MailBoxEndPoint struct {
	name     string
	run      func(self EndPoint, group *sync.WaitGroup)
	partners map[string]*MailBoxEndPoint
	tracer   trace.Tracer
	buffer   chan taggedMessage
	received []taggedMessage
}

func MakeMailBoxEndPoint(name string, runFunc func(self EndPoint, group *sync.WaitGroup)) EndPoint {
	return &MailBoxEndPoint{
		name:     name,
		run:      runFunc,
		partners: make(map[string]*MailBoxEndPoint),
		buffer:   make(chan taggedMessage, 1),
		tracer:   otel.Tracer(name),
	}
}

func (m *MailBoxEndPoint) Name() string {
	return m.name
}

func (m *MailBoxEndPoint) Tracer() trace.Tracer {
	return m.tracer
}

func (m *MailBoxEndPoint) Send(ctx context.Context, partner string, message Message) error {
	var span trace.Span
	_, span = m.tracer.Start(ctx, "Send")
	defer span.End()
	span.SetAttributes(
		msgLabelKey.String(message.Label),
		partnerKey.String(partner),
		actionKey.String(actionSend),
		currentRoleKey.String(m.Name()),
	)
	if _, exists := m.partners[partner]; !exists {
		return fmt.Errorf("%s is trying to send a message to an unconnected endpoint %s", m.name, partner)
	}
	m.partners[partner].buffer <- taggedMessage{message, m.name}
	return nil
}

func (m *MailBoxEndPoint) RecvSync(ctx context.Context, partner string) (Message, error) {
	var span trace.Span
	_, span = m.tracer.Start(ctx, "Recv")
	defer span.End()
	if partner == "" {
		// receive from any
		if len(m.received) > 0 {
			msg := m.received[0]
			m.received = m.received[1:]
			span.SetAttributes(
				msgLabelKey.String(msg.Label),
				partnerKey.String(msg.origin),
				actionKey.String(actionRecv),
				currentRoleKey.String(m.Name()),
			)
			return msg.Message, nil
		}
		msg := <-m.buffer
		span.SetAttributes(
			msgLabelKey.String(msg.Label),
			partnerKey.String(msg.origin),
			actionKey.String(actionRecv),
			currentRoleKey.String(m.Name()),
		)
		return msg.Message, nil
	} else {
		// receive from specified partner
		if _, exists := m.partners[partner]; !exists {
			return Message{}, fmt.Errorf("%s is trying to send a message to an unconnected endpoint %s", m.name, partner)
		}
		for idx, msg := range m.received {
			if msg.origin == partner {
				m.received = append(m.received[0:idx], m.received[idx+1:]...)
				span.SetAttributes(
					msgLabelKey.String(msg.Label),
					partnerKey.String(msg.origin),
					actionKey.String(actionRecv),
					currentRoleKey.String(m.Name()),
				)
				return msg.Message, nil
			}
		}
		for {
			newMsg := <-m.buffer
			if newMsg.origin == partner {
				span.SetAttributes(
					msgLabelKey.String(newMsg.Label),
					partnerKey.String(newMsg.origin),
					actionKey.String(actionRecv),
					currentRoleKey.String(m.Name()),
				)
				return newMsg.Message, nil
			} else {
				m.received = append(m.received, newMsg)
			}
		}
	}
}

func (m *MailBoxEndPoint) RecvAsync(ctx context.Context, partner string) (*Message, error) {
	var span trace.Span
	_, span = m.tracer.Start(ctx, "Recv (async)")
	defer span.End()
	if partner == "" {
		// receive from any
		if len(m.received) > 0 {
			msg := m.received[0]
			m.received = m.received[1:]
			span.SetAttributes(
				msgLabelKey.String(msg.Label),
				partnerKey.String(msg.origin),
				actionKey.String(actionRecv),
				currentRoleKey.String(m.Name()),
			)
			return &msg.Message, nil
		}
		select {
		case msg := <-m.buffer:
			span.SetAttributes(
				msgLabelKey.String(msg.Label),
				partnerKey.String(msg.origin),
				actionKey.String(actionRecv),
				currentRoleKey.String(m.Name()),
			)
			return &msg.Message, nil
		default:
			return nil, nil
		}
	} else {
		// receive from specified partner
		if _, exists := m.partners[partner]; !exists {
			return nil, fmt.Errorf("%s is trying to send a message to an unconnected endpoint %s", m.name, partner)
		}
		for idx, msg := range m.received {
			if msg.origin == partner {
				span.SetAttributes(
					msgLabelKey.String(msg.Label),
					partnerKey.String(msg.origin),
					actionKey.String(actionRecv),
					currentRoleKey.String(m.Name()),
				)
				m.received = append(m.received[0:idx], m.received[idx+1:]...)
				return &msg.Message, nil
			}
		}
		for {
			select {
			case newMsg := <-m.buffer:
				if newMsg.origin == partner {
					span.SetAttributes(
						msgLabelKey.String(newMsg.Label),
						partnerKey.String(newMsg.origin),
						actionKey.String(actionRecv),
						currentRoleKey.String(m.Name()),
					)
					return &newMsg.Message, nil
				} else {
					m.received = append(m.received, newMsg)
				}
			default:
				return nil, nil
			}
		}
	}
}

func (m *MailBoxEndPoint) Connect(other EndPoint) error {
	other_, ok := other.(*MailBoxEndPoint)
	if ok {
		return connectMailboxEndpoints(m, other_)
	}
	return fmt.Errorf("cannot connect mailbox `endpoint to non-mailbox endpoint")
}

func connectMailboxEndpoints(ep1 *MailBoxEndPoint, ep2 *MailBoxEndPoint) error {
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
	ep2.partners[ep1.name] = ep1
	return nil
}

func (m *MailBoxEndPoint) Run(group *sync.WaitGroup) {
	m.run(m, group)
}
