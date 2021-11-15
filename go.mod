module github.com/fangyi-zhou/mpst-examples

go 1.17

require (
	github.com/fangyi-zhou/mpst-tracing v0.0.0-20211026231637-2aef111310a3
	go.opentelemetry.io/otel v1.2.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.2.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.2.0
	go.opentelemetry.io/otel/sdk v1.2.0
	go.opentelemetry.io/otel/trace v1.2.0
	google.golang.org/grpc v1.42.0
)

require github.com/cenkalti/backoff/v4 v4.1.2 // indirect

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.2.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.2.0
	go.opentelemetry.io/proto/otlp v0.11.0 // indirect
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	golang.org/x/sys v0.0.0-20211113001501-0c823b97ae02 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20211112145013-271947fe86fd // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

replace github.com/fangyi-zhou/mpst-tracing v0.0.0 => ../mpst-tracing
