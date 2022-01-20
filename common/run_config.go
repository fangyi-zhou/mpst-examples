package common

import "sync"

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
