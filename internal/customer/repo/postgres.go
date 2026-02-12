package repo

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	domain "shipment-customer-service/internal/domain/customer"
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

func (r *PostgresRepo) UpsertCustomer(ctx context.Context, idn string) (domain.Customer, error) {
	ctx, span := r.tracer.Start(ctx, "customer.repo.UpsertCustomer")
	defer span.End()

	row := r.db.QueryRowContext(ctx, `
		INSERT INTO customers (id, idn)
		VALUES ($1, $2)
		ON CONFLICT (idn) DO UPDATE SET idn = EXCLUDED.idn
		RETURNING id::text, idn, created_at
	`, uuid.NewString(), idn)

	var customer domain.Customer
	if err := row.Scan(&customer.ID, &customer.IDN, &customer.CreatedAt); err != nil {
		return domain.Customer{}, err
	}

	return customer, nil
}

func (r *PostgresRepo) GetCustomerByIDN(ctx context.Context, idn string) (domain.Customer, error) {
	ctx, span := r.tracer.Start(ctx, "customer.repo.GetCustomerByIDN")
	defer span.End()

	row := r.db.QueryRowContext(ctx, `
		SELECT id::text, idn, created_at
		FROM customers
		WHERE idn = $1
	`, idn)

	var customer domain.Customer
	if err := row.Scan(&customer.ID, &customer.IDN, &customer.CreatedAt); err != nil {
		return domain.Customer{}, err
	}

	return customer, nil
}
