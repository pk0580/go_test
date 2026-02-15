# Build stage
FROM golang:1.24-alpine AS builder

# RUN apk add --no-cache protobuf

# RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
#    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

COPY service-go/ ./
RUN go mod tidy
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Run stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 50051

CMD ["./main"]
