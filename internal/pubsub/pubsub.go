package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

type AckType int

const (
	DurableQueue SimpleQueueType = iota
	TransientQueue
)

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

var QueueState = map[SimpleQueueType]string{
	DurableQueue:   "durable",
	TransientQueue: "transient",
}

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonBody, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{ContentType: "application/json", Body: jsonBody})
	if err != nil {
		return err
	}

	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var bodyBuffer bytes.Buffer
	gobEncoder := gob.NewEncoder(&bodyBuffer)
	err := gobEncoder.Encode(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{ContentType: "application/gob", Body: bodyBuffer.Bytes()})
	if err != nil {
		return err
	}

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	queue, err := ch.QueueDeclare(queueName,
		queueType.isDurable(),
		queueType.isTransient(),
		queueType.isTransient(),
		false,
		amqp.Table{"x-dead-letter-exchange": "peril_dlx"})
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	err = ch.QueueBind(queue.Name, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return ch, queue, nil
}

func (sqt SimpleQueueType) isDurable() bool {
	return sqt == DurableQueue
}

func (sqt SimpleQueueType) isTransient() bool {
	return sqt == TransientQueue
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	delivery, err := ch.Consume(queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range delivery {
			var msg T
			json.Unmarshal(d.Body, &msg)
			switch acktype := handler(msg); acktype {
			case Ack:
				d.Ack(false)
			case NackRequeue:
				d.Nack(false, true)
			case NackDiscard:
				d.Nack(false, false)
			}
		}
	}()
	return nil
}
