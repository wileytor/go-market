package rabbitmq

import (
	"fmt"

	amqp "github.com/streadway/amqp"
)

func (r *RabbitMQ) ConsumeMessage(queueName string, handler func(amqp.Delivery)) error {
	msgs, err := r.Channel.Consume(
		queueName,
		"",    //consumer tag,
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // аргументы
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages from queue: %w", err)
	}
	for msg := range msgs {
		handler(msg)
	}
	return nil
}
