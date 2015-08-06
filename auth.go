package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Syfaro/telegram-bot-api"
)

func (u *Updater) Auth(answer string) {
	qs := GetQuestions(u.conf, "questions")
	index := time.Now().Hour() % len(qs)
	answer = strings.ToLower(answer)
	answer = strings.TrimSpace(answer)
	if u.update.Message.Chat.ID > 0 {
		if u.IsAuthed() {
			msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
				"已经验证过了，你还想验证，你是不是傻？⊂彡☆))д`)`")
			msg.ReplyToMessageID = u.update.Message.MessageID
			u.bot.SendMessage(msg)
			return
		}

		if qs[index].A.Has(answer) {
			u.redis.SAdd("tgAuthUser", strconv.Itoa(u.update.Message.From.ID))
			log.Printf("%d --- %s Auth OK",
				u.update.Message.From.ID, u.update.Message.From.UserName)
			msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
				"验证成功喵~！\n原来你不是外星人呢😊")
			u.SetStatus("")
			u.bot.SendMessage(msg)
		} else {
			log.Printf("%d --- %s Auth Fail",
				u.update.Message.From.ID, u.update.Message.From.UserName)
			msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
				"答案不对不对！你一定是外星人！不跟你玩了喵！\n"+
					"重新验证一下吧\n请问："+qs[index].Q)
			u.bot.SendMessage(msg)
		}
	}
}

func (u *Updater) IsAuthed() bool {
	if u.redis.SIsMember("tgAuthUser",
		strconv.Itoa(u.update.Message.From.ID)).Val() {
		return true
	}
	return false
}

func (u *Updater) SendQuestion() {
	if u.update.Message.Chat.ID < 0 {
		msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
			"需要通过中文验证之后才能使用本功能哟~\n"+
				"点击奴家的头像进入私聊模式，进行验证吧")
		u.bot.SendMessage(msg)
		return
	}
	qs := GetQuestions(u.conf, "questions")
	index := time.Now().Hour() % len(qs)
	msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
		"需要通过中文验证之后才能使用本功能哟~\n请问："+
			qs[index].Q+"\n把答案发给奴家就可以了呢")
	u.SetStatus("auth")
	u.bot.SendMessage(msg)
}
