package twobuyer

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"google.golang.org/grpc"
	"log"
	"time"
)

// https://github.com/open-telemetry/opentelemetry-go/blob/master/example/namedtracer/main.go
func InitStdoutTracer() func() {
	var err error
	exp, err := stdout.NewExporter(stdout.WithPrettyPrint())
	if err != nil {
		log.Panicf("failed to initialize stdout exporter %v\n", err)
		return func() {}
	}
	bsp := trace.NewBatchSpanProcessor(exp)
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tp)
	return func() {}
}

// https://github.com/open-telemetry/opentelemetry-go/blob/master/example/jaeger/main.go
func InitJaegerTracer() func() {
	// Create and install Jaeger export pipeline.
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
		jaeger.WithSDKOptions(
			trace.WithSampler(trace.AlwaysSample()),
			trace.WithResource(resource.NewWithAttributes(
				semconv.ServiceNameKey.String("TwoBuyer"),
			)),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	return flush
}

// https://github.com/open-telemetry/opentelemetry-go/blob/master/example/otel-collector/main.go
func InitOtlpTracer() func() {
	ctx := context.Background()

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns
	driver := otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint("localhost:30080"),
		otlpgrpc.WithDialOption(grpc.WithBlock()), // useful for testing
	)
	exp, err := otlp.NewExporter(ctx, driver)
	if err != nil {
		log.Panicf("Failed to create exporter, %v\n", err)
		return nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("TwoBuyer"),
		),
	)
	if err != nil {
		log.Panicf("Failed to create resource, %v\n", err)
	}

	bsp := trace.NewBatchSpanProcessor(exp)
	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
	)

	cont := controller.New(
		processor.New(
			simple.NewWithExactDistribution(),
			exp,
		),
		controller.WithExporter(exp),
		controller.WithCollectPeriod(2*time.Second),
	)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tracerProvider)
	global.SetMeterProvider(cont.MeterProvider())

	err = cont.Start(context.Background())
	if err != nil {
		log.Panicf("Failed to start controller, %v\n", err)
	}

	return func() {
		err = cont.Stop(context.Background())
		if err != nil {
			log.Panicf("Failed to stop controller, %v\n", err)
		}

		err = tracerProvider.Shutdown(ctx)
		if err != nil {
			log.Panicf("Failed to shutdown TraceProvider, %v\n", err)
		}
	}
}
