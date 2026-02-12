module shipment-customer-service

go 1.24.0

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.8.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.65.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.65.0
	go.opentelemetry.io/otel v1.40.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.40.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.40.0
	go.opentelemetry.io/otel/sdk v1.40.0
	go.opentelemetry.io/otel/trace v1.40.0
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/cenkalti/backoff/v5 v5.0.3
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/felixge/httpsnoop v1.0.4
	github.com/go-logr/logr v1.4.3
	github.com/go-logr/stdr v1.2.2
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.7
	github.com/jackc/pgpassfile v1.0.0
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
	github.com/jackc/puddle/v2 v2.2.2
	go.opentelemetry.io/auto/sdk v1.2.1
	go.opentelemetry.io/otel/metric v1.40.0
	go.opentelemetry.io/proto/otlp v1.9.0
	golang.org/x/net v0.49.0
	golang.org/x/sync v0.19.0
	golang.org/x/sys v0.40.0
	golang.org/x/text v0.33.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260128011058-8636f8732409
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260128011058-8636f8732409
)
