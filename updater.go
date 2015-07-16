package main

import (
	"log"
	"strconv"
	"time"

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

func (u *Updater) Groups(categories []string, x, y int) {
	if u.update.Message.Chat.ID < 0 {
		msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
			"使用姿势不对呢喵~ ＞▽＜\n本功能只限私聊使用")
		u.bot.SendMessage(msg)
		return
	}
	category := To2dSlice(categories, x, y)

	msg := tgbotapi.NewMessage(u.update.Message.Chat.ID, "你想要查看哪些群组呢😋")
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        category,
		OneTimeKeyboard: true,
		ResizeKeyboard:  true,
	}
	u.bot.SendMessage(msg)
}

func (u *Updater) SendQuestion() {
	qs := GetQuestions(u.conf, "questions")
	index := time.Now().Hour() % len(qs)
	msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
		"需要通过中文验证之后才能使用本功能哟~\n请问："+
			qs[index].Q+"\n发送 /auth [答案] 给奴家就可以了呢")
	u.bot.SendMessage(msg)
}

func (u *Updater) Auth(answer string) {
	qs := GetQuestions(u.conf, "questions")
	index := time.Now().Hour() % len(qs)
	if qs[index].A.Has(answer) {
		u.redis.SAdd("tgAuthUser", strconv.Itoa(u.update.Message.From.ID))
		log.Printf("%d --- %s Auth OK",
			u.update.Message.From.ID, u.update.Message.From.UserName)
		msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
			"验证成功喵~！\n原来你不是外星人呢😊")
		u.bot.SendMessage(msg)
	} else {
		log.Printf("%d --- %s Auth Fail",
			u.update.Message.From.ID, u.update.Message.From.UserName)
		msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
			"答案不对不对！你一定是外星人！不跟你玩了喵！\n"+
				"重新验证一下吧\n请问："+qs[index].Q)
		u.bot.SendMessage(msg)
	}
}

func (u *Updater) isAuthed() bool {
	if u.redis.SIsMember("tgAuthUser",
		strconv.Itoa(u.update.Message.From.ID)).Val() {
		return true
	}
	return false
}

func (u *Updater) SetRule(rule string) {
	if u.update.Message.Chat.ID < 0 {
		if u.isAuthed() {
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
				"AutoRule Disable!")
			u.bot.SendMessage(msg)
		} else {
			u.redis.Set("tgGroupAutoRule:"+chatIDStr,
				strconv.FormatBool(true), -1)
			msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
				"AutoRule Enable!")
			u.bot.SendMessage(msg)
		}
	}
}

func (u *Updater) Rule() {
	chatIDStr := strconv.Itoa(u.update.Message.Chat.ID)
	if u.redis.Exists("tgGroupRule:" + chatIDStr).Val() {
		msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
			u.redis.Get("tgGroupRule:"+chatIDStr).Val())
		u.bot.SendMessage(msg)
	} else {
		msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
			YamlList2String(u.conf, "rules"))
		u.bot.SendMessage(msg)
	}
}

func (u *Updater) BotReply(msgText string) {
	chatIDStr := strconv.Itoa(u.update.Message.Chat.ID)
	enableGroupLimit, _ := u.conf.GetBool("enableGroupLimit")
	limitInterval, _ := u.conf.Get("limitInterval")
	limitTimes, _ := u.conf.GetInt("limitTimes")

	if !u.isAuthed() {
		u.SendQuestion()
		return
	}

	if enableGroupLimit && u.update.Message.Chat.ID < 0 {
		if u.redis.Exists(chatIDStr).Val() {
			u.redis.Incr(chatIDStr)
			counter, _ := u.redis.Get(chatIDStr).Int64()
			if counter >= limitTimes {
				log.Printf("--- %s --- 防刷屏 ---",
					u.update.Message.Chat.Title)
				msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
					"刷屏是坏孩纸~！\n聪明宝宝是会跟奴家私聊的哟😊\n@"+
						u.bot.Self.UserName)
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
	isSubscribe, _ := strconv.ParseBool(u.redis.HGet("tgSubscribe",
		chatIDStr).Val())

	if u.update.Message.Chat.ID < 0 {
		msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
			"群组订阅功能已取消，需要订阅功能的话，请私聊奴家呢o(￣ˇ￣)o")
		u.bot.SendMessage(msg)
		return
	}

	if u.isAuthed() {
		u.SendQuestion()
		return
	}

	if isSubscribe {
		msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
			"已经订阅过，就不要重复订阅啦😘")
		u.bot.SendMessage(msg)
	} else {
		u.redis.HSet("tgSubscribe", chatIDStr, strconv.FormatBool(true))
		u.redis.HIncrBy("tgSubscribeTimes", chatIDStr, 1)
		msg := tgbotapi.NewMessage(u.update.Message.Chat.ID,
			"订阅成功\n以后奴家知道新的群组的话，会第一时间告诉你哟😊\n"+
				"(订阅仅对当前会话有效)")
		u.bot.SendMessage(msg)
	}
}

func (u *Updater) UnSubscribe() {
	chatIDStr := strconv.Itoa(u.update.Message.Chat.ID)
	var msg tgbotapi.MessageConfig
	if u.redis.HExists("tgSubscribe", chatIDStr).Val() {
		u.redis.HDel("tgSubscribe", chatIDStr)
		times, _ := u.redis.HIncrBy("tgSubscribeTimes", chatIDStr, 1).Result()
		if times > 5 {
			msg = tgbotapi.NewMessage(u.update.Message.Chat.ID,
				"订了退，退了订，你烦不烦嘛！！！⊂彡☆))∀`)`")
			u.redis.HDel("tgSubscribeTimes", chatIDStr)
		} else {
			msg = tgbotapi.NewMessage(u.update.Message.Chat.ID,
				"好伤心，退订了就不能愉快的玩耍了呢😭")
		}
	} else {
		msg = tgbotapi.NewMessage(u.update.Message.Chat.ID,
			"你都还没订阅，让人家怎么退订嘛！o(≧口≦)o")
	}
	u.bot.SendMessage(msg)
}

func (u *Updater) Broadcast(msgText string) {
	master, _ := u.conf.Get("master")
	if u.update.Message.Chat.UserName == master &&
		u.redis.Exists("tgSubscribe").Val() {

		subStates := u.redis.HGetAllMap("tgSubscribe").Val()

		for k, v := range subStates {
			chatid, _ := strconv.Atoi(k)
			subState, _ := strconv.ParseBool(v)

			if subState && chatid > 0 {
				log.Printf("sending boardcast to %d ...", chatid)
				msg := tgbotapi.NewMessage(chatid, msgText)
				go func(k string) {
					u.bot.SendMessage(msg)
					log.Printf("%s --- done", k)
				}(k)
			}
		}
	}
}
