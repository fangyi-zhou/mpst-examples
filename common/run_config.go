package common

import "sync"

type TracerInitFunc = func() func()

func RunEndpoints(tracerInitFunc TracerInitFunc, endpoints []EndPoint) {
	for i := 0; i < len(endpoints); i++ {
		for j := i + 1; j < len(endpoints); j++ {
			connectEndpoints(&endpoints[i], &endpoints[j])
		}
	}
	shutdown := tracerInitFunc()
	defer shutdown()
	var wg sync.WaitGroup
	wg.Add(len(endpoints))
	for _, ep := range endpoints {
		go ep.Run(&wg)
	}
	wg.Wait()
}
