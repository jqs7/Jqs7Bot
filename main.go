package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/redis.v3"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/fatih/set"
	"github.com/kylelemons/go-gypsy/yaml"
)

func main() {
	rc := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rc.Close()

	// Init categories
	categories := []string{
		"Linux", "Programming", "Software",
		"影音", "科幻", "ACG", "IT", "社区",
		"闲聊", "资源", "同城", "Others",
	}
	categoriesSet := set.New(set.NonThreadSafe)
	for _, v := range categories {
		categoriesSet.Add(v)
	}

	conf, err := yaml.ReadFile("botconf.yaml")
	if err != nil {
		log.Panic(err)
	}

	botapi, _ := conf.Get("botapi")
	baiduAPI, _ := conf.Get("baiduTransKey")
	bot, err := tgbotapi.NewBotAPI(botapi)
	if err != nil {
		log.Panic(err)
	}

	botname := bot.Self.UserName

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.UpdatesChan(u)

	tips := VimTipsChan(100)

	for update := range updates {

		// Ignore Outdated Updates
		if time.Since(time.Unix(int64(update.Message.Date), 0)) > time.Hour {
			continue
		}

		u := Updater{
			redis:  rc,
			bot:    bot,
			update: update,
			conf:   conf,
		}

		// Logger
		startWithSlash, _ := regexp.MatchString("^/", update.Message.Text)
		atBot, _ := regexp.MatchString("@"+botname, update.Message.Text)
		if update.Message.Chat.ID > 0 || startWithSlash || atBot {
			log.Printf("[%d](%s) -- [%s] -- %s",
				update.Message.Chat.ID, update.Message.Chat.Title,
				update.Message.From.UserName, update.Message.Text,
			)
		}

		// Auto Rule When New Member Jion Group
		if update.Message.NewChatParticipant.ID != 0 {
			chatIDStr := strconv.Itoa(u.update.Message.Chat.ID)
			if u.redis.Exists("tgGroupAutoRule:" + chatIDStr).Val() {
				go u.Rule()
			}
		}

		switch update.Message.Text {

		case "/help", "/start", "/help@" + botname, "/start@" + botname:
			go u.Start()

		case "/rules", "/rules@" + botname:
			go u.Rule()

		case "/about", "/about@" + botname:
			go u.BotReply(YamlList2String(conf, "about"))

		case "/other_resources", "/other_resources@" + botname:
			go u.BotReply(YamlList2String(conf, "其他资源"))

		case "/subscribe", "/subscribe@" + botname:
			go u.Subscribe()

		case "/unsubscribe", "/unsubscribe@" + botname:
			go u.UnSubscribe()

		case "/autorule":
			go u.AutoRule()

		case "/groups", "/groups@" + botname:
			go u.Groups(categories, 3, 5)

		case "/vimtips":
			t := <-tips
			go u.BotReply(t.Content + "\n" + t.Comment)

		default:
			s := strings.Split(update.Message.Text, " ")
			if len(s) >= 2 && s[0] == "/broadcast" {
				msg := strings.Join(s[1:], " ")
				go u.Broadcast(msg)
			} else if len(s) >= 2 && s[0] == "/setrule" {
				rule := strings.Join(s[1:], " ")
				go u.SetRule(rule)
			} else if len(s) >= 2 && s[0] == "/auth" {
				answer := strings.Join(s[1:], " ")
				answer = strings.Trim(answer, "[]")
				go u.Auth(answer)
			} else if len(s) >= 2 && s[0] == "/e64" {
				in := strings.Join(s[1:], " ")
				go u.BotReply(E64(in))
			} else if len(s) >= 2 && s[0] == "/d64" {
				in := strings.Join(s[1:], " ")
				go u.BotReply(D64(in))
			} else if len(s) >= 2 && s[0] == "/trans" {
				in := strings.Join(s[1:], " ")
				go u.BotReply(BaiduTranslate(baiduAPI, in))
			} else if update.Message.Chat.ID > 0 &&
				categoriesSet.Has(update.Message.Text) {
				// custom keyboard reply
				go u.BotReply(YamlList2String(conf, update.Message.Text))
			}
		}
	}
}
