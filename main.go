package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Syfaro/telegram-bot-api"
)

func main() {
	defer rc.Close()

	botname := bot.Self.UserName

	for update := range bot.Updates {

		// Ignore Outdated Updates
		if time.Since(time.Unix(int64(update.Message.Date), 0)) > time.Hour {
			continue
		}

		// Logger
		startWithSlash, _ := regexp.MatchString("^/", update.Message.Text)
		atBot, _ := regexp.MatchString("@"+botname, update.Message.Text)
		if update.Message.Chat.ID > 0 || startWithSlash || atBot {
			log.Printf("[%d](%s) -- [%s] -- %s",
				update.Message.Chat.ID, update.Message.Chat.Title,
				update.Message.From.UserName, update.Message.Text,
			)
		}

		// Field the message text
		s := strings.FieldsFunc(update.Message.Text,
			func(r rune) bool {
				switch r {
				case '\t', '\v', '\f', '\r', ' ', 0xA0:
					return true
				}
				return false
			})

		p := Processor{false, s, update}

		// Auto Rule When New Member Join Group
		if update.Message.NewChatParticipant.ID != 0 {
			chatIDStr := strconv.Itoa(update.Message.Chat.ID)
			if rc.Exists("tgGroupAutoRule:" + chatIDStr).Val() {
				go func() {
					msg := tgbotapi.NewMessage(update.Message.NewChatParticipant.ID,
						"欢迎加入 "+update.Message.Chat.Title+"\n 以下是群组规则：")
					bot.SendMessage(msg)
					if rc.Exists("tgGroupRule:" + chatIDStr).Val() {
						msg := tgbotapi.NewMessage(
							update.Message.NewChatParticipant.ID,
							rc.Get("tgGroupRule:"+chatIDStr).Val())
						bot.SendMessage(msg)
					}
				}()
			}
		}

		p.saveSticker()
		p.analytics()

		if len(s) > 0 {
			go func(p Processor) {
				p.start("/help", "/start", "/help@"+botname, "/start@"+botname)
				p.rule("/rule", "/rule@"+botname)
				p.about("/about", "/about@"+botname)
				p.otherResources("/other_resources", "/other_resources@"+botname)
				p.subscribe("/subscribe", "/subscribe@"+botname)
				p.unsubscribe("/unsubscribe", "/unsubscribe@"+botname)
				p.autoRule("/autorule")
				p.groups("/groups", "/groups@"+botname)
				p.cancel("/cancel")
				p.rand("/rand")
				p.setRule("/setrule")
				p.base64("/e64", "/d64")
				p.google("/gg")
				p.trans("/trans")
				p.setMan("/setman")
				p.rmMan("/rmman")
				p.man("/man")
				p.broadcast("/broadcast")
				p.reload("/reload")
				p.stat("/os", "/df", "/free", "/redis")
				p.statistics("/rain")
				p.turing("@" + botname)
				p._default()
			}(p)
		}
	}
}
