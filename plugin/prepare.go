package plugin

import (
	"strconv"
	"time"

	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/jqs7/bb"
)

type Prepare struct{ bb.Base }

func (p *Prepare) Run() {
	p.autoRule()
	p.analytics()
	if time.Since(p.Message.Time()) > time.Hour {
		panic("out dated message")
	}
}

func (p *Prepare) autoRule() {
	if p.Message.NewChatParticipant.ID != 0 {
		rc := conf.Redis
		chatIDStr := strconv.Itoa(p.ChatID)
		if rc.Exists("tgGroupAutoRule:" + chatIDStr).Val() {
			p.NewMessage(p.Message.NewChatParticipant.ID,
				"欢迎加入 "+p.Message.Chat.Title+"\n 以下是群组规则：").Send()
			if rc.Exists("tgGroupRule:" + chatIDStr).Val() {
				p.NewMessage(
					p.Message.NewChatParticipant.ID,
					rc.Get("tgGroupRule:"+chatIDStr).Val(),
				).Send()
			}
		}
	}
}

func (p *Prepare) analytics() {
	rc := conf.Redis
	day, month := true, false
	key := func(getDay bool) string {
		return "tgAnalytics:" + GetDate(getDay, 0)
	}
	totalKey := func(getDay bool) string {
		return "tgTotalAnalytics:" + GetDate(getDay, 0)
	}

	rc.HSet("tgUsersID", strconv.Itoa(p.Message.From.ID),
		FromUserName(p.Message.From))
	rc.HSet("tgUsersName", FromUserName(p.Message.From),
		strconv.Itoa(p.Message.From.ID))

	switch {
	case rc.TTL(key(day)).Val() < 0:
		rc.Expire(key(day), time.Hour*(24*3+3))
	case rc.TTL(key(month)).Val() < 0:
		rc.Expire(key(month), time.Hour*(24*30*3+3))
	}

	if p.FromGroup || p.FromSuperGroup {
		rc.Incr(totalKey(day))
		rc.ZIncrBy(key(day), 1, strconv.Itoa(p.Message.From.ID))
		rc.Incr(totalKey(month))
		rc.ZIncrBy(key(month), 1, strconv.Itoa(p.Message.From.ID))
	}
}
