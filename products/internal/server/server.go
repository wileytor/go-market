package server

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/lahnasti/go-market/lib/rabbitmq"
	"github.com/lahnasti/go-market/products/internal/repository"
	"github.com/rs/zerolog"
)

type Server struct {
	Db         repository.Repository
	ErrorChan  chan error
	deleteChan chan int
	Valid      *validator.Validate
	log        zerolog.Logger
	Rabbit     *rabbitmq.RabbitMQ
}

func NewServer(ctx context.Context, db repository.Repository, zlog *zerolog.Logger, rabbitClient *rabbitmq.RabbitMQ) *Server {
	validate := validator.New()
	dChan := make(chan int, 5)
	errChan := make(chan error)
	srv := &Server{
		Db:         db,
		deleteChan: dChan,
		ErrorChan:  errChan,
		log:        *zlog,
		Valid:      validate,
		Rabbit:     rabbitClient,
	}
	go srv.deleter(ctx)
	return srv
}

func (s *Server) Close() {
	if s.Rabbit != nil {
		s.Rabbit.CloseRabbit()
	}
}