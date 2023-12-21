package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"

	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial("amqp://localhost:5672/")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"chat_count", // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		panic(err)
	}

	nameCh, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	qname, err := nameCh.QueueDeclare(
		"char_names",
		false, false, false, false, nil,
	)

	// Consume the messages
	nameMsgs, err := nameCh.Consume(
		qname.Name,
		"",
		true, false, false, false, nil,
	)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8080", nil)
	}()

	chatUsers := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "viewers_total",
		Help: "total viewers",
	})
	prometheus.MustRegister(chatUsers)

	charCounters := map[string]prometheus.Counter{
		"cole":     prometheus.NewCounter(prometheus.CounterOpts{Name: "cole", Help: "cole mentions"}),
		"jc":       prometheus.NewCounter(prometheus.CounterOpts{Name: "jc", Help: "jc mentions"}),
		"jimmy":    prometheus.NewCounter(prometheus.CounterOpts{Name: "jimmy", Help: "jimmy mentions"}),
		"megan":    prometheus.NewCounter(prometheus.CounterOpts{Name: "megan", Help: "megan mentions"}),
		"shinji":   prometheus.NewCounter(prometheus.CounterOpts{Name: "shinji", Help: "shinji mentions"}),
		"summer":   prometheus.NewCounter(prometheus.CounterOpts{Name: "summer", Help: "summer mentions"}),
		"tayleigh": prometheus.NewCounter(prometheus.CounterOpts{Name: "tayleigh", Help: "tayleigh mentions"}),
		"trisha":   prometheus.NewCounter(prometheus.CounterOpts{Name: "trisha", Help: "trisha mentions"}),
		"brian":    prometheus.NewCounter(prometheus.CounterOpts{Name: "brian", Help: "brian mentions"}),
		"tj":       prometheus.NewCounter(prometheus.CounterOpts{Name: "tj", Help: "tj mentions"}),
	}

	//reset counter to position before restart (wow)
	charCounters["cole"].Add(1245)
	charCounters["jimmy"].Add(2471)
	charCounters["jc"].Add(927)
	charCounters["brian"].Add(702)
	charCounters["megan"].Add(889)
	charCounters["shinji"].Add(1403)
	charCounters["summer"].Add(2594)
	charCounters["tayleigh"].Add(539)
	charCounters["trisha"].Add(852)
	charCounters["tj"].Add(2018)

	for name := range charCounters {
		prometheus.MustRegister(charCounters[name])
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			fmt.Printf("Received: %s\n", d.Body)
			count, err := strconv.Atoi(string(d.Body))
			if err != nil {
				fmt.Println("Couldn't convert count to int")
			}
			if count == 0 {
				continue
			}
			chatUsers.Set(float64(count))
		}
	}()

	go func() {
		for name := range nameMsgs {
			nameStr := string(name.Body)
			println("received name: ", nameStr)
			charCounters[nameStr].Add(1)
		}
	}()

	fmt.Println("Waiting for messages.")
	<-forever
}
