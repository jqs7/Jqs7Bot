package main

import (
	"log"
	"strings"
	"time"

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
				log.Println("Fail to get vim-tips , retry ...")
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
				log.Println("Fail to get Hitokoto , retry ...")
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

func (u *Updater) SaveSticker() {
	if u.update.Message.Sticker.FileID != "" &&
		u.update.Message.Chat.Title == "群组娘的后宫" {
		u.redis.SAdd("tgStickers", u.update.Message.Sticker.FileID)
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
