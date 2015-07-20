package main

import (
	"log"
	"strconv"

	"github.com/Syfaro/telegram-bot-api"
)

func (u *Updater) SetRule(rule string) {
	if u.update.Message.Chat.ID < 0 {
		if u.IsAuthed() {
			chatIDStr := strconv.Itoa(u.update.Message.Chat.ID)
			log.Printf("setting rule %s to %s", rule, chatIDStr)
			u.redis.Set("tgGroupRule:"+chatIDStr, rule, -1)
			msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
				"新的群组规则Get！✔️\n以下是新的规则：\n\n"+rule)
			u.bot.SendMessage(msg)
		} else {
			u.SendQuestion()
		}
	}
}

func (u *Updater) AutoRule() {
	if u.update.Message.Chat.ID < 0 {
		chatIDStr := strconv.Itoa(u.update.Message.Chat.ID)
		if u.redis.Exists("tgGroupAutoRule:" + chatIDStr).Val() {
			u.redis.Del("tgGroupAutoRule:" + chatIDStr)
			msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
				"AutoRule Disabled!")
			u.bot.SendMessage(msg)
		} else {
			u.redis.Set("tgGroupAutoRule:"+chatIDStr,
				strconv.FormatBool(true), -1)
			msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
				"AutoRule Enabled!")
			u.bot.SendMessage(msg)
		}
	}
}

func (u *Updater) Rule(chatID int) {
	chatIDStr := strconv.Itoa(u.update.Message.Chat.ID)
	if u.redis.Exists("tgGroupRule:" + chatIDStr).Val() {
		msg := tgbotapi.NewMessage(chatID,
			u.redis.Get("tgGroupRule:"+chatIDStr).Val())
		u.bot.SendMessage(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID,
			YamlList2String(u.conf, "rules"))
		u.bot.SendMessage(msg)
	}
}
