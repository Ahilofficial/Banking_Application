# Multi-stage Dockerfile for Banking Microservices
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_PATH
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /service ${SERVICE_PATH}

FROM alpine:3.19
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /service /app/service

EXPOSE 8000 8001 8002 8003 8004 8005

CMD ["/app/service"]
