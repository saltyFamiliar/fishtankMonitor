package messages

import (
	"fmt"
	"github.com/streadway/amqp"
)

func InitQueue(qname, url string) (*amqp.Channel, *amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	_, err = ch.QueueDeclare(
		qname,
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, nil, err
	}

	return ch, conn, nil
}

func PublishStringMetric(ch *amqp.Channel, metric, key string) error {
	err := ch.Publish(
		"",
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(metric),
		})

	return err
}

func DeclareAndConsume(ch *amqp.Channel, name string) (<-chan amqp.Delivery, error) {
	q, err := ch.QueueDeclare(name, false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("declare %s queue: %v", name, err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("consume %s queue: %v", name, err)
	}

	return msgs, nil
}
