.PHONY: run-api build-api run-worker build-worker clean swagger swagger-install docker-up docker-down docker-logs tidy lint
 
# ── Variables ──
APP_NAME   := kvault
API_CMD_PATH   := ./cmd/api
API_BINARY     := bin/$(APP_NAME)_api
WORKER_CMD_PATH   := ./cmd/worker
WORKER_BINARY     := bin/$(APP_NAME)_worker
SWAGGER_OUT := ./docs
 
# ── Development ──
## run-api: Run api dev server 
run-api:
	go run $(API_CMD_PATH)/main.go

## build-api: Build api dev server binary
build-api:
	go build -o $(API_BINARY) $(API_CMD_PATH)/main.go

## run-worker: Run worker dev server
run-worker:
	go run $(WORKER_CMD_PATH)/main.go

## build-worker: Build worker dev server binary
build-worker:
	go build -o $(WORKER_BINARY) $(WORKER_CMD_PATH)/main.go

## clean: Clean bin directory
clean:
	rm -rf bin/
 
# ── go.mod ──
## tidy: Tidy go.mod
tidy:
	go mod tidy
	go mod verify
 
# ── Testing ──
# test:
# 	go test ./... -v

# test-cover:
# 	go test ./... -coverprofile=coverage.out
# 	go tool cover -html=coverage.out
 
# ── Swagger ──
## swagger: Generate swagger docs
swagger:
	swag init \
		--dir ./cmd/api,./internal/httpx,./internal/handlers \
		--output $(SWAGGER_OUT)

## swagger-install: Install swagger dependencies
swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest
 
# ── Docker ──
## docker-up: Start docker containers
docker-up:
	docker-compose up -d

## docker-down: Stop docker containers
docker-down:
	docker-compose down

## docker-logs: Follow docker logs
docker-logs:
	docker-compose logs -f

## docker-rebuild: Rebuild docker containers
docker-rebuild:
	docker-compose up -d --build
 
# ── Helpers ──
## help: Show help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'