#!/bin/sh

# URL RabbitMQ
RABBITMQ_HOST="rabbitmq"
RABBITMQ_PORT=5672
TIMEOUT=300 # 5 минут

# Проверяем доступность RabbitMQ
echo "Ожидание RabbitMQ..."
for i in $(seq $TIMEOUT); do
  if nc -z $RABBITMQ_HOST $RABBITMQ_PORT; then
    echo "RabbitMQ доступен!"
    exit 0
  fi
  sleep 1
done

echo "RabbitMQ не доступен. Прерывание."
exit 1