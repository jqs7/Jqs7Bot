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
	vimTipsCache, _ := conf.GetInt("vimTipsCache")
	bot, err := tgbotapi.NewBotAPI(botapi)
	if err != nil {
		log.Panic(err)
	}

	botname := bot.Self.UserName

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.UpdatesChan(u)

	tips := VimTipsChan(int(vimTipsCache))

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

		// Auto Rule When New Member Join Group
		if update.Message.NewChatParticipant.ID != 0 {
			chatIDStr := strconv.Itoa(u.update.Message.Chat.ID)
			if u.redis.Exists("tgGroupAutoRule:" + chatIDStr).Val() {
				go func() {
					msg := tgbotapi.NewMessage(update.Message.NewChatParticipant.ID,
						"欢迎加入 "+update.Message.Chat.Title+"\n 以下是群组规则：")
					bot.SendMessage(msg)
					u.Rule(update.Message.NewChatParticipant.ID)
				}()
			}
		}

		s := strings.FieldsFunc(update.Message.Text,
			func(r rune) bool {
				switch r {
				case '\t', '\v', '\f', '\r', ' ', 0xA0:
					return true
				}
				return false
			})

		if len(s) > 0 {
			switch s[0] {
			case "/help", "/start", "/help@" + botname, "/start@" + botname:
				go u.Start()
			case "/rules", "/rules@" + botname:
				go u.Rule(update.Message.Chat.ID)
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
			case "/setrule":
				if len(s) >= 2 {
					rule := strings.Join(s[1:], " ")
					go u.SetRule(rule)
				}
			case "/e64":
				if len(s) >= 2 {
					in := strings.Join(s[1:], " ")
					go u.BotReply(E64(in))
				}
			case "d64":
				if len(s) >= 2 {
					in := strings.Join(s[1:], " ")
					go u.BotReply(D64(in))
				}
			case "/trans":
				if update.Message.ReplyToMessage != nil &&
					update.Message.ReplyToMessage.Text != "" {
					go u.BotReply(BaiduTranslate(baiduAPI,
						update.Message.ReplyToMessage.Text))
				} else if len(s) >= 2 {
					in := strings.Join(s[1:], " ")
					go u.BotReply(BaiduTranslate(baiduAPI, in))
				}
			case "/setman":
				if len(s) >= 3 {
					value := strings.Join(s[2:], " ")
					go u.SetMan(s[1], value)
				}
			case "/rmman":
				if len(s) >= 2 {
					go u.RmMan(s[1:]...)
				}
			case "/man":
				if len(s) == 1 {
					go u.ListMan()
				} else {
					go u.Man(s[1])
				}
			case "/broadcast":
				if len(s) == 1 {
					go u.PreBroadcast()
				} else if len(s) >= 2 {
					msg := strings.Join(s[1:], " ")
					go u.Broadcast(msg)
				}
			default:
				if update.Message.Chat.ID > 0 {
					switch u.GetStatus() {
					case "auth":
						go u.Auth(update.Message.Text)
					case "broadcast":
						go u.Broadcast(update.Message.Text)
					default:
						if categoriesSet.Has(update.Message.Text) {
							// custom keyboard reply
							go u.BotReply(YamlList2String(conf, update.Message.Text))
						}
					}
				}
			}
		}
	}
}
