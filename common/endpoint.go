package common

import (
	"context"
	"sync"

	"github.com/fangyi-zhou/mpst-tracing/labels"
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
