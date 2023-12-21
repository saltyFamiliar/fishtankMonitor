package messages

import "github.com/streadway/amqp"

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
