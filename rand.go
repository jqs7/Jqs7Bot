package main

import (
	"strings"
	"time"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/franela/goreq"
	"gopkg.in/redis.v3"
)

type Tips struct {
	Content string
	Comment string
}

func (t Tips) GetChan(bufferSize int) (out chan Tips) {
	out = make(chan Tips, bufferSize)
	go func() {
		for {
			var tips Tips
			res, err := goreq.Request{
				Uri:     "http://vim-tips.com/random_tips/json",
				Timeout: 777 * time.Millisecond,
			}.Do()
			if err != nil {
				loge.Warning("Fail to get vim-tips , retry ...")
				continue
			}
			res.Body.FromJsonTo(&tips)
			out <- tips
		}
	}()
	return out
}

func (t Tips) ToString() string {
	return t.Content + "\n" + t.Comment
}

type Hitokoto struct {
	Hitokoto string
	Source   string
}

func (h Hitokoto) GetChan(bufferSize int) (out chan Hitokoto) {
	out = make(chan Hitokoto, bufferSize)
	go func() {
		for {
			var h Hitokoto
			res, err := goreq.Request{
				Uri:     "http://api.hitokoto.us/rand",
				Timeout: 777 * time.Millisecond,
			}.Do()
			if err != nil {
				loge.Warning("Fail to get Hitokoto , retry ...")
				continue
			}
			res.Body.FromJsonTo(&h)
			out <- h
		}
	}()
	return out
}

func (h Hitokoto) ToString() string {
	if h.Source == "" {
		return h.Hitokoto
	}
	return "「" + strings.Trim(h.Source, "《》") + "」" + "\n" + h.Hitokoto
}

func (p *Processor) saveSticker() {
	if p.update.Message.Sticker.FileID != "" &&
		p.isMaster() {
		r, _ := rc.SAdd("tgStickers", p.update.Message.Sticker.FileID).Result()
		if r == 1 {
			msg := tgbotapi.NewMessage(p.chatid(),
				"又学会了一种新的姿势了呢(＾o＾)ﾉ")
			bot.SendMessage(msg)
		}
	}
}

func RandSticker(redis *redis.Client) (out chan string) {
	out = make(chan string, 1)
	go func() {
		for {
			rand := redis.SRandMember("tgStickers").Val()
			out <- rand
		}
	}()
	return out
}

func (p *Processor) rand(command ...string) {
	f := func() {
		chatid := p.chatid()
		if len(p.s) >= 2 {
			switch p.s[1] {
			case "v":
				v := <-vimtips
				msg := tgbotapi.NewMessage(chatid, v.ToString())
				bot.SendMessage(msg)
			case "h":
				h := <-hitokoto
				msg := tgbotapi.NewMessage(chatid, h.ToString())
				bot.SendMessage(msg)
			case "s":
				sid := <-sticker
				s := tgbotapi.
					NewStickerShare(p.update.Message.Chat.ID, sid)
				bot.SendSticker(s)
			}
		} else {
			select {
			case v := <-vimtips:
				msg := tgbotapi.NewMessage(chatid, v.ToString())
				bot.SendMessage(msg)
			case h := <-hitokoto:
				msg := tgbotapi.NewMessage(chatid, h.ToString())
				bot.SendMessage(msg)
			case sid := <-sticker:
				s := tgbotapi.
					NewStickerShare(p.update.Message.Chat.ID, sid)
				bot.SendSticker(s)
			default:
				msg := tgbotapi.NewMessage(chatid,
					"诶诶?群组娘迷路了呢_(:з」∠)_")
				bot.SendMessage(msg)
			}
		}
	}
	p.hitter(f, command...)
}
