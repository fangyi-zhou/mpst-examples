module github.com/fangyi-zhou/mpst-examples

go 1.17

require (
	github.com/fangyi-zhou/mpst-tracing v0.0.0
	go.opentelemetry.io/otel v1.11.1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.11.1
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.11.1
	go.opentelemetry.io/otel/sdk v1.11.1
	go.opentelemetry.io/otel/trace v1.11.1
	google.golang.org/grpc v1.51.0
)

require github.com/grpc-ecosystem/grpc-gateway/v2 v2.14.0 // indirect

require (
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.11.1
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.11.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.11.1
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	golang.org/x/net v0.2.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	google.golang.org/genproto v0.0.0-20221201204527-e3fa12d562f3 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace github.com/fangyi-zhou/mpst-tracing v0.0.0 => ../mpst-tracing
