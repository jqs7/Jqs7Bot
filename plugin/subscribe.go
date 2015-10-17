package plugin

import (
	"log"
	"strconv"
	"strings"

	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/jqs7/bb"
)

type Subscribe struct{ Default }

func (s *Subscribe) Run() {
	chatIDStr := strconv.Itoa(s.ChatID)
	isSubscribe, _ := strconv.ParseBool(conf.Redis.HGet("tgSubscribe",
		chatIDStr).Val())

	if s.FromChannel {
		if !s.isAuthed() {
			s.sendQuestion()
			return
		}

		if isSubscribe {
			s.NewMessage(s.ChatID,
				"已经订阅过，就不要重复订阅啦😘").Send()
		} else {
			conf.Redis.HSet("tgSubscribe", chatIDStr, strconv.FormatBool(true))
			conf.Redis.HIncrBy("tgSubscribeTimes", chatIDStr, 1)
			s.NewMessage(s.ChatID,
				"订阅成功\n以后奴家知道新的群组的话，会第一时间告诉你哟😊\n"+
					"(订阅仅对当前会话有效)").Send()
		}
	}
}

type UnSubscribe struct{ bb.Base }

func (u *UnSubscribe) Run() {
	if u.FromGroup {
		return
	}
	chatIDStr := strconv.Itoa(u.ChatID)
	rc := conf.Redis
	if rc.HExists("tgSubscribe", chatIDStr).Val() {
		rc.HDel("tgSubscribe", chatIDStr)
		times, _ := rc.HIncrBy("tgSubscribeTimes", chatIDStr, 1).Result()
		if times > 5 {
			u.NewMessage(u.ChatID,
				"订了退，退了订，你烦不烦嘛！！！⊂彡☆))∀`)`").Send()
			rc.HDel("tgSubscribeTimes", chatIDStr)
			return
		}
		u.NewMessage(u.ChatID,
			"好伤心，退订了就不能愉快的玩耍了呢😭").Send()
		return
	}
	u.NewMessage(u.ChatID,
		"你都还没订阅，让人家怎么退订嘛！o(≧口≦)o").Send()
	return
}

type Broadcast struct{ Default }

func (b *Broadcast) Run() {
	if len(b.Args) == 1 && b.isMaster() &&
		!b.FromGroup {
		b.NewMessage(b.ChatID,
			"Send me the Broadcast (＾o＾)ﾉ").Send()
		b.setStatus("broadcast")
		return
	}
	if len(b.Args) >= 2 {
		text := strings.Join(b.Args[1:], " ")
		b.bc(text)
	}
}

func (b *Default) bc(text string) {
	if b.isMaster() &&
		conf.Redis.Exists("tgSubscribe").Val() {
		subStates := conf.Redis.HGetAllMap("tgSubscribe").Val()

		for k, v := range subStates {
			chatid, _ := strconv.Atoi(k)
			subState, _ := strconv.ParseBool(v)

			if subState && chatid > 0 {
				log.Printf("sending boardcast to %d ... \n", chatid)
				go b.NewMessage(chatid, text).Send()
			}
		}
	}
}
