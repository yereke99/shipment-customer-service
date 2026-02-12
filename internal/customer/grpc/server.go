package grpc

import (
	"context"
	"errors"
	"log/slog"
	"time"

	customerpb "shipment-customer-service/api/proto"
	"shipment-customer-service/internal/customer/service"
	"shipment-customer-service/internal/platform/telemetry"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	customerpb.UnimplementedCustomerServiceServer
	service *service.Service
	logger  *slog.Logger
}

func NewServer(service *service.Service, logger *slog.Logger) *Server {
	return &Server{service: service, logger: logger}
}

func (s *Server) UpsertCustomer(ctx context.Context, req *customerpb.UpsertCustomerRequest) (*customerpb.CustomerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	customer, err := s.service.UpsertCustomer(ctx, req.GetIdn())
	if err != nil {
		return nil, mapError(err)
	}

	s.logger.Info(
		"upsert_customer",
		slog.String("idn", req.GetIdn()),
		slog.String("trace_id", telemetry.TraceID(ctx)),
	)

	return &customerpb.CustomerResponse{
		Id:        customer.ID,
		Idn:       customer.IDN,
		CreatedAt: customer.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func (s *Server) GetCustomer(ctx context.Context, req *customerpb.GetCustomerRequest) (*customerpb.CustomerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	customer, err := s.service.GetCustomer(ctx, req.GetIdn())
	if err != nil {
		return nil, mapError(err)
	}

	s.logger.Info(
		"get_customer",
		slog.String("idn", req.GetIdn()),
		slog.String("trace_id", telemetry.TraceID(ctx)),
	)

	return &customerpb.CustomerResponse{
		Id:        customer.ID,
		Idn:       customer.IDN,
		CreatedAt: customer.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidIDN):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
