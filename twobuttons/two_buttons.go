package twobuttons

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/fangyi-zhou/mpst-examples/common"
	"go.opentelemetry.io/otel/trace"
)

func runA(self common.EndPoint, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	var span trace.Span
	ctx, span = self.Tracer().Start(ctx, "TwoButtons Endpoint A")
	defer span.End()
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	fmt.Println("A: Sending STOP")
	self.Send(ctx, "M", common.Message{Label: "STOP_from_A", Value: nil})
	var actualStop, _ = self.RecvSync(ctx, "M")
	fmt.Printf("A: Got stop signal from M, with label %s\n", actualStop.Label)
}

func runB(self common.EndPoint, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	var span trace.Span
	ctx, span = self.Tracer().Start(ctx, "TwoButtons Endpoint B")
	defer span.End()
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	fmt.Println("B: Sending STOP")
	self.Send(ctx, "M", common.Message{Label: "STOP_from_B", Value: nil})
	var actualStop, _ = self.RecvSync(ctx, "M")
	fmt.Printf("B: Got stop signal from M, with label %s\n", actualStop.Label)
}

func runM(self common.EndPoint, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	var span trace.Span
	ctx, span = self.Tracer().Start(ctx, "TwoButtons Endpoint M")
	defer span.End()
	var gotSignal = false
	var coinFlip = rand.Int()%2 == 0
	for !gotSignal {
		if coinFlip {
			// Check A first
			fmt.Println("M: Checking stop signal at A")
			msg, _ := self.RecvAsync(ctx, "A")
			if msg != nil {
				fmt.Println("M: Got stop signal from A, notifying STOPPED_A")
				gotSignal = true
				self.Send(ctx, "A", common.Message{Label: "STOPPED_A", Value: nil})
				self.Send(ctx, "B", common.Message{Label: "STOPPED_A", Value: nil})
			}
		} else {
			// Check B first
			fmt.Println("M: Checking stop signal at B")
			msg, _ := self.RecvAsync(ctx, "B")
			if msg != nil {
				fmt.Println("M: Got stop signal from B, notifying STOPPED_B")
				gotSignal = true
				self.Send(ctx, "A", common.Message{Label: "STOPPED_B", Value: nil})
				self.Send(ctx, "B", common.Message{Label: "STOPPED_B", Value: nil})
			}
		}
		coinFlip = !coinFlip
		time.Sleep(time.Duration(10) * time.Millisecond)
	}
}

func runMMailbox(self common.EndPoint, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	var span trace.Span
	ctx, span = self.Tracer().Start(ctx, "TwoButtons Endpoint M")
	defer span.End()
	fmt.Println("M: Waiting for the first stop signal")
	msg, _ := self.RecvSync(ctx, "")
	if msg.Label == "STOP_from_A" {
		fmt.Println("M: Got stop signal from A, notifying STOPPED_A")
		self.Send(ctx, "A", common.Message{Label: "STOPPED_A", Value: nil})
		self.Send(ctx, "B", common.Message{Label: "STOPPED_A", Value: nil})
	} else if msg.Label == "STOP_from_B" {
		fmt.Println("M: Got stop signal from B, notifying STOPPED_B")
		self.Send(ctx, "A", common.Message{Label: "STOPPED_B", Value: nil})
		self.Send(ctx, "B", common.Message{Label: "STOPPED_B", Value: nil})
	} else {
		panic("Unexpected label")
	}
}

func MakeEndpoints() []common.EndPoint {
	a := common.MakeP2PEndPoint("A", runA)
	b := common.MakeP2PEndPoint("B", runB)
	m := common.MakeP2PEndPoint("M", runM)
	return []common.EndPoint{a, b, m}
}

func MakeMailboxEndpoints() []common.EndPoint {
	a := common.MakeMailBoxEndPoint("A", runA)
	b := common.MakeMailBoxEndPoint("B", runB)
	m := common.MakeMailBoxEndPoint("M", runMMailbox)
	return []common.EndPoint{a, b, m}
}

func RunAll() {
	rand.Seed(time.Now().UnixMilli())
	common.RunEndpoints(common.InitOtlpTracer, "TwoButtons", MakeEndpoints())
}

func RunAllMailbox() {
	rand.Seed(time.Now().UnixMilli())
	common.RunEndpoints(common.InitOtlpTracer, "TwoButtons", MakeMailboxEndpoints())
}
