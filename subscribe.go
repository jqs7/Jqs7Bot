package main

import (
	"strconv"
	"strings"

	"github.com/Syfaro/telegram-bot-api"
)

func (p *Processor) subscribe(command ...string) {
	f := func() {
		chatIDStr := strconv.Itoa(p.chatid())
		isSubscribe, _ := strconv.ParseBool(rc.HGet("tgSubscribe",
			chatIDStr).Val())

		if p.update.Message.IsGroup() {
			msg := tgbotapi.NewMessage(p.chatid(),
				"群组订阅功能已取消，需要订阅功能的话，请私聊奴家呢o(￣ˇ￣)o")
			bot.SendMessage(msg)
			return
		}

		if !p.isAuthed() {
			p.sendQuestion()
			return
		}

		if isSubscribe {
			msg := tgbotapi.NewMessage(p.chatid(),
				"已经订阅过，就不要重复订阅啦😘")
			bot.SendMessage(msg)
		} else {
			rc.HSet("tgSubscribe", chatIDStr, strconv.FormatBool(true))
			rc.HIncrBy("tgSubscribeTimes", chatIDStr, 1)
			msg := tgbotapi.NewMessage(p.chatid(),
				"订阅成功\n以后奴家知道新的群组的话，会第一时间告诉你哟😊\n"+
					"(订阅仅对当前会话有效)")
			bot.SendMessage(msg)
		}
	}
	p.hitter(f, command...)
}

func (p *Processor) _broadcast(text string) {
	if p.isMaster() &&
		rc.Exists("tgSubscribe").Val() {
		subStates := rc.HGetAllMap("tgSubscribe").Val()

		for k, v := range subStates {
			chatid, _ := strconv.Atoi(k)
			subState, _ := strconv.ParseBool(v)

			if subState && chatid > 0 {
				loger.Infof("sending boardcast to %d ...", chatid)
				msg := tgbotapi.NewMessage(chatid, text)
				go func(k string) {
					bot.SendMessage(msg)
					loger.Infof("%s --- done", k)
				}(k)
			}
		}
	}
}

func (p *Processor) unsubscribe(command ...string) {
	f := func() {
		chatIDStr := strconv.Itoa(p.chatid())
		var msg tgbotapi.MessageConfig
		if rc.HExists("tgSubscribe", chatIDStr).Val() {
			rc.HDel("tgSubscribe", chatIDStr)
			times, _ := rc.HIncrBy("tgSubscribeTimes", chatIDStr, 1).Result()
			if times > 5 {
				msg = tgbotapi.NewMessage(p.chatid(),
					"订了退，退了订，你烦不烦嘛！！！⊂彡☆))∀`)`")
				rc.HDel("tgSubscribeTimes", chatIDStr)
			} else {
				msg = tgbotapi.NewMessage(p.chatid(),
					"好伤心，退订了就不能愉快的玩耍了呢😭")
			}
		} else {
			msg = tgbotapi.NewMessage(p.chatid(),
				"你都还没订阅，让人家怎么退订嘛！o(≧口≦)o")
		}
		bot.SendMessage(msg)
	}
	p.hitter(f, command...)
}

func (p *Processor) broadcast(command ...string) {
	f := func() {
		if len(p.s) == 1 && p.isMaster() &&
			!p.update.Message.IsGroup() {
			msg := tgbotapi.NewMessage(p.chatid(),
				"Send me the Broadcast (＾o＾)ﾉ")
			bot.SendMessage(msg)
			p.setStatus("broadcast")
			return
		}
		if len(p.s) >= 2 {
			text := strings.Join(p.s[1:], " ")
			p._broadcast(text)
		}
	}
	p.hitter(f, command...)
}
