package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lahnasti/go-market/auth/internal/config"
	"github.com/lahnasti/go-market/auth/internal/logger"
	"github.com/lahnasti/go-market/auth/internal/repository"
	"github.com/lahnasti/go-market/auth/internal/server"
	"github.com/lahnasti/go-market/auth/internal/server/routes"
	"github.com/lahnasti/go-market/lib/rabbitmq"
	"golang.org/x/sync/errgroup"
)

// @title Go-market
// @version 1.0
// @description This is a sample server for market.
// @host localhost:8080
// @BasePath /
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		fmt.Println("Received shutdown signal")
		<-c
		cancel()
	}()
	fmt.Println("Server starting")
	cfg := config.ReadConfig()
	zlog := logger.SetupLogger(cfg.DebugFlag)

	rabbitURL := os.Getenv("RABBITMQ_URL")
	rabbit, err := rabbitmq.InitRabbit(rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer rabbit.CloseRabbit()
	log.Println("RabbitMQ initialized, starting listener for user_check_queue")

	err = repository.EnsureAuthDatabaseExists(cfg.DBAddr)
	if err != nil {
		fmt.Println("Failed to ensure auth database exists:", err)
		return
	}

	pool, err := initDB(ctx, cfg.DBAddr)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Connection DB failed")
	}
	defer pool.Close()

	err = repository.Migrations(cfg.DBAddr, cfg.MPath, zlog)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Init migrations failed")
	}

	dbStorage, err := repository.NewDB(pool)
	if err != nil {
		panic(err)
	}
	defer dbStorage.Close()

	group, gCtx := errgroup.WithContext(ctx)
	srv := server.NewServer(gCtx, dbStorage, zlog, rabbit)
	
	server.StartListener(srv)

	group.Go(func() error {
		r := routes.SetupAuthRoutes(srv)
		zlog.Info().Msg("Server was started")

		if err := r.Run(cfg.Addr); err != nil {
			return err
		}
		return nil
	})

	group.Go(func() error {
		err := <-srv.ErrorChan
		return err
	})

	group.Go(func() error {
		<-gCtx.Done()
		return gCtx.Err()
	})

	if err := group.Wait(); err != nil {
		zlog.Fatal().Err(err).Msg("Error during server shutdown")
	} else {
		zlog.Info().Msg("Server excited gracefully")
	}
}

func initDB(ctx context.Context, addr string) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error
	for i := 0; i < 7; i++ {
		time.Sleep(2 * time.Second)
		pool, err = pgxpool.New(ctx, addr)
		if err == nil {
			return pool, nil
		}
	}

	return nil, fmt.Errorf("database initialization error: %w", err)
}