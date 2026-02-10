# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY service-go/go.mod ./
# COPY service-go/go.sum ./
RUN go mod download

COPY service-go/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Run stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

CMD ["./main"]
