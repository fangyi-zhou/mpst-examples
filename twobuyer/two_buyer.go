package twobuyer

import (
	"context"
	"fmt"
	"github.com/fangyi-zhou/mpst-examples/common"
	"go.opentelemetry.io/otel/trace"
	"math/rand"
	"sync"
)

func runA(self *common.EndPoint, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	var span trace.Span
	ctx, span = self.Tracer.Start(ctx, "TwoBuyer Endpoint A")
	defer span.End()
	// Send query to S
	var query = rand.Intn(100)
	fmt.Println("A: Sending query", query)
	self.Send(ctx, "S", common.Message{Label: "query", Value: query})
	// Receive a quote
	var quote = self.Recv(ctx, "S").Value.(int)
	var otherShare = self.Recv(ctx, "B").Value.(int)
	if otherShare*2 >= quote {
		// 1 stands for ok
		self.Send(ctx, "S", common.Message{Label: "buy", Value: 1})
	} else {
		self.Send(ctx, "S", common.Message{Label: "buy", Value: 1})
	}
}

func runS(self *common.EndPoint, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	var span trace.Span
	ctx, span = self.Tracer.Start(ctx, "TwoBuyer Endpoint S")
	defer span.End()
	// Receive a query
	var query = self.Recv(ctx, "A").Value.(int)
	// Send a quote
	var quote = query * 2
	fmt.Println("S: Sending quote", quote)
	self.Send(ctx, "A", common.Message{Label: "quote", Value: quote})
	self.Send(ctx, "B", common.Message{Label: "quote", Value: quote})
	var decision = self.Recv(ctx, "A").Value.(int)
	if decision == 1 {
		fmt.Println("Succeed!")
	} else {
		fmt.Println("Failed to succeed!")
	}
}

func runB(self *common.EndPoint, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	var span trace.Span
	ctx, span = self.Tracer.Start(ctx, "TwoBuyer Endpoint B")
	defer span.End()
	// Receive a quote
	var quote = self.Recv(ctx, "S").Value.(int)
	// Propose a share
	var share = quote/2 + rand.Intn(10) - 5
	fmt.Println("B: Proposing share", share)
	self.Send(ctx, "A", common.Message{Label: "share", Value: share})
}

func MakeEndpoints() []*common.EndPoint {
	a := common.MakeEndPoint("A", runA)
	b := common.MakeEndPoint("B", runB)
	s := common.MakeEndPoint("S", runS)
	return []*common.EndPoint{a, b, s}
}

func RunAll() {
	common.RunEndpoints(common.InitOtlpTracer, MakeEndpoints())
}
