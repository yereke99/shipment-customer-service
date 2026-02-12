COMPOSE_FILE := build/docker-compose.yml
COMPOSE := docker compose -f $(COMPOSE_FILE)

.PHONY: run up down stop restart logs ps build rebuild pull config test test-e2e test-all test-cases fmt proto clean

run:
	$(COMPOSE) up -d --build

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

stop:
	$(COMPOSE) stop

restart:
	$(COMPOSE) down
	$(COMPOSE) up -d --build

logs:
	$(COMPOSE) logs -f --tail=200

ps:
	$(COMPOSE) ps

build:
	$(COMPOSE) build

rebuild:
	$(COMPOSE) build --no-cache

pull:
	$(COMPOSE) pull

config:
	$(COMPOSE) config

test:
	$(MAKE) test-e2e

test-e2e:
	bash tests/e2e/task_check.sh

test-all:
	$(MAKE) test-e2e

test-cases:
	@echo "e2e"
	@echo "server start and curl post"
	@echo "grpc port 9090 works inside docker net"
	@echo "grpc port 9090 is not exposed outside"
	@echo "server start and curl get"

fmt:
	gofmt -w $$(rg --files -g '*.go')

proto:
	PATH="$$(go env GOPATH)/bin:$$PATH" protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/proto/customer.proto

clean:
	$(COMPOSE) down -v --remove-orphans
