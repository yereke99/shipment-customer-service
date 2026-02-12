package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"shipment-customer-service/internal/platform/telemetry"
	shipmentgrpc "shipment-customer-service/internal/shipment/grpc"
	httptransport "shipment-customer-service/internal/shipment/http"
	shipmentrepo "shipment-customer-service/internal/shipment/repo"
	shipmentservice "shipment-customer-service/internal/shipment/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	provider, err := telemetry.InitProvider(ctx, "shipment-service", env("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317"))
	if err != nil {
		logger.Error("otel_init_failed", slog.String("error", err.Error()))
		return
	}
	defer shutdownTracer(provider, logger)

	db, err := sql.Open("pgx", env("POSTGRES_DSN", "postgres://app:app@postgres:5432/app?sslmode=disable"))
	if err != nil {
		logger.Error("db_open_failed", slog.String("error", err.Error()))
		return
	}
	defer db.Close()

	repo := shipmentrepo.NewPostgresRepo(db, otel.Tracer("shipment-db"))
	if err := waitForDB(ctx, db, 30, time.Second); err != nil {
		logger.Error("db_ping_failed", slog.String("error", err.Error()))
		return
	}

	conn, err := grpc.DialContext(
		ctx,
		env("CUSTOMER_GRPC_ADDR", "envoy:9090"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logger.Error("grpc_dial_failed", slog.String("error", err.Error()))
		return
	}
	defer conn.Close()

	customerClient := shipmentgrpc.NewCustomerClientService(conn)
	service := shipmentservice.New(repo, customerClient)
	handler := httptransport.NewHandler(service, logger)

	httpServer := &http.Server{
		Addr:              ":" + env("HTTP_PORT", "8080"),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	logger.Info("shipment_service_started", slog.String("addr", httpServer.Addr))

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http_serve_failed", slog.String("error", err.Error()))
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http_shutdown_failed", slog.String("error", err.Error()))
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func shutdownTracer(provider *sdktrace.TracerProvider, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := provider.Shutdown(ctx); err != nil {
		logger.Error("otel_shutdown_failed", slog.String("error", err.Error()))
	}
}

func waitForDB(ctx context.Context, db *sql.DB, attempts int, delay time.Duration) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = db.PingContext(ctx)
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return err
}
