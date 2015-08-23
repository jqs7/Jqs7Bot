package main

import (
	"encoding/base64"
	"strings"
	"unicode/utf8"

	"github.com/Syfaro/telegram-bot-api"
)

func (p *Processor) base64(command ...string) {
	f := func() {
		switch p.s[0] {
		case "/e64":
			if p.update.Message.ReplyToMessage != nil &&
				p.update.Message.ReplyToMessage.Text != "" {
				msg := tgbotapi.NewMessage(p.chatid(),
					E64(p.update.Message.ReplyToMessage.Text))
				bot.SendMessage(msg)
			} else if len(p.s) >= 2 {
				in := strings.Join(p.s[1:], " ")
				msg := tgbotapi.NewMessage(p.chatid(), E64(in))
				bot.SendMessage(msg)
			}
		case "/d64":
			if len(p.s) >= 2 {
				in := strings.Join(p.s[1:], " ")
				msg := tgbotapi.NewMessage(p.chatid(), D64(in))
				bot.SendMessage(msg)
			}
		}
	}
	p.hitter(f, command...)
}

func E64(in string) string {
	return base64.StdEncoding.EncodeToString([]byte(in))
}

func D64(in string) string {
	out, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return "解码系统出现故障，请查看弹药是否填充无误"
	}
	if utf8.Valid(out) {
		return string(out)
	}
	return "解码结果包含不明物体，群组娘已将之上交国家"
}
