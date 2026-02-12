package service

import (
	"context"
	"database/sql"
	"errors"
	"regexp"

	"shipment-customer-service/internal/customer/repo"
	domain "shipment-customer-service/internal/domain/customer"
)

var idnPattern = regexp.MustCompile(`^\d{12}$`)

var (
	ErrInvalidIDN = errors.New("invalid idn")
	ErrNotFound   = errors.New("customer not found")
)

type Service struct {
	repo *repo.PostgresRepo
}

func New(repository *repo.PostgresRepo) *Service {
	return &Service{repo: repository}
}

func (s *Service) UpsertCustomer(ctx context.Context, idn string) (domain.Customer, error) {
	if !idnPattern.MatchString(idn) {
		return domain.Customer{}, ErrInvalidIDN
	}

	return s.repo.UpsertCustomer(ctx, idn)
}

func (s *Service) GetCustomer(ctx context.Context, idn string) (domain.Customer, error) {
	if !idnPattern.MatchString(idn) {
		return domain.Customer{}, ErrInvalidIDN
	}

	customer, err := s.repo.GetCustomerByIDN(ctx, idn)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Customer{}, ErrNotFound
	}
	if err != nil {
		return domain.Customer{}, err
	}

	return customer, nil
}
