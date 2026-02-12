# Shipment Customer Service

Минимальный проект с REST, gRPC, Envoy, Postgres и OpenTelemetry.

## Запуск

```bash
make run
```

## Проверка API

```bash
curl -X POST http://localhost:8080/api/v1/shipments \
  -H "Content-Type: application/json" \
  -d '{"route":"ALMATY->ASTANA","price":120000,"customer":{"idn":"990101123456"}}'
```

```bash
curl http://localhost:8080/api/v1/shipments/<id>
```

## Трейсы

Jaeger: `http://localhost:16686`

## Тесты

```bash
make test-cases
make test-e2e
make test-all
```

## Остановка

```bash
make down
```
