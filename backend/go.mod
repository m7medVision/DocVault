module github.com/docvault/backend

go 1.26.1

require (
	github.com/casbin/casbin/v3 v3.8.1
	github.com/casbin/gorm-adapter/v3 v3.41.0
	github.com/coder/websocket v1.8.14
	github.com/getsentry/sentry-go v0.27.0
	github.com/go-redis/redis_rate/v10 v10.0.1
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.9.1
	github.com/minio/minio-go/v7 v7.0.99
	github.com/pressly/goose/v3 v3.27.0
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/redis/go-redis/v9 v9.18.0
	github.com/stretchr/testify v1.11.1
	golang.org/x/crypto v0.51.0
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.9.1 // indirect
	github.com/casbin/govaluate v1.10.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/glebarez/go-sqlite v1.22.0 // indirect
	github.com/glebarez/sqlite v1.11.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/joho/godotenv v1.5.1
	github.com/klauspost/compress v1.18.4 // indirect
	github.com/klauspost/cpuid/v2 v2.2.11 // indirect
	github.com/klauspost/crc32 v1.3.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/microsoft/go-mssqldb v1.9.6 // indirect
	github.com/minio/crc64nvme v1.1.1 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/tinylib/msgp v1.6.1 // indirect
	go.opentelemetry.io/proto/otlp v1.1.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/exp v0.0.0-20260218203240-3dfff04db8fa // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260217215200-42d3e9bedb6d // indirect
	google.golang.org/grpc v1.79.1 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.6.0 // indirect
	gorm.io/driver/sqlserver v1.6.3 // indirect
	gorm.io/plugin/dbresolver v1.6.2 // indirect
	modernc.org/libc v1.68.0 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	modernc.org/sqlite v1.46.1 // indirect
)

require (
	go.opentelemetry.io/otel v1.40.0 // OpenTelemetry core
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.24.0 // OTLP gRPC metric exporter
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.24.0 // indirect; OTLP trace exporter
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.24.0 // OTLP gRPC trace exporter
	go.opentelemetry.io/otel/metric v1.40.0 // OpenTelemetry metrics
	go.opentelemetry.io/otel/sdk v1.39.0 // OpenTelemetry SDK
	go.opentelemetry.io/otel/sdk/metric v1.39.0 // OpenTelemetry metrics SDK
	go.opentelemetry.io/otel/trace v1.40.0 // OpenTelemetry tracing
	golang.org/x/net v0.55.0 // indirect; HTTP client/server utilities
	golang.org/x/text v0.37.0 // indirect
)
