package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	customerpb "shipment-customer-service/api/proto"
	domain "shipment-customer-service/internal/domain/shipment"
)

type mockRepo struct {
	createFn func(ctx context.Context, route string, price float64, customerID string) (domain.Shipment, error)
	getFn    func(ctx context.Context, id string) (domain.Shipment, error)
}

func (m *mockRepo) CreateShipment(ctx context.Context, route string, price float64, customerID string) (domain.Shipment, error) {
	if m.createFn == nil {
		return domain.Shipment{}, nil
	}
	return m.createFn(ctx, route, price, customerID)
}

func (m *mockRepo) GetShipment(ctx context.Context, id string) (domain.Shipment, error) {
	if m.getFn == nil {
		return domain.Shipment{}, nil
	}
	return m.getFn(ctx, id)
}

type mockCustomerClient struct {
	upsertFn func(ctx context.Context, idn string) (*customerpb.CustomerResponse, error)
}

func (m *mockCustomerClient) UpsertCustomer(ctx context.Context, idn string) (*customerpb.CustomerResponse, error) {
	if m.upsertFn == nil {
		return nil, nil
	}
	return m.upsertFn(ctx, idn)
}

func TestCreateValidation(t *testing.T) {
	tests := []struct {
		name  string
		input domain.CreateShipmentInput
		err   error
	}{
		{name: "invalid route", input: domain.CreateShipmentInput{Route: "   ", Price: 1, CustomerIDN: "990101123456"}, err: domain.ErrInvalidRoute},
		{name: "invalid price", input: domain.CreateShipmentInput{Route: "A-B", Price: 0, CustomerIDN: "990101123456"}, err: domain.ErrInvalidPrice},
		{name: "invalid idn", input: domain.CreateShipmentInput{Route: "A-B", Price: 1, CustomerIDN: "123"}, err: domain.ErrInvalidIDN},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc := New(&mockRepo{}, &mockCustomerClient{})
			_, err := svc.Create(context.Background(), tc.input)
			if !errors.Is(err, tc.err) {
				t.Fatalf("Create() error = %v, want %v", err, tc.err)
			}
		})
	}
}

func TestCreateSuccess(t *testing.T) {
	var gotRoute string
	var gotPrice float64
	var gotCustomerID string
	var gotIDN string

	svc := New(
		&mockRepo{
			createFn: func(ctx context.Context, route string, price float64, customerID string) (domain.Shipment, error) {
				gotRoute = route
				gotPrice = price
				gotCustomerID = customerID
				return domain.Shipment{ID: "s1", Status: "CREATED", CustomerID: customerID}, nil
			},
		},
		&mockCustomerClient{
			upsertFn: func(ctx context.Context, idn string) (*customerpb.CustomerResponse, error) {
				gotIDN = idn
				return &customerpb.CustomerResponse{Id: "c1"}, nil
			},
		},
	)

	got, err := svc.Create(context.Background(), domain.CreateShipmentInput{
		Route:       "  ALMATY->ASTANA  ",
		Price:       120000,
		CustomerIDN: " 990101123456 ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.ID != "s1" {
		t.Fatalf("Create() id = %s, want s1", got.ID)
	}
	if gotRoute != "ALMATY->ASTANA" {
		t.Fatalf("Create() route passed to repo = %q", gotRoute)
	}
	if gotPrice != 120000 {
		t.Fatalf("Create() price passed to repo = %v", gotPrice)
	}
	if gotIDN != "990101123456" {
		t.Fatalf("Create() idn passed to customer client = %q", gotIDN)
	}
	if gotCustomerID != "c1" {
		t.Fatalf("Create() customerID passed to repo = %q", gotCustomerID)
	}
}

func TestCreateCustomerError(t *testing.T) {
	wantErr := errors.New("grpc failed")
	svc := New(
		&mockRepo{},
		&mockCustomerClient{upsertFn: func(ctx context.Context, idn string) (*customerpb.CustomerResponse, error) {
			return nil, wantErr
		}},
	)

	_, err := svc.Create(context.Background(), domain.CreateShipmentInput{Route: "A-B", Price: 1, CustomerIDN: "990101123456"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Create() error = %v, want %v", err, wantErr)
	}
}

func TestGet(t *testing.T) {
	now := time.Now().UTC()
	want := domain.Shipment{ID: "11111111-1111-1111-1111-111111111111", Route: "A-B", Price: 1, Status: "CREATED", CustomerID: "c1", CreatedAt: now}

	t.Run("invalid shipment id", func(t *testing.T) {
		svc := New(&mockRepo{}, &mockCustomerClient{})
		_, err := svc.Get(context.Background(), "bad-id")
		if !errors.Is(err, domain.ErrInvalidShipmentID) {
			t.Fatalf("Get() error = %v, want %v", err, domain.ErrInvalidShipmentID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := New(&mockRepo{getFn: func(ctx context.Context, id string) (domain.Shipment, error) {
			return domain.Shipment{}, sql.ErrNoRows
		}}, &mockCustomerClient{})
		_, err := svc.Get(context.Background(), want.ID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("Get() error = %v, want %v", err, domain.ErrNotFound)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		wantErr := errors.New("db failed")
		svc := New(&mockRepo{getFn: func(ctx context.Context, id string) (domain.Shipment, error) {
			return domain.Shipment{}, wantErr
		}}, &mockCustomerClient{})
		_, err := svc.Get(context.Background(), want.ID)
		if !errors.Is(err, wantErr) {
			t.Fatalf("Get() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("success", func(t *testing.T) {
		svc := New(&mockRepo{getFn: func(ctx context.Context, id string) (domain.Shipment, error) {
			return want, nil
		}}, &mockCustomerClient{})
		got, err := svc.Get(context.Background(), want.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got != want {
			t.Fatalf("Get() = %+v, want %+v", got, want)
		}
	})
}
