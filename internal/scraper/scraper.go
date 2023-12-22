package scraper

import (
	"github.com/tebeka/selenium"
	"strings"
)

func InitChromeDriver() (selenium.WebDriver, *selenium.Service, error) {
	const (
		chromeDriverPath = "./chromedriver"
		port             = 9515
	)

	opts := []selenium.ServiceOption{
		//selenium.Output(os.Stderr),
	}

	service, err := selenium.NewChromeDriverService(chromeDriverPath, port, opts...)
	if err != nil {
		panic(err)
	}

	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, "http://localhost:9515/wd/hub")
	if err != nil {
		panic(err)
	}

	return wd, service, nil
}

func GetChatCount(wd selenium.WebDriver) (string, error) {
	div, err := wd.FindElement(selenium.ByXPATH, "//div[contains(@class, 'chat_count')]")
	if err != nil {
		return "", err
	}

	return div.Text()
}

func GetNewMsgs(wd selenium.WebDriver, seenMsgs map[string]bool) ([]string, error) {
	spans, err := wd.FindElements(selenium.ByXPATH, "//span[contains(@class, 'chat-message-default_message__')]")
	if err != nil {
		return nil, err
	}

	var newMsgs []string
	for _, s := range spans {
		msgTxt, _ := s.Text()
		if _, ok := seenMsgs[msgTxt]; !ok {
			seenMsgs[msgTxt] = true
			if err != nil {
				msgTxt = "error getting message"
			}
			newMsgs = append(newMsgs, msgTxt)
		}
	}

	return newMsgs, nil
}

type Character struct {
	Names    []string
	Mentions int
}

func BuildNameMap(characters []*Character) map[string]*Character {
	nameMap := make(map[string]*Character)

	for _, char := range characters {
		for _, name := range char.Names {
			nameMap[name] = char
		}
	}

	return nameMap
}

func FindCharactersInMsg(nameCharMap map[string]*Character, msg string, resCh chan *Character) {
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
