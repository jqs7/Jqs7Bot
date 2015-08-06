package main

import (
	"strconv"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/kylelemons/go-gypsy/yaml"
	"gopkg.in/redis.v3"
)

type Updater struct {
	redis  *redis.Client
	bot    *tgbotapi.BotAPI
	update tgbotapi.Update
	conf   *yaml.File
}

func (u *Updater) Start() {
	u.BotReply(YamlList2String(u.conf, "help"))
}

func (u *Updater) IsMaster() bool {
	master, _ := u.conf.Get("master")
	if u.update.Message.Chat.UserName == master {
		return true
	}
	return false
}

func (u *Updater) Groups(categories []string, x, y int) {
	if u.update.Message.Chat.ID < 0 {
		msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
			"使用姿势不对呢喵~ ＞▽＜\n本功能只限私聊使用")
		u.bot.SendMessage(msg)
		return
	}
	category := To2dSlice(categories, x, y)

	msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
		"你想要查看哪些群组呢😋\n(为保护群组不被外星人攻击，"+
			"请勿将群链接转发到群组中，或者公布到网络上)")
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        category,
		OneTimeKeyboard: true,
		ResizeKeyboard:  true,
	}
	u.bot.SendMessage(msg)
}

func (u *Updater) BotReply(msgText string) {
	if !u.IsAuthed() {
		u.SendQuestion()
		return
	}
	msg := tgbotapi.NewMessage(u.update.Message.Chat.ID, msgText)
	u.bot.SendMessage(msg)
	return
}

func (u *Updater) BotReplyNoPreview(msgText string) {
	if !u.IsAuthed() {
		u.SendQuestion()
		return
	}
	msg := tgbotapi.NewMessage(u.update.Message.Chat.ID, msgText)
	msg.DisableWebPagePreview = true
	u.bot.SendMessage(msg)
	return
}

func (u *Updater) SetStatus(status string) {
	if status == "" {
		u.redis.Del("tgStatus:" +
			strconv.Itoa(u.update.Message.Chat.ID))
		return
	} else {
		u.redis.Set("tgStatus:"+
			strconv.Itoa(u.update.Message.Chat.ID), status, -1)
	}
}

func (u *Updater) GetStatus() string {
	if u.redis.Exists("tgStatus:" +
		strconv.Itoa(u.update.Message.Chat.ID)).Val() {
		return u.redis.Get("tgStatus:" +
			strconv.Itoa(u.update.Message.Chat.ID)).Val()
	}
	return ""
}

func (u *Updater) Cancel() {
	if u.update.Message.Chat.ID > 0 {
		msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
			"群组娘已完成零态重置")
		u.SetStatus("")
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardHide{
			HideKeyboard: true,
		}
		u.bot.SendMessage(msg)
	}
}
