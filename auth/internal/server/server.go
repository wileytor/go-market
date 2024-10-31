package server

import (
	"context"
	"github.com/wileytor/go-market/common/rabbitmq"

	"github.com/go-playground/validator"
	"github.com/rs/zerolog"
	"github.com/wileytor/go-market/auth/internal/repository"
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
