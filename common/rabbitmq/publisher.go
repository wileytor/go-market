package rabbitmq

import amqp "github.com/streadway/amqp"

func (r *RabbitMQ) PublishMessage(queueName string, message []byte)error {
	err := r.Channel.Publish(
		"", // default exchange
		queueName,
		false, // mandatory,  если true, то сообщение не будет утеряно, если очередь не найдена.
		false, // immediate, если true, сообщение отправляется только если есть потребители, готовые принять.
		amqp.Publishing{
            ContentType: "application/json",
            Body:        message,
        },
	)
	return err
}