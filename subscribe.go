package main

import (
	"log"
	"strconv"

	"github.com/Syfaro/telegram-bot-api"
)

func (u *Updater) Subscribe() {
	chatIDStr := strconv.Itoa(u.update.Message.Chat.ID)
	isSubscribe, _ := strconv.ParseBool(u.redis.HGet("tgSubscribe",
		chatIDStr).Val())

	if u.update.Message.Chat.ID < 0 {
		msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
			"群组订阅功能已取消，需要订阅功能的话，请私聊奴家呢o(￣ˇ￣)o")
		u.bot.SendMessage(msg)
		return
	}

	if !u.IsAuthed() {
		u.SendQuestion()
		return
	}

	if isSubscribe {
		msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
			"已经订阅过，就不要重复订阅啦😘")
		u.bot.SendMessage(msg)
	} else {
		u.redis.HSet("tgSubscribe", chatIDStr, strconv.FormatBool(true))
		u.redis.HIncrBy("tgSubscribeTimes", chatIDStr, 1)
		msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
			"订阅成功\n以后奴家知道新的群组的话，会第一时间告诉你哟😊\n"+
				"(订阅仅对当前会话有效)")
		u.bot.SendMessage(msg)
	}
}

func (u *Updater) UnSubscribe() {
	chatIDStr := strconv.Itoa(u.update.Message.Chat.ID)
	var msg tgbotapi.MessageConfig
	if u.redis.HExists("tgSubscribe", chatIDStr).Val() {
		u.redis.HDel("tgSubscribe", chatIDStr)
		times, _ := u.redis.HIncrBy("tgSubscribeTimes", chatIDStr, 1).Result()
		if times > 5 {
			msg = tgbotapi.NewMessage(u.update.Message.Chat.ID,
				"订了退，退了订，你烦不烦嘛！！！⊂彡☆))∀`)`")
			u.redis.HDel("tgSubscribeTimes", chatIDStr)
		} else {
			msg = tgbotapi.NewMessage(u.update.Message.Chat.ID,
				"好伤心，退订了就不能愉快的玩耍了呢😭")
		}
	} else {
		msg = tgbotapi.NewMessage(u.update.Message.Chat.ID,
			"你都还没订阅，让人家怎么退订嘛！o(≧口≦)o")
	}
	u.bot.SendMessage(msg)
}

func (u *Updater) PreBroadcast() {
	if u.IsMaster() && u.update.Message.Chat.ID > 0 {
		u.BotReply("Send me the Broadcast (＾o＾)ﾉ")
		u.SetStatus("broadcast")
	}
}

func (u *Updater) Broadcast(msgText string) {
	if u.IsMaster() &&
		u.redis.Exists("tgSubscribe").Val() {

		subStates := u.redis.HGetAllMap("tgSubscribe").Val()

		for k, v := range subStates {
			chatid, _ := strconv.Atoi(k)
			subState, _ := strconv.ParseBool(v)

			if subState && chatid > 0 {
				log.Printf("sending boardcast to %d ...", chatid)
				msg := tgbotapi.NewMessage(chatid, msgText)
				go func(k string) {
					u.bot.SendMessage(msg)
					log.Printf("%s --- done", k)
				}(k)
			}
		}
	}
}
