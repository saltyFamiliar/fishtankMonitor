package main

import (
	"fishScraper/internal/messages"
	"fishScraper/internal/scraper"
	"fishScraper/internal/utils"
	"fmt"
	"strings"
	"time"
)

type Character struct {
	names    []string
	mentions int
}

func BuildNameMap(characters []*Character) map[string]*Character {
	nameMap := make(map[string]*Character)

	for _, char := range characters {
		for _, name := range char.names {
			nameMap[name] = char
		}
	}

	return nameMap
}

func findCharactersInMsg(nameCharMap map[string]*Character, msg string, resCh chan *Character) {
	mentioned := make(map[*Character]bool)
	words := strings.Fields(msg)
	for _, w := range words {
		if char, ok := nameCharMap[strings.ToLower(w)]; ok {
			if !mentioned[char] {
				resCh <- char
			}
			mentioned[char] = true
		}
	}
}

var charList = []*Character{
	{[]string{"cole"}, 0},
	{[]string{"jc"}, 0},
	{[]string{"jimmy", "jimbo"}, 0},
	{[]string{"megan", "meg", "bert"}, 0},
	{[]string{"shinji"}, 0},
	{[]string{"summer"}, 0},
	{[]string{"tayleigh", "tay"}, 0},
	{[]string{"trisha", "trish"}, 0},
	{[]string{"brian"}, 0},
	{[]string{"tj"}, 0},
}

func main() {
	wd, service, err := scraper.InitChromeDriver()
	utils.Must("initialize chrome driver and selenium service", err)
	utils.Must("connect to website", wd.Get("https://fishtank.live"))

	mqUrl := "amqp://localhost:5672"
	mqCh, conn, err := messages.InitQueue("chat_count", mqUrl)
	utils.Must("initialize chat_count queue", err)
	nameCh, nameConn, err := messages.InitQueue("char_names", mqUrl)
	utils.Must("initialize char_names queue", err)

	defer func() {
		defer utils.Must("close chrome driver", wd.Close())
		defer utils.Must("close selenium service", service.Stop())
		defer utils.Must("close chat_count mq connection", conn.Close())
		defer utils.Must("close chat_count mq channel", mqCh.Close())
		defer utils.Must("close char_names mq connection", nameConn.Close())
		defer utils.Must("close char_names mq channel", nameCh.Close())
	}()

	fmt.Println("Press enter when you've logged in and the chat count has loaded.")
	_, _ = fmt.Scanln()

	//poll users online
	go func() {
		for {
			time.Sleep(30 * time.Second)
			count, err := scraper.GetChatCount(wd)
			if err != nil || count == "00000" {
				println("Error encountered")
				continue
			}

			err = messages.PublishStringMetric(mqCh, count, "chat_count")
			if err != nil {
				println("problem publishing to rabbitMQ")
			}
		}
	}()

	//get new messages
	msgCh := make(chan []string)
	go func() {
		seenMsgs := make(map[string]bool)
		for {
			time.Sleep(2 * time.Second)
			newMsgs, err := scraper.GetNewMsgs(wd, seenMsgs)
			if err != nil {
				println("Error encountered")
			}

			msgCh <- newMsgs
		}
	}()

	mentionCh := make(chan *Character)

	//publish mentions
	go func() {
		for char := range mentionCh {
			err = messages.PublishStringMetric(nameCh, char.names[0], "char_names")
			if err != nil {
				println("problem publishing name")
			}

		}
	}()

	//search for mentions
	nameCharMap := BuildNameMap(charList)
	for msgs := range msgCh {
		for _, msg := range msgs {
			go findCharactersInMsg(nameCharMap, msg, mentionCh)
		}
	}
}
