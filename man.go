package main

import (
	"strconv"
	"strings"

	"github.com/Syfaro/telegram-bot-api"
)

func (p *Processor) listMan() {
	chatid := p.update.Message.Chat.ID
	if p.update.Message.IsGroup() {
		var result string
		fields := rc.HGetAllMap("tgMan:" +
			strconv.Itoa(p.update.Message.Chat.ID)).Val()
		for k := range fields {
			result += k + "\n"
		}
		msg := tgbotapi.NewMessage(chatid, result)
		bot.SendMessage(msg)
	}
}

func (p *Processor) setMan(command ...string) {
	f := func() {
		if len(p.s) >= 3 {
			value := strings.Join(p.s[2:], " ")
			if !p.isAuthed() {
				p.sendQuestion()
				return
			}

			if p.update.Message.IsGroup() {
				rc.HSet("tgMan:"+strconv.Itoa(p.chatid()),
					p.s[1], value)
				msg := tgbotapi.NewMessage(p.chatid(), p.s[1]+":\n"+value)
				bot.SendMessage(msg)
			}
		}
	}
	p.hitter(f, command...)
}

func (p *Processor) rmMan(command ...string) {
	f := func() {
		fields := p.s[1:]
		for k := range fields {
			rc.HDel("tgMan:"+
				strconv.Itoa(p.update.Message.Chat.ID), fields[k])
		}
		p.listMan()
	}
	p.hitter(f, command...)
}

func (p *Processor) man(command ...string) {
	f := func() {
		if len(p.s) == 1 {
			p.listMan()
		} else {
			if !p.update.Message.IsGroup() {
				return
			}
			if p.s[1] == "man" && !rc.HExists("tgMan:"+
				strconv.Itoa(p.chatid()), "man").Val() {
				msg := tgbotapi.NewMessage(p.chatid(),
					"你在慢慢个什么鬼啦！(σﾟ∀ﾟ)σ")
				bot.SendMessage(msg)
				return
			}
			msg := tgbotapi.NewMessage(p.chatid(),
				rc.HGet("tgMan:"+strconv.Itoa(p.chatid()), p.s[1]).Val())
			bot.SendMessage(msg)
		}
	}
	p.hitter(f, command...)
}
