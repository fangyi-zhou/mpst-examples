package common

import (
	"fmt"
	"sync"
)

type TracerInitFunc = func(string) func()

func RunEndpoints(tracerInitFunc TracerInitFunc, serviceName string, endpoints []EndPoint) {
	for i := 0; i < len(endpoints); i++ {
		for j := i + 1; j < len(endpoints); j++ {
			endpoints[i].Connect(endpoints[j])
		}
	}
	shutdown := tracerInitFunc(serviceName)
	defer shutdown()
	var wg sync.WaitGroup
	wg.Add(len(endpoints))
	for _, ep := range endpoints {
		go ep.Run(&wg)
	}
	wg.Wait()
}

func RunEndpointsMulti(tracerInitFunc TracerInitFunc, serviceName string, endpoints []EndPoint, iterations int) {
	for i := 0; i < len(endpoints); i++ {
		for j := i + 1; j < len(endpoints); j++ {
			endpoints[i].Connect(endpoints[j])
		}
	}
	shutdown := tracerInitFunc(serviceName)
	defer shutdown()
	for i := 0; i < iterations; i++ {
		var wg sync.WaitGroup
		fmt.Println("Iteration", i)
		wg.Add(len(endpoints))
		for _, ep := range endpoints {
			go ep.Run(&wg)
		}
		wg.Wait()
		for _, ep := range endpoints {
			ep.Clear()
		}
	}
}
