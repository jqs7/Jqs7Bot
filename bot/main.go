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
		log.Println(update.Message.From.UserName, update.Message.Text)

		u := Updater{bot, groups, update.Message.Chat.ID}

		switch update.Message.Text {

		case "/help", "/start", "/help@" + botname, "/start@" + botname:
			u.SendStrings("help")

		case "/rules", "/rules@" + botname:
			u.SendStrings("rules")

		case "/about", "/about@" + botname:
			u.SendStrings("about")

		case "/linux", "/linux@" + botname:
			u.SendStrings("Linux")

		case "/programming", "/programming@" + botname:
			u.SendStrings("Programming")

		case "/software", "/software@" + botname:
			u.SendStrings("Software")

		case "/videos", "/videos@" + botname:
			u.SendStrings("影音")

		case "/sci_fi", "/sci_fi@" + botname:
			u.SendStrings("科幻")

		case "/acg", "/acg@" + botname:
			u.SendStrings("ACG")

		case "/it", "/it@" + botname:
			u.SendStrings("IT")

		case "/free_chat", "/free_chat@" + botname:
			u.SendStrings("闲聊")

		case "/resources", "/resources@" + botname:
			u.SendStrings("资源")

		case "/same_city", "/same_city@" + botname:
			u.SendStrings("同城")

		case "/others", "/others@" + botname:
			u.SendStrings("Others")

		case "/other_resources", "/other_resources@" + botname:
			u.SendStrings("其他资源")

		}
	}
}

type Updater struct {
	bot    *tgbotapi.BotAPI
	config *yaml.File
	chatId int
}

func (u *Updater) SendString(msgText string) error {
	msgText, err := u.config.Get(msgText)
	if err != nil {
		return err
	}
	msgText = strings.Replace(msgText, "\\n", "\n", -1)
	msg := tgbotapi.NewMessage(u.chatId, msgText)
	u.bot.SendMessage(msg)
	return nil
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
		v = strings.Replace(v, "|-", " ", -1)
		resultGroup = append(resultGroup, v)
	}
	result := strings.Join(resultGroup, "\n")
	msg := tgbotapi.NewMessage(u.chatId, result)
	u.bot.SendMessage(msg)
	return nil
}
