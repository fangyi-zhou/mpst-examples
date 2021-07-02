package common

import (
	"context"
	"log"
	"time"

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
)

// https://github.com/open-telemetry/opentelemetry-go/blob/master/example/namedtracer/main.go
func InitStdoutTracer() func() {
	var err error
	exp, err := stdout.NewExporter(stdout.WithPrettyPrint())
	if err != nil {
		log.Panicf("failed to initialize stdout exporter %v\n", err)
	}
	bsp := trace.NewBatchSpanProcessor(exp)
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tp)
	return func() {}
}

const endpointUri = "http://localhost:14268/api/traces"

// https://github.com/open-telemetry/opentelemetry-go/blob/master/example/jaeger/main.go
func InitJaegerTracer() func() {
	// Create and install Jaeger export pipeline.
	exp, err := jaeger.NewRawExporter(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpointUri)),
	)
	if err != nil {
		log.Fatal(err)
	}
	tp := trace.NewTracerProvider(
		// Always be sure to batch in production.
		trace.WithBatcher(exp),
		// Record information about this application in an Resource.
		trace.WithResource(resource.NewWithAttributes(
			semconv.ServiceNameKey.String("TwoBuyer"),
		)),
	)
	otel.SetTracerProvider(tp)
	return func() {}
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
		log.Fatalf("failed to create exporter %e", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("TwoBuyer"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource %e", err)
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
	if err := cont.Start(context.Background()); err != nil {
		log.Panicf("failed to start controller %e", err)
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
