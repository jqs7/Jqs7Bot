package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/Syfaro/telegram-bot-api"
)

func (p *Processor) setRule(command ...string) {
	f := func() {
		if len(p.s) < 2 ||
			!p.update.Message.IsGroup() {
			return
		}
		rule := strings.Join(p.s[1:], " ")
		if p.isAuthed() {
			chatIDStr := strconv.Itoa(p.chatid())
			log.Printf("setting rule %s to %s", rule, chatIDStr)
			rc.Set("tgGroupRule:"+chatIDStr, rule, -1)
			msg := tgbotapi.NewMessage(p.chatid(),
				"新的群组规则Get！✔️\n以下是新的规则：\n\n"+rule)
			bot.SendMessage(msg)
		} else {
			p.sendQuestion()
		}
	}
	p.hitter(f, command...)
}

func (p *Processor) autoRule(command ...string) {
	f := func() {
		if p.update.Message.IsGroup() {
			chatIDStr := strconv.Itoa(p.chatid())
			if rc.Exists("tgGroupAutoRule:" + chatIDStr).Val() {
				rc.Del("tgGroupAutoRule:" + chatIDStr)
				msg := tgbotapi.NewMessage(p.chatid(),
					"AutoRule Disabled!")
				bot.SendMessage(msg)
			} else {
				rc.Set("tgGroupAutoRule:"+chatIDStr,
					strconv.FormatBool(true), -1)
				msg := tgbotapi.NewMessage(p.chatid(),
					"AutoRule Enabled!")
				bot.SendMessage(msg)
			}
		}
	}
	p.hitter(f, command...)
}

func (p *Processor) rule(command ...string) {
	f := func() {
		chatIDStr := strconv.Itoa(p.chatid())
		if rc.Exists("tgGroupRule:" + chatIDStr).Val() {
			msg := tgbotapi.NewMessage(p.chatid(),
				rc.Get("tgGroupRule:"+chatIDStr).Val())
			bot.SendMessage(msg)
		} else {
			msg := tgbotapi.NewMessage(p.chatid(),
				YamlList2String(conf, "rules"))
			bot.SendMessage(msg)
		}
	}
	p.hitter(f, command...)
}
