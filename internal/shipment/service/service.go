package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	domain "shipment-customer-service/internal/domain/shipment"
	"shipment-customer-service/internal/shipment/grpc"
)

type ShipmentRepository interface {
	CreateShipment(ctx context.Context, route string, price float64, customerID string) (domain.Shipment, error)
	GetShipment(ctx context.Context, id string) (domain.Shipment, error)
}

type Service struct {
	repo           ShipmentRepository
	customerClient grpc.CustomerClient
}

func New(repository ShipmentRepository, customerClient grpc.CustomerClient) *Service {
	return &Service{repo: repository, customerClient: customerClient}
}

func (s *Service) Create(ctx context.Context, input domain.CreateShipmentInput) (domain.Shipment, error) {
	route := strings.TrimSpace(input.Route)
	if route == "" {
		return domain.Shipment{}, domain.ErrInvalidRoute
	}
	if input.Price <= 0 {
		return domain.Shipment{}, domain.ErrInvalidPrice
	}

	idn := strings.TrimSpace(input.CustomerIDN)
	if !domain.IsValidIDN(idn) {
		return domain.Shipment{}, domain.ErrInvalidIDN
	}

	customer, err := s.customerClient.UpsertCustomer(ctx, idn)
	if err != nil {
		return domain.Shipment{}, err
	}

	return s.repo.CreateShipment(ctx, route, input.Price, customer.GetId())
}

func (s *Service) Get(ctx context.Context, id string) (domain.Shipment, error) {
	if _, err := uuid.Parse(id); err != nil {
		return domain.Shipment{}, domain.ErrInvalidShipmentID
	}

	shipment, err := s.repo.GetShipment(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Shipment{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Shipment{}, err
	}

	return shipment, nil
}
