package queue

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func Connect(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	rmq := &RabbitMQ{
		conn:    conn,
		channel: ch,
	}

	// Declara as filas
	if err := rmq.declareQueues(); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return rmq, nil
}

func (r *RabbitMQ) declareQueues() error {
	queues := []string{
		"payment.created",
		"payment.approved",
		"utmify.pending",
		"utmify.approved",
	}

	for _, queue := range queues {
		_, err := r.channel.QueueDeclare(
			queue, // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RabbitMQ) Publish(queue string, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return r.channel.Publish(
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (r *RabbitMQ) Consume(queue string, handler func([]byte) error) error {
	msgs, err := r.channel.Consume(
		queue, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				log.Printf("Erro ao processar mensagem da fila %s: %v", queue, err)
				msg.Nack(false, true) // requeue
			} else {
				msg.Ack(false)
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
