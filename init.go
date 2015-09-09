package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/papertrail"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/carlescere/scheduler"
	"github.com/fatih/set"
	"github.com/kylelemons/go-gypsy/yaml"
	"gopkg.in/mgo.v2"
	"gopkg.in/redis.v3"
)

var (
	loge          = logrus.New()
	bot           *tgbotapi.BotAPI
	conf          *yaml.File
	rc            *redis.Client
	mc            *mgo.Session
	categories    []string
	vimtips       chan Tips
	hitokoto      chan Hitokoto
	sticker       chan string
	turingAPI     string
	categoriesSet set.Interface
	msID          string
	msSecret      string
)

func init() {
	var err error
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

	loge.Level = logrus.DebugLevel

	LoadConf()
	botapi, _ := conf.Get("botapi")
	vimTipsCache, _ := conf.GetInt("vimTipsCache")
	hitokotoCache, _ := conf.GetInt("hitokotoCache")
	vimtips = new(Tips).GetChan(int(vimTipsCache))
	hitokoto = new(Hitokoto).GetChan(int(hitokotoCache))
	sticker = RandSticker(rc)
	papertrailUrl, _ := conf.Get("papertrailUrl")
	papertrailPort, _ := conf.GetInt("papertrailPort")

	hook, err := logrus_papertrail.NewPapertrailHook(
		papertrailUrl, int(papertrailPort), "nyan")
	if err != nil {
		loge.Println(err)
	} else {
		loge.Hooks.Add(hook)
	}

	bot, err = tgbotapi.NewBotAPI(botapi)
	if err != nil {
		loge.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	bot.UpdatesChan(u)

	initRss()
	MIndex()
	scheduler.Every().Day().At("01:10").Run(dailySave)
	go GinServer()
}

func LoadConf() {
	var err error
	conf, err = yaml.ReadFile("botconf.yaml")
	if err != nil {
		loge.Panic(err)
	}
	turingAPI, _ = conf.Get("turingBotKey")
	msID, _ = conf.Get("msTransId")
	msSecret, _ = conf.Get("msTransSecret")
}
