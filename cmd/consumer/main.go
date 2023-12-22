package main

import (
	. "fishScraper/internal/consumer"
	"fishScraper/internal/messages"
	"fishScraper/internal/utils"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/streadway/amqp"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	filepath := "mentions.csv"

	conn, err := amqp.Dial("amqp://localhost:5672/")
	utils.Must("connection to amqp server", err)

	ch, err := conn.Channel()
	utils.Must("create chat_count connection channel", err)

	defer func() {
		_ = conn.Close()
		_ = ch.Close()
	}()

	chatCount, err := messages.DeclareAndConsume(ch, "chat_count")
	charNames, err := messages.DeclareAndConsume(ch, "char_names")
	msgTotal, err := messages.DeclareAndConsume(ch, "message_total")

	chatUsers := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "viewers_total",
		Help: "total viewers",
	})
	prometheus.MustRegister(chatUsers)

	charCounters := map[string]*Character{
		"cole":     {prometheus.NewCounter(prometheus.CounterOpts{Name: "cole", Help: "cole mentions"}), 0},
		"jc":       {prometheus.NewCounter(prometheus.CounterOpts{Name: "jc", Help: "jc mentions"}), 0},
		"jimmy":    {prometheus.NewCounter(prometheus.CounterOpts{Name: "jimmy", Help: "jimmy mentions"}), 0},
		"megan":    {prometheus.NewCounter(prometheus.CounterOpts{Name: "megan", Help: "megan mentions"}), 0},
		"shinji":   {prometheus.NewCounter(prometheus.CounterOpts{Name: "shinji", Help: "shinji mentions"}), 0},
		"summer":   {prometheus.NewCounter(prometheus.CounterOpts{Name: "summer", Help: "summer mentions"}), 0},
		"tayleigh": {prometheus.NewCounter(prometheus.CounterOpts{Name: "tayleigh", Help: "tayleigh mentions"}), 0},
		"trisha":   {prometheus.NewCounter(prometheus.CounterOpts{Name: "trisha", Help: "trisha mentions"}), 0},
		"brian":    {prometheus.NewCounter(prometheus.CounterOpts{Name: "brian", Help: "brian mentions"}), 0},
		"tj":       {prometheus.NewCounter(prometheus.CounterOpts{Name: "tj", Help: "tj mentions"}), 0},
	}
	ReadCounts(filepath, charCounters)
	for name := range charCounters {
		prometheus.MustRegister(charCounters[name].Counter)
	}

	messageCounter := prometheus.NewCounter(prometheus.CounterOpts{Name: "message_total", Help: "total messages"})
	prometheus.MustRegister(messageCounter)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		utils.Must("create prometheus metrics server", http.ListenAndServe(":8080", nil))
	}()

	go func() {
		for d := range chatCount {
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
		for name := range charNames {
			nameStr := string(name.Body)
			println("received name: ", nameStr)
			charCounters[nameStr].Counter.Add(1)
			charCounters[nameStr].Count++
		}
	}()

	go func() {
		for msgCount := range msgTotal {
			totalStr := string(msgCount.Body)
			floatTotal, _ := strconv.ParseFloat(totalStr, 64)
			println("total messages: ", floatTotal)
			messageCounter.Add(floatTotal)
		}
	}()

	fmt.Println("Waiting for messages.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh
	WriteCounts(filepath, charCounters)
}
