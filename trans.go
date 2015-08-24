package main

import (
	"strings"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/st3v/translator/microsoft"
)

func (p *Processor) trans(command ...string) {
	f := func() {
		if p.update.Message.ReplyToMessage != nil &&
			p.update.Message.ReplyToMessage.Text != "" &&
			len(p.s) < 2 {
			in := p.update.Message.ReplyToMessage.Text
			result := p.translator(in)
			msg := tgbotapi.NewMessage(p.chatid(), result)
			bot.SendMessage(msg)
		} else if len(p.s) >= 2 {
			in := strings.Join(p.s[1:], " ")
			result := p.translator(in)
			msg := tgbotapi.NewMessage(p.chatid(), result)
			bot.SendMessage(msg)
		}
	}
	p.hitter(f, command...)
}

func MsDetect(clientID, clientSecret, in string) (string, error) {
	t := microsoft.NewTranslator(clientID, clientSecret)
	return t.Detect(in)
}

func MsTrans(clientID, clientSecret, in string) (out string) {
	t := microsoft.NewTranslator(clientID, clientSecret)
	from, err := MsDetect(clientID, clientSecret, in)
	if err != nil {
		return "警报！弹药系统过载！请放宽后重试"
	}
	switch from {
	case "zh-CHS", "zh-CHT":
		out, err = t.Translate(in, from, "en")
	default:
		out, err = t.Translate(in, from, "zh-CHS")
	}
	if err != nil {
		return "可怜的群组娘被母舰放逐了X﹏X"
	}
	return out
}

func (p *Processor) translator(in string) string {
	result := make(chan string)
	typingChan := make(chan bool)
	p.sendTyping(typingChan)
	go func() {
		result <- MsTrans(msID, msSecret, in)
	}()
	<-typingChan
	return <-result
}
