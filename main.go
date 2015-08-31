package main

import (
	"regexp"
	"strings"
	"time"
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
		loger.Debugf("%+v", update)
		startWithSlash, _ := regexp.MatchString("^/", update.Message.Text)
		atBot, _ := regexp.MatchString("@"+botname, update.Message.Text)
		if update.Message.Chat.ID > 0 || startWithSlash || atBot {
			loger.Infof("[%d](%s) -- [%s] -- %s",
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

		p._autoRule()
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
				p.rss("/rss")
				p.rmrss("/rmrss")
				p.turing("@" + botname)
				p._default()
			}(p)
		}
	}
}
