package twobuyer

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/fangyi-zhou/mpst-examples/common"
	"go.opentelemetry.io/otel/trace"
)

func runA(self common.EndPoint, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	var span trace.Span
	ctx, span = self.Tracer().Start(ctx, "TwoBuyer Endpoint A")
	defer span.End()
	// Send query to S
	var query = rand.Intn(100)
	fmt.Println("A: Sending query", query)
	self.Send(ctx, "S", common.Message{Label: "query", Value: query})
	// Receive a quote
	var quoteMsg, _ = self.RecvSync(ctx, "S")
	var quote = quoteMsg.Value.(int)
	var otherShareMsg, _ = self.RecvSync(ctx, "B")
	var otherShare = otherShareMsg.Value.(int)
	if otherShare*2 >= quote {
		// 1 stands for ok
		self.Send(ctx, "S", common.Message{Label: "buy", Value: 1})
	} else {
		self.Send(ctx, "S", common.Message{Label: "buy", Value: 0})
	}
}

func runABad(self common.EndPoint, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	var span trace.Span
	ctx, span = self.Tracer().Start(ctx, "TwoBuyer Endpoint A")
	defer span.End()
	// Send query to S
	var query = rand.Intn(100)
	fmt.Println("A: Sending query", query)
	self.Send(ctx, "S", common.Message{Label: "query", Value: query})
	// Receive a quote
	var quoteMsg, _ = self.RecvSync(ctx, "S")
	var quote = quoteMsg.Value.(int)
	var otherShareMsg, _ = self.RecvSync(ctx, "B")
	var otherShare = otherShareMsg.Value.(int)
	if otherShare*2 >= quote {
		// 1 stands for ok
		self.Send(ctx, "S", common.Message{Label: "purchase", Value: 1})
	} else {
		self.Send(ctx, "S", common.Message{Label: "purchase", Value: 0})
	}
}

func runS(self common.EndPoint, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	var span trace.Span
	ctx, span = self.Tracer().Start(ctx, "TwoBuyer Endpoint S")
	defer span.End()
	// Receive a query
	var queryMsg, _ = self.RecvSync(ctx, "A")
	var query = queryMsg.Value.(int)
	// Send a quote
	var quote = query * 2
	fmt.Println("S: Sending quote", quote)
	self.Send(ctx, "A", common.Message{Label: "quoteA", Value: quote})
	self.Send(ctx, "B", common.Message{Label: "quoteB", Value: quote})
	var decisionMsg, _ = self.RecvSync(ctx, "A")
	var decision = decisionMsg.Value.(int)
	if decision == 1 {
		fmt.Println("Succeed!")
	} else {
		fmt.Println("Failed to succeed!")
	}
}

func runB(self common.EndPoint, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	var span trace.Span
	ctx, span = self.Tracer().Start(ctx, "TwoBuyer Endpoint B")
	defer span.End()
	// Receive a quote
	var quoteMsg, _ = self.RecvSync(ctx, "S")
	var quote = quoteMsg.Value.(int)
	// Propose a share
	var share = quote/2 + rand.Intn(10) - 5
	fmt.Println("B: Proposing share", share)
	self.Send(ctx, "A", common.Message{Label: "share", Value: share})
}

func MakeEndpoints() []common.EndPoint {
	a := common.MakeP2PEndPoint("A", runA)
	b := common.MakeP2PEndPoint("B", runB)
	s := common.MakeP2PEndPoint("S", runS)
	return []common.EndPoint{a, b, s}
}

func RunAll() {
	common.RunEndpoints(common.InitOtlpTracer, MakeEndpoints())
}

func RunAllBad() {
	common.RunEndpoints(common.InitOtlpTracer, MakeBadEndpoints())
}

func MakeBadEndpoints() []common.EndPoint {
	a := common.MakeP2PEndPoint("A", runABad)
	b := common.MakeP2PEndPoint("B", runB)
	s := common.MakeP2PEndPoint("S", runS)
	return []common.EndPoint{a, b, s}
}
