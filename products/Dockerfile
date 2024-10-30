# Dockerfile для products сервиса
FROM golang:1.23.1 AS builder
WORKDIR /products

COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn
RUN go mod download

COPY . .
RUN apt-get update && apt-get install -y netcat-openbsd
COPY wait-for-rabbitmq.sh /products/wait-for-rabbitmq.sh
COPY run-migrations.sh /products/run-migrations.sh
RUN chmod +x /products/wait-for-rabbitmq.sh
RUN chmod +x /products/run-migrations.sh
RUN CGO_ENABLED=0 GOOS=linux go build -o products ./cmd/main.go
COPY migrations /products/migrations

FROM alpine
WORKDIR /products
COPY --from=builder /products/products .
COPY --from=builder /products/wait-for-rabbitmq.sh ./wait-for-rabbitmq.sh
COPY --from=builder /products/migrations ./migrations
CMD ["sh", "-c", "/products/wait-for-rabbitmq.sh && /products/run-migrations.sh && ./products"]
 