#!/bin/sh

# Устанавливаем переменные окружения для подключения к базе данных
DB_DSN="postgres://nastya:pgspgs@localhost:5434/products?sslmode=disable"
MIGRATE_PATH="./products/migrations"

# Ждем, пока база данных станет доступной
echo "Waiting for database to be available..."
until nc -z -v -w30 localhost 5434; do
  echo "Waiting for database connection..."
  sleep 1
done

echo "Database is up! Running migrations..."

# Принудительная установка миграций, если состояние базы данных грязное
migrate -path $MIGRATE_PATH -database $DB_DSN force 1

# Выполняем миграции
migrate -path $MIGRATE_PATH -database $DB_DSN up

echo "Migrations completed successfully!"
