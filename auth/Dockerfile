# Dockerfile для auth сервиса
FROM golang:1.23.1 AS builder
WORKDIR /auth

COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn
RUN go mod download

COPY . .
RUN apt-get update && apt-get install -y netcat-openbsd
COPY wait-for-rabbitmq.sh /auth/wait-for-rabbitmq.sh
RUN chmod +x /auth/wait-for-rabbitmq.sh
# Сборка статически слинкованного бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -o auth ./cmd/main.go
COPY migrations /auth/migrations

FROM alpine
WORKDIR /auth
COPY --from=builder /auth/auth .
COPY --from=builder /auth/wait-for-rabbitmq.sh ./wait-for-rabbitmq.sh
COPY --from=builder /auth/migrations ./migrations
CMD ["sh", "-c", "/auth/wait-for-rabbitmq.sh && ./auth"]

