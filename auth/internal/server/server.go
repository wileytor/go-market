package server

import (
	"context"

	"github.com/go-playground/validator"
	"github.com/lahnasti/go-market/auth/internal/repository"
	"github.com/lahnasti/go-market/lib/rabbitmq"
	"github.com/rs/zerolog"
)

type Server struct {
	Db        repository.UserRepository
	ErrorChan chan error
	Valid     *validator.Validate
	log       zerolog.Logger
	Rabbit    *rabbitmq.RabbitMQ
}

func NewServer(ctx context.Context, db repository.UserRepository, zlog *zerolog.Logger, rabbitClient *rabbitmq.RabbitMQ) *Server {
	validate := validator.New()
	errChan := make(chan error)
	srv := &Server{
		Db:        db,
		ErrorChan: errChan,
		log:       *zlog,
		Valid:     validate,
		Rabbit:    rabbitClient,
	}
	return srv
}

func (s *Server) Close() {
	if s.Rabbit != nil {
		s.Rabbit.CloseRabbit()
	}
}
