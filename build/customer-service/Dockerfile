FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/customer-service ./cmd/customer-service

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/customer-service /customer-service
ENTRYPOINT ["/customer-service"]
