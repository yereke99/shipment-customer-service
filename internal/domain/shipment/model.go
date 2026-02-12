package shipment

import "time"

type Shipment struct {
	ID         string
	Route      string
	Price      float64
	Status     string
	CustomerID string
	CreatedAt  time.Time
}

type CreateShipmentInput struct {
	Route       string
	Price       float64
	CustomerIDN string
}

type CreateShipmentRequest struct {
	Route    string                 `json:"route"`
	Price    float64                `json:"price"`
	Customer CreateShipmentCustomer `json:"customer"`
}

type CreateShipmentCustomer struct {
	IDN string `json:"idn"`
}

type CreateShipmentResponse struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	CustomerID string `json:"customerId"`
}

type GetShipmentResponse struct {
	ID         string  `json:"id"`
	Route      string  `json:"route"`
	Price      float64 `json:"price"`
	Status     string  `json:"status"`
	CustomerID string  `json:"customerId"`
	CreatedAt  string  `json:"created_at"`
}
