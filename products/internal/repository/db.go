package repository

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func EnsureMarketDatabaseExists(connString string) error {
	const dbName = "products"
	connStringForCreation := "postgres://nastya:pgspgs@db:5432/postgres?sslmode=disable"

	conn, err := pgx.Connect(context.Background(), connStringForCreation)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer conn.Close(context.Background())

	err = EnsureDatabaseExists(conn, dbName)
	if err != nil {
		return err
	}

	// После успешного создания базы данных, возвращаем строку для подключения к auth
	os.Setenv("DB_ADDR", "postgres://nastya:pgspgs@db:5432/products?sslmode=disable")
	return EnsureDatabaseExists(conn, dbName)
}

// Общая функция для проверки и создания базы данных
func EnsureDatabaseExists(conn *pgx.Conn, dbName string) error {
	// Проверяем, существует ли база данных
	var exists bool
	err := conn.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Если база данных не существует, создаём её
	if !exists {
		_, err = conn.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Printf("Database %s created successfully", dbName)
	}

	return nil
}
