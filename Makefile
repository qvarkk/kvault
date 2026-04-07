.PHONY: run build clean test swagger docker-up docker-down docker-logs tidy lint
 
# ── Variables ──
APP_NAME   := kvault
API_CMD_PATH   := ./cmd/api
API_BINARY     := bin/$(APP_NAME)_api
SWAGGER_OUT := ./docs
 
# ── Development ──
run-api:
	go run $(API_CMD_PATH)/main.go

build-api:
	go build -o $(API_BINARY) $(API_CMD_PATH)/main.go

clean:
	rm -rf bin/
 
# ── go.mod ──
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
swagger:
	swag init \
		--dir ./cmd/api,./internal/httpx,./internal/handlers \
		--output $(SWAGGER_OUT)

swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest
 
# ── Docker ──
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-rebuild:
	docker-compose up -d --build
 
# ── Helpers ──
help:
	@echo "Usage: make <target>"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'