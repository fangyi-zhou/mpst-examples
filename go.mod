module github.com/fangyi-zhou/mpst-examples

go 1.15

require (
	github.com/fangyi-zhou/mpst-tracing v0.0.0-20210623022711-e5a037806520
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/exporters/otlp v0.20.0
	go.opentelemetry.io/otel/exporters/stdout v0.20.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.20.0
	go.opentelemetry.io/otel/metric v0.20.0
	go.opentelemetry.io/otel/sdk v0.20.0
	go.opentelemetry.io/otel/sdk/metric v0.20.0
	go.opentelemetry.io/otel/trace v0.20.0
	google.golang.org/grpc v1.38.0
)

replace github.com/fangyi-zhou/mpst-tracing v0.0.0 => ../mpst-tracing
