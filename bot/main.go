package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"gopkg.in/redis.v3"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/kylelemons/go-gypsy/yaml"
)

func main() {
	rc := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	conf, err := yaml.ReadFile("botconf.yaml")
	if err != nil {
		log.Panic(err)
	}

	botapi, _ := conf.Get("botapi")

	bot, err := tgbotapi.NewBotAPI(botapi)
	if err != nil {
		log.Panic(err)
	}

	botname := bot.Self.UserName

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.UpdatesChan(u)

	for update := range updates {

		log.Printf("[%d](%s) -- [%s] -- %s",
			update.Message.Chat.ID, update.Message.Chat.Title,
			update.Message.From.UserName, update.Message.Text,
		)

		u := Updater{
			redis:  rc,
			bot:    bot,
			update: update,
			conf:   conf,
		}

		switch update.Message.Text {

		case "/help", "/start", "/help@" + botname, "/start@" + botname:
			go u.SendMessage(YamlList2String(conf, "help"))

		case "/rules", "/rules@" + botname:
			go u.SendMessage(YamlList2String(conf, "rules"))

		case "/about", "/about@" + botname:
			go u.SendMessage(YamlList2String(conf, "about"))

		case "/linux", "/linux@" + botname:
			go u.SendMessage(YamlList2String(conf, "Linux"))

		case "/programming", "/programming@" + botname:
			go u.SendMessage(YamlList2String(conf, "Programming"))

		case "/software", "/software@" + botname:
			go u.SendMessage(YamlList2String(conf, "Software"))

		case "/videos", "/videos@" + botname:
			go u.SendMessage(YamlList2String(conf, "影音"))

		case "/sci_fi", "/sci_fi@" + botname:
			go u.SendMessage(YamlList2String(conf, "科幻"))

		case "/acg", "/acg@" + botname:
			go u.SendMessage(YamlList2String(conf, "ACG"))

		case "/it", "/it@" + botname:
			go u.SendMessage(YamlList2String(conf, "IT"))

		case "/free_chat", "/free_chat@" + botname:
			go u.SendMessage(YamlList2String(conf, "闲聊"))

		case "/resources", "/resources@" + botname:
			go u.SendMessage(YamlList2String(conf, "资源"))

		case "/same_city", "/same_city@" + botname:
			go u.SendMessage(YamlList2String(conf, "同城"))

		case "/others", "/others@" + botname:
			go u.SendMessage(YamlList2String(conf, "Others"))

		case "/other_resources", "/other_resources@" + botname:
			go u.SendMessage(YamlList2String(conf, "其他资源"))

		case "/subscribe", "/subscribe@" + botname:
			go u.Subscribe()

		case "/unsubscribe", "/unsubscribe@" + botname:
			go u.UnSubscribe()

		default:
			s := strings.Split(update.Message.Text, " ")
			if len(s) > 1 && s[0] == "/broadcast" {
				msg := strings.Join(s[1:], " ")
				go u.Broadcast(msg)
			}
		}
	}
}

type Updater struct {
	redis  *redis.Client
	bot    *tgbotapi.BotAPI
	update tgbotapi.Update
	conf   *yaml.File
}

func (u *Updater) SendMessage(msgText string) {
	chatIDStr := strconv.Itoa(u.update.Message.Chat.ID)
	enableGroupLimit, _ := u.conf.GetBool("enableGroupLimit")
	limitInterval, _ := u.conf.Get("limitInterval")
	limitTimes, _ := u.conf.GetInt("limitTimes")

	if enableGroupLimit && u.update.Message.Chat.ID < 0 {
		if u.redis.Exists(chatIDStr).Val() {
			u.redis.Incr(chatIDStr)
			counter, _ := u.redis.Get(chatIDStr).Int64()
			if counter >= limitTimes {
				log.Println("--- " + u.update.Message.Chat.Title + " --- " + "防刷屏 ---")
				msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
					"刷屏是坏孩纸~！\n聪明宝宝是会跟奴家私聊的哟😊\n@"+u.bot.Self.UserName)
				msg.ReplyToMessageID = u.update.Message.MessageID
				u.bot.SendMessage(msg)
				return
			}
		} else {
			expire, _ := time.ParseDuration(limitInterval)
			u.redis.Set(chatIDStr, "0", expire)
		}
	}

	msg := tgbotapi.NewMessage(u.update.Message.Chat.ID, msgText)
	u.bot.SendMessage(msg)
	return
}

func (u *Updater) Subscribe() {
	chatIDStr := strconv.Itoa(u.update.Message.Chat.ID)
	u.redis.HSet("tgSubscribe", chatIDStr, strconv.FormatBool(true))
	msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
		"订阅成功\n以后奴家知道新的群组的话，会第一时间告诉你哟😊\n(订阅仅对当前会话有效)")
	u.bot.SendMessage(msg)
}

func (u *Updater) UnSubscribe() {
	chatIDStr := strconv.Itoa(u.update.Message.Chat.ID)
	//u.redis.HSet("tgSubscribe", chatIDStr, strconv.FormatBool(false))
	u.redis.HDel("tgSubscribe", chatIDStr)
	u.SendMessage("好伤心，退订了就不能愉快的玩耍了呢😭")
}

func (u *Updater) Broadcast(msgText string) {
	master, _ := u.conf.Get("master")
	if u.update.Message.Chat.UserName == master &&
		u.redis.Exists("tgSubscribe").Val() {

		subStates := u.redis.HGetAllMap("tgSubscribe").Val()

		for k, v := range subStates {
			chatid, _ := strconv.Atoi(k)
			subState, _ := strconv.ParseBool(v)

			if subState {
				log.Printf("sending boardcast to %d ...", chatid)
				msg := tgbotapi.NewMessage(chatid, msgText)
				u.bot.SendMessage(msg)
			}
		}
	}
}

func YamlList2String(config *yaml.File, text string) string {
	count, err := config.Count(text)
	if err != nil {
		log.Println(err)
		return ""
	}

	var resultGroup []string
	for i := 0; i < count; i++ {
		v, err := config.Get(text + "[" + strconv.Itoa(i) + "]")
		if err != nil {
			log.Println(err)
			return ""
		}
		resultGroup = append(resultGroup, v)
	}

	result := strings.Join(resultGroup, "\n")
	result = strings.Replace(result, "\\n", "", -1)

	return result
}
