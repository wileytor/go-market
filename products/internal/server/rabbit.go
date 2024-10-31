package server

import "github.com/streadway/amqp"

func (s *Server) EnsureConnection() error {
	if s.Rabbit.Connection == nil || s.Rabbit.Connection.IsClosed() {
		var err error
		s.Rabbit.Connection, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err != nil {
			return err
		}
		s.Rabbit.Channel, err = s.Rabbit.Connection.Channel()
		if err != nil {
			return err
		}
	}
	return nil
}
