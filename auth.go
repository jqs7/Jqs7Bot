package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Syfaro/telegram-bot-api"
)

func (p *Processor) auth(answer string) {
	qs := GetQuestions(conf, "questions")
	index := time.Now().Hour() % len(qs)
	answer = strings.ToLower(answer)
	answer = strings.TrimSpace(answer)
	if !p.update.Message.IsGroup() {
		if p.isAuthed() {
			msg := tgbotapi.NewMessage(p.chatid(),
				"已经验证过了，你还想验证，你是不是傻？⊂彡☆))д`)`")
			msg.ReplyToMessageID = p.update.Message.MessageID
			bot.SendMessage(msg)
			return
		}

		if qs[index].A.Has(answer) {
			rc.SAdd("tgAuthUser", strconv.Itoa(p.update.Message.From.ID))
			log.Printf("%d --- %s Auth OK",
				p.update.Message.From.ID, p.update.Message.From.UserName)
			msg := tgbotapi.NewMessage(p.chatid(),
				"验证成功喵~！\n原来你不是外星人呢😊")
			p.setStatus("")
			bot.SendMessage(msg)
			p.start("/start")
		} else {
			log.Printf("%d --- %s Auth Fail",
				p.update.Message.From.ID, p.update.Message.From.UserName)
			msg := tgbotapi.NewMessage(p.chatid(),
				"答案不对不对！你一定是外星人！不跟你玩了喵！\n"+
					"重新验证一下吧\n请问："+qs[index].Q)
			bot.SendMessage(msg)
		}
	}
}

func (p *Processor) isAuthed() bool {
	if rc.SIsMember("tgAuthUser",
		strconv.Itoa(p.update.Message.From.ID)).Val() {
		return true
	}
	return false
}

func (p *Processor) sendQuestion() {
	if p.update.Message.Chat.ID < 0 {
		msg := tgbotapi.NewMessage(p.update.Message.Chat.ID,
			"需要通过中文验证之后才能使用本功能哟~\n"+
				"点击奴家的头像进入私聊模式，进行验证吧")
		bot.SendMessage(msg)
		return
	}
	qs := GetQuestions(conf, "questions")
	index := time.Now().Hour() % len(qs)
	msg := tgbotapi.NewMessage(p.update.Message.Chat.ID,
		"需要通过中文验证之后才能使用本功能哟~\n请问："+
			qs[index].Q+"\n把答案发给奴家就可以了呢")
	p.setStatus("auth")
	bot.SendMessage(msg)
}
