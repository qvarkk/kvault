FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init --dir ./cmd/api,./internal/httpx,./internal/handlers/web --output ./docs

ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o api ./cmd/api && \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o worker ./cmd/worker && \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o migrate ./cmd/migrate

FROM alpine:3.21 AS final

WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/worker .
COPY --from=builder /app/migrate .

CMD ["./api"]
