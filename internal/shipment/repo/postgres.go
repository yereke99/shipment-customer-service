package repo

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	domain "shipment-customer-service/internal/domain/shipment"
)

type PostgresRepo struct {
	db     *sql.DB
	tracer trace.Tracer
}

func NewPostgresRepo(db *sql.DB, tracer trace.Tracer) *PostgresRepo {
	return &PostgresRepo{db: db, tracer: tracer}
}

func (r *PostgresRepo) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *PostgresRepo) CreateShipment(ctx context.Context, route string, price float64, customerID string) (domain.Shipment, error) {
	ctx, span := r.tracer.Start(ctx, "shipment.repo.CreateShipment")
	defer span.End()

	row := r.db.QueryRowContext(ctx, `
		INSERT INTO shipments (id, route, price, customer_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, route, price::text, status, customer_id::text, created_at
	`, uuid.NewString(), route, price, customerID)

	var shipment domain.Shipment
	var priceText string
	if err := row.Scan(&shipment.ID, &shipment.Route, &priceText, &shipment.Status, &shipment.CustomerID, &shipment.CreatedAt); err != nil {
		return domain.Shipment{}, err
	}

	parsedPrice, err := strconv.ParseFloat(priceText, 64)
	if err != nil {
		return domain.Shipment{}, err
	}
	shipment.Price = parsedPrice

	return shipment, nil
}

func (r *PostgresRepo) GetShipment(ctx context.Context, id string) (domain.Shipment, error) {
	ctx, span := r.tracer.Start(ctx, "shipment.repo.GetShipment")
	defer span.End()

	row := r.db.QueryRowContext(ctx, `
		SELECT id::text, route, price::text, status, customer_id::text, created_at
		FROM shipments
		WHERE id = $1
	`, id)

	var shipment domain.Shipment
	var priceText string
	if err := row.Scan(&shipment.ID, &shipment.Route, &priceText, &shipment.Status, &shipment.CustomerID, &shipment.CreatedAt); err != nil {
		return domain.Shipment{}, err
	}

	parsedPrice, err := strconv.ParseFloat(priceText, 64)
	if err != nil {
		return domain.Shipment{}, err
	}
	shipment.Price = parsedPrice

	return shipment, nil
}
