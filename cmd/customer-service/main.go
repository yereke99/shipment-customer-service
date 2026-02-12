package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"

	customerpb "shipment-customer-service/api/proto"
	customergrpc "shipment-customer-service/internal/customer/grpc"
	customerrepo "shipment-customer-service/internal/customer/repo"
	customerservice "shipment-customer-service/internal/customer/service"
	"shipment-customer-service/internal/platform/telemetry"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	provider, err := telemetry.InitProvider(ctx, "customer-service", env("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317"))
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

	repo := customerrepo.NewPostgresRepo(db, otel.Tracer("customer-db"))
	if err := waitForDB(ctx, db, 30, time.Second); err != nil {
		logger.Error("db_ping_failed", slog.String("error", err.Error()))
		return
	}

	service := customerservice.New(repo)
	server := customergrpc.NewServer(service, logger)

	lis, err := net.Listen("tcp", ":"+env("GRPC_PORT", "9090"))
	if err != nil {
		logger.Error("listen_failed", slog.String("error", err.Error()))
		return
	}
	defer lis.Close()

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	customerpb.RegisterCustomerServiceServer(grpcServer, server)

	errCh := make(chan error, 1)
	go func() {
		errCh <- grpcServer.Serve(lis)
	}()

	logger.Info("customer_service_started", slog.String("addr", lis.Addr().String()))

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Error("grpc_serve_failed", slog.String("error", err.Error()))
		}
	}

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		grpcServer.Stop()
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
