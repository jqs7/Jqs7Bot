package main

import (
	"log"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/fatih/set"
	"github.com/kylelemons/go-gypsy/yaml"
	"gopkg.in/redis.v3"
)

var (
	bot           *tgbotapi.BotAPI
	conf          *yaml.File
	rc            *redis.Client
	categories    []string
	vimtips       chan Tips
	hitokoto      chan Hitokoto
	sticker       chan string
	baiduAPI      string
	turingAPI     string
	msID          string
	msSecret      string
	categoriesSet set.Interface
)

func init() {
	rc = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Init categories
	categories = []string{
		"Linux", "Programming", "Software",
		"影音", "科幻", "ACG", "IT", "社区",
		"闲聊", "资源", "同城", "Others",
	}
	categoriesSet = set.New(set.NonThreadSafe)
	for _, v := range categories {
		categoriesSet.Add(v)
	}

	var err error
	conf, err = yaml.ReadFile("botconf.yaml")
	if err != nil {
		log.Panic(err)
	}

	botapi, _ := conf.Get("botapi")
	baiduAPI, _ = conf.Get("baiduTransKey")
	turingAPI, _ = conf.Get("turingBotKey")
	msID, _ = conf.Get("msTransId")
	msSecret, _ = conf.Get("msTransSecret")
	vimTipsCache, _ := conf.GetInt("vimTipsCache")
	hitokotoCache, _ := conf.GetInt("hitokotoCache")
	vimtips = new(Tips).GetChan(int(vimTipsCache))
	hitokoto = new(Hitokoto).GetChan(int(hitokotoCache))
	sticker = RandSticker(rc)

	bot, err = tgbotapi.NewBotAPI(botapi)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	bot.UpdatesChan(u)
}
