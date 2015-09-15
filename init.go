package main

import (
	"net/http"

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
	rc     *redis.Client
	mc     *mgo.Session
	mgoURL string

	loge = logrus.New()

	runMode string
	bot     *tgbotapi.BotAPI

	conf          *yaml.File
	categories    []string
	categoriesSet set.Interface

	vimtips  chan Tips
	hitokoto chan Hitokoto
	sticker  chan string

	turingAPI string
	msID      string
	msSecret  string
)

func init() {
	var err error
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
	redisPass, _ := conf.Get("redisPass")
	vimTipsCache, _ := conf.GetInt("vimTipsCache")
	hitokotoCache, _ := conf.GetInt("hitokotoCache")
	vimtips = new(Tips).GetChan(int(vimTipsCache))
	hitokoto = new(Hitokoto).GetChan(int(hitokotoCache))
	papertrailURL, _ := conf.Get("papertrailUrl")
	papertrailPort, _ := conf.GetInt("papertrailPort")

	rc = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: redisPass,
	})

	sticker = RandSticker(rc)

	//logger
	hook, err := logrus_papertrail.NewPapertrailHook(
		papertrailURL, int(papertrailPort), "nyan")
	if err != nil {
		loge.Println(err)
	} else {
		loge.Hooks.Add(hook)
	}

	//bot init
	bot, err = tgbotapi.NewBotAPI(botapi)
	if err != nil {
		loge.Panic(err)
	}

	if runMode == "debug" {
		hook := tgbotapi.NewWebhook("")
		bot.SetWebhook(hook)
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		bot.UpdatesChan(u)
	} else {
		hook := tgbotapi.NewWebhookWithCert("https://jqs7.com:8443/"+bot.Token, "crt.pem")
		bot.SetWebhook(hook)
		bot.ListenForWebhook("/" + bot.Token)
		go http.ListenAndServeTLS(":8443", "crt.pem", "key.pem", nil)
	}

	initRss()
	MIndex()
	dailySave()
	scheduler.Every().Day().At("00:05").Run(dailySave)
	go GinServer()
}

func LoadConf() {
	var err error
	conf, err = yaml.ReadFile("botconf.yaml")
	if err != nil {
		loge.Panic(err)
	}
	runMode, _ = conf.Get("runMode")
	turingAPI, _ = conf.Get("turingBotKey")
	mgoURL, _ = conf.Get("mgoUrl")
	msID, _ = conf.Get("msTransId")
	msSecret, _ = conf.Get("msTransSecret")
}
