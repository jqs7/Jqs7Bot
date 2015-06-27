package main

import (
	"log"
	"strings"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/astaxie/beego/config"
)

func main() {
	groups, err := config.NewConfig("ini", "groups.conf")
	if err != nil {
		log.Println(err)
	}

	bot, err := tgbotapi.NewBotAPI(groups.String("botapi"))
	botname := groups.String("botusername")
	if err != nil {
		log.Println(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.UpdatesChan(u)

	for update := range updates {
		log.Println(update.Message.From.UserName, update.Message.Text)

		u := Updater{bot, update.Message.Chat.ID}

		switch update.Message.Text {
		case "/help", "/start", "/help@" + botname, "/start@" + botname:
			u.NewMessage(groups.String("help"))
		case "/rules", "/rules@" + botname:
			u.NewMessage(groups.String("rules"))
		case "/about", "/about@" + botname:
			u.NewMessage(groups.String("about"))
		case "/linux", "/linux@" + botname:
			u.NewMessage(groups.String("linux"))
		case "/programming", "/programming@" + botname:
			u.NewMessage(groups.String("programming"))
		case "/software", "/software@" + botname:
			u.NewMessage(groups.String("software"))
		case "/videos", "/videos@" + botname:
			u.NewMessage(groups.String("影音"))
		case "/acg", "/acg@" + botname:
			u.NewMessage(groups.String("ACG"))
		case "/it", "/it@" + botname:
			u.NewMessage(groups.String("IT"))
		case "/free_chat", "/free_chat@" + botname:
			u.NewMessage(groups.String("闲聊"))
		case "/resources", "/resources@" + botname:
			u.NewMessage(groups.String("资源"))
		case "/same-city", "/same_city@" + botname:
			u.NewMessage(groups.String("同城"))
		case "/others", "/others@" + botname:
			u.NewMessage(groups.String("Others"))
		case "/other_resources", "/other_resources@" + botname:
			u.NewMessage(groups.String("其他资源"))
		}

	}
}

type Updater struct {
	bot    *tgbotapi.BotAPI
	chatId int
}

func (u *Updater) NewMessage(msgText string) {
	msgText = strings.Replace(msgText, "\\n", "\n", -1)
	msg := tgbotapi.NewMessage(u.chatId, msgText)
	u.bot.SendMessage(msg)
}
