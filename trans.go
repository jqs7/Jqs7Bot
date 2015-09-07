package main

import (
	"strings"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/st3v/translator"
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

type MsTrans struct {
	t translator.Translator
}

func (m *MsTrans) New() {
	m.t = microsoft.NewTranslator(msID, msSecret)
}

func (m *MsTrans) Detect(in string) (string, error) {
	return m.t.Detect(in)
}

func (m *MsTrans) Trans(in, from, to string) (string, error) {
	return m.t.Translate(in, from, to)
}

func ZhTrans(in string) (out string) {
	m := &MsTrans{}
	m.New()
	from, err := m.Detect(in)
	if err != nil {
		return "警报！弹药系统过载！请放宽后重试"
	}
	switch from {
	case "zh-CHS", "zh-CHT":
		out, err = m.Trans(in, from, "en")
	default:
		out, err = m.Trans(in, from, "zh-CHS")
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
		result <- ZhTrans(in)
	}()
	<-typingChan
	return <-result
}
