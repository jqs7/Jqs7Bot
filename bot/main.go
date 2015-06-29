package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/kylelemons/go-gypsy/yaml"
)

func main() {
	groups, err := yaml.ReadFile("botconf.yaml")
	if err != nil {
		log.Println(err)
	}

	botname, err := groups.Get("botusername")
	botapi, err := groups.Get("botapi")
	if err != nil {
		log.Println(err)
	}

	bot, err := tgbotapi.NewBotAPI(botapi)
	if err != nil {
		log.Println(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.UpdatesChan(u)

	for update := range updates {
		log.Printf("[%s]  %s", update.Message.From.UserName, update.Message.Text)

		u := Updater{bot, groups, update.Message.Chat.ID}

		switch update.Message.Text {

		case "/help", "/start", "/help@" + botname, "/start@" + botname:
			go u.SendStrings("help")

		case "/rules", "/rules@" + botname:
			go u.SendStrings("rules")

		case "/about", "/about@" + botname:
			go u.SendStrings("about")

		case "/linux", "/linux@" + botname:
			go u.SendStrings("Linux")

		case "/programming", "/programming@" + botname:
			go u.SendStrings("Programming")

		case "/software", "/software@" + botname:
			go u.SendStrings("Software")

		case "/videos", "/videos@" + botname:
			go u.SendStrings("影音")

		case "/sci_fi", "/sci_fi@" + botname:
			go u.SendStrings("科幻")

		case "/acg", "/acg@" + botname:
			go u.SendStrings("ACG")

		case "/it", "/it@" + botname:
			go u.SendStrings("IT")

		case "/free_chat", "/free_chat@" + botname:
			go u.SendStrings("闲聊")

		case "/resources", "/resources@" + botname:
			go u.SendStrings("资源")

		case "/same_city", "/same_city@" + botname:
			go u.SendStrings("同城")

		case "/others", "/others@" + botname:
			go u.SendStrings("Others")

		case "/other_resources", "/other_resources@" + botname:
			go u.SendStrings("其他资源")

		}
	}
}

type Updater struct {
	bot    *tgbotapi.BotAPI
	config *yaml.File
	chatId int
}

func (u *Updater) SendStrings(msgText string) error {
	count, err := u.config.Count(msgText)
	if err != nil {
		log.Println(err)
		return err
	}
	var resultGroup []string
	for i := 0; i < count; i++ {
		v, err := u.config.Get(msgText + "[" + strconv.Itoa(i) + "]")
		if err != nil {
			log.Println(err)
			return err
		}
		resultGroup = append(resultGroup, v)
	}
	result := strings.Join(resultGroup, "\n")
	result = strings.Replace(result, "\\n", "", -1)
	msg := tgbotapi.NewMessage(u.chatId, result)
	u.bot.SendMessage(msg)
	return nil
}
