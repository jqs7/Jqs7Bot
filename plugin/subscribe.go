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
	userIDStr := strconv.Itoa(s.Message.From.ID)
	isSubscribe, _ := strconv.ParseBool(conf.Redis.HGet("tgSubscribe",
		userIDStr).Val())

	if !s.isAuthed() {
		s.sendQuestion()
		return
	}

	if isSubscribe {
		s.NewMessage(s.ChatID,
			"已经订阅过，就不要重复订阅啦😘").Send()
	} else {
		conf.Redis.HSet("tgSubscribe", userIDStr, strconv.FormatBool(true))
		s.NewMessage(s.ChatID, "订阅成功\n以后奴家知道新的群组的话，会第一时间告诉你哟😊").Send()
	}
}

type UnSubscribe struct{ bb.Base }

func (u *UnSubscribe) Run() {
	userIDStr := strconv.Itoa(u.Message.From.ID)
	rc := conf.Redis
	if rc.HExists("tgSubscribe", userIDStr).Val() {
		rc.HDel("tgSubscribe", userIDStr)
		u.NewMessage(u.Message.From.ID,
			"好伤心，退订了就不能愉快的玩耍了呢😭").Send()
		return
	}
	u.NewMessage(u.Message.From.ID,
		"你都还没订阅，让人家怎么退订嘛！o(≧口≦)o").Send()
	return
}

type Broadcast struct{ Default }

func (b *Broadcast) Run() {
	if b.isMaster() {
		if len(b.Args) == 1 && !b.FromGroup {
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
