FROM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

ENV API_PORT=6767

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o api ./cmd/api

FROM alpine:3.21 AS final

WORKDIR /app

COPY --from=builder /app/api .

EXPOSE $API_PORT

CMD ["./api"]