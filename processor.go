package main

import (
	"strconv"
	"strings"

	"github.com/Syfaro/telegram-bot-api"
)

type Processor struct {
	hited  bool
	s      []string
	update tgbotapi.Update
}

func (p *Processor) hitter(f func(), command ...string) {
	if p.hited {
		return
	}
	for k := range command {
		if p.s[0] == command[k] {
			p.hited = true
			break
		}
	}
	if p.hited {
		f()
	}
}

func (p *Processor) start(command ...string) {
	f := func() {
		msg := tgbotapi.NewMessage(p.chatid(),
			YamlList2String(conf, "help"))
		bot.SendMessage(msg)
	}
	p.hitter(f, command...)
}

func (p *Processor) about(command ...string) {
	f := func() {
		msg := tgbotapi.NewMessage(p.chatid(),
			YamlList2String(conf, "help"))
		bot.SendMessage(msg)
	}
	p.hitter(f, command...)
}

func (p *Processor) otherResources(command ...string) {
	f := func() {
		msg := tgbotapi.NewMessage(p.chatid(),
			YamlList2String(conf, "其他资源"))
		bot.SendMessage(msg)
	}
	p.hitter(f, command...)
}

func (p *Processor) groups(command ...string) {
	f := func() {
		if p.update.Message.IsGroup() {
			msg := tgbotapi.NewMessage(p.chatid(),
				"使用姿势不对呢喵~ ＞▽＜\n本功能只限私聊使用")
			bot.SendMessage(msg)
			return
		}
		category := To2dSlice(categories, 3, 5)

		msg := tgbotapi.NewMessage(p.chatid(),
			"你想要查看哪些群组呢😋\n(为保护群组不被外星人攻击，"+
				"请勿将群链接转发到群组中，或者公布到网络上)")
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:        category,
			OneTimeKeyboard: true,
			ResizeKeyboard:  true,
		}
		bot.SendMessage(msg)
	}
	p.hitter(f, command...)
}

func (p *Processor) cancel(command ...string) {
	f := func() {
		if !p.update.Message.IsGroup() {
			msg := tgbotapi.NewMessage(p.chatid(),
				"群组娘已完成零态重置")
			p.setStatus("")
			msg.ReplyMarkup = tgbotapi.ReplyKeyboardHide{
				HideKeyboard: true,
			}
			bot.SendMessage(msg)
		}
	}
	p.hitter(f, command...)
}

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

func (p *Processor) google(command ...string) {
	f := func() {
		if len(p.s) >= 2 {
			q := strings.Join(p.s[1:], " ")
			msg := tgbotapi.NewMessage(p.chatid(), Google(q))
			msg.DisableWebPagePreview = true
			bot.SendMessage(msg)
		}
	}
	p.hitter(f, command...)
}

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

func (p *Processor) reload(command ...string) {
	f := func() {
		LoadConf()
		msg := tgbotapi.NewMessage(p.chatid(), "群组娘已完成弹药重装(ゝ∀･)")
		bot.SendMessage(msg)
	}
	p.hitter(f, command...)
}

func (p *Processor) _autoRule() {
	if p.update.Message.NewChatParticipant.ID != 0 {
		chatIDStr := strconv.Itoa(p.chatid())
		if rc.Exists("tgGroupAutoRule:" + chatIDStr).Val() {
			go func() {
				msg := tgbotapi.NewMessage(p.update.Message.NewChatParticipant.ID,
					"欢迎加入 "+p.update.Message.Chat.Title+"\n 以下是群组规则：")
				bot.SendMessage(msg)
				if rc.Exists("tgGroupRule:" + chatIDStr).Val() {
					msg := tgbotapi.NewMessage(
						p.update.Message.NewChatParticipant.ID,
						rc.Get("tgGroupRule:"+chatIDStr).Val())
					bot.SendMessage(msg)
				}
			}()
		}
	}
}

func (p *Processor) _default() {
	if p.hited {
		return
	}
	if !p.update.Message.IsGroup() {
		switch p.getStatus() {
		case "auth":
			p.auth(p.update.Message.Text)
		case "broadcast":
			p._broadcast(p.update.Message.Text)
			p.setStatus("")
		default:
			if categoriesSet.Has(p.update.Message.Text) {
				// custom keyboard reply
				msg := tgbotapi.NewMessage(p.chatid(),
					YamlList2String(conf, p.update.Message.Text))
				bot.SendMessage(msg)
			} else {
				p._turing(p.update.Message.Text)
			}
		}
	} else if p.update.Message.ReplyToMessage != nil &&
		p.update.Message.ReplyToMessage.From.UserName == bot.Self.UserName {
		p._turing(p.update.Message.Text)
	}
}

func (p *Processor) getStatus() string {
	if rc.Exists("tgStatus:" + strconv.Itoa(p.chatid())).Val() {
		return rc.Get("tgStatus:" + strconv.Itoa(p.chatid())).Val()
	}
	return ""
}

func (p *Processor) isMaster() bool {
	master, _ := conf.Get("master")
	if p.update.Message.From.UserName == master {
		return true
	}
	return false
}

func (p *Processor) chatid() int {
	return p.update.Message.Chat.ID
}

func (p *Processor) sendTyping(done chan bool) {
	go func() {
		typing := tgbotapi.NewChatAction(p.update.Message.Chat.ID,
			tgbotapi.ChatTyping)
		bot.SendChatAction(typing)
		done <- true
	}()
}

func (p *Processor) setStatus(status string) {
	if status == "" {
		rc.Del("tgStatus:" +
			strconv.Itoa(p.update.Message.Chat.ID))
		return
	} else {
		rc.Set("tgStatus:"+
			strconv.Itoa(p.update.Message.Chat.ID), status, -1)
	}
}
