package http

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	domain "shipment-customer-service/internal/domain/shipment"
	"shipment-customer-service/internal/platform/telemetry"
	"shipment-customer-service/internal/shipment/service"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	service *service.Service
	logger  *slog.Logger
}

func NewHandler(service *service.Service, logger *slog.Logger) http.Handler {
	h := &Handler{service: service, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.healthCheckHandler)
	mux.HandleFunc("POST /api/v1/shipments", h.createShipment)
	mux.HandleFunc("GET /api/v1/shipments/{id}", h.getShipment)
	return otelhttp.NewHandler(mux, "shipment-http")
}

func (h *Handler) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) createShipment(w http.ResponseWriter, r *http.Request) {
	var request domain.CreateShipmentRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, domain.ErrorResponse{Error: "invalid request body"})
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, domain.ErrorResponse{Error: "invalid request body"})
		return
	}

	shipment, err := h.service.Create(r.Context(), domain.CreateShipmentInput{
		Route:       request.Route,
		Price:       request.Price,
		CustomerIDN: request.Customer.IDN,
	})
	if err != nil {
		statusCode, message := mapCreateError(err)
		writeJSON(w, statusCode, domain.ErrorResponse{Error: message})
		return
	}

	h.logger.Info(
		"shipment_created",
		slog.String("shipment_id", shipment.ID),
		slog.String("trace_id", telemetry.TraceID(r.Context())),
	)

	writeJSON(w, http.StatusCreated, domain.CreateShipmentResponse{
		ID:         shipment.ID,
		Status:     shipment.Status,
		CustomerID: shipment.CustomerID,
	})
}

func (h *Handler) getShipment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	shipment, err := h.service.Get(r.Context(), id)
	if err != nil {
		statusCode, message := mapGetError(err)
		writeJSON(w, statusCode, domain.ErrorResponse{Error: message})
		return
	}

	h.logger.Info(
		"shipment_fetched",
		slog.String("shipment_id", shipment.ID),
		slog.String("trace_id", telemetry.TraceID(r.Context())),
	)

	writeJSON(w, http.StatusOK, domain.GetShipmentResponse{
		ID:         shipment.ID,
		Route:      shipment.Route,
		Price:      shipment.Price,
		Status:     shipment.Status,
		CustomerID: shipment.CustomerID,
		CreatedAt:  shipment.CreatedAt.UTC().Format(time.RFC3339),
	})
}

func mapCreateError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrInvalidRoute):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrInvalidPrice):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrInvalidIDN):
		return http.StatusBadRequest, err.Error()
	}

	if grpcStatus, ok := status.FromError(err); ok {
		switch grpcStatus.Code() {
		case codes.InvalidArgument:
			return http.StatusBadRequest, grpcStatus.Message()
		case codes.Unavailable:
			return http.StatusServiceUnavailable, "customer service unavailable"
		default:
			return http.StatusBadGateway, grpcStatus.Message()
		}
	}

	return http.StatusInternalServerError, "internal error"
}

func mapGetError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrInvalidShipmentID):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, err.Error()
	default:
		return http.StatusInternalServerError, "internal error"
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
