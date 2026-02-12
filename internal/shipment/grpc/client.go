package grpc

import (
	"context"

	customerpb "shipment-customer-service/api/proto"

	"google.golang.org/grpc"
)

type CustomerClient interface {
	UpsertCustomer(ctx context.Context, idn string) (*customerpb.CustomerResponse, error)
}

type GRPCClient struct {
	client customerpb.CustomerServiceClient
}

func (c *GRPCClient) UpsertCustomer(ctx context.Context, idn string) (*customerpb.CustomerResponse, error) {
	return c.client.UpsertCustomer(ctx, &customerpb.UpsertCustomerRequest{Idn: idn})
}

func NewCustomerClientService(conn *grpc.ClientConn) *GRPCClient {
	return &GRPCClient{client: customerpb.NewCustomerServiceClient(conn)}
}
