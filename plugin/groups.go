package plugin

import (
	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/jqs7/Jqs7Bot/helper"
	"github.com/jqs7/bb"
)

type Groups struct{ bb.Base }

func (g *Groups) Run() {
	if g.FromGroup {
		g.NewMessage(g.ChatID,
			"使用姿势不对呢喵~ ＞▽＜\n本功能只限私聊使用").
			Send()
		return
	}
	category := helper.To2dSlice(conf.Categories, 3, 5)

	g.NewMessage(g.ChatID,
		"你想要查看哪些群组呢😋\n(为保护群组不被外星人攻击，"+
			"请勿将群链接转发到群组中，或者公布到网络上)").
		ReplyMarkup(g.NewReplyKeyboardMarkup(category).
		OneTimeKeyboard().ResizeKeyboard().Done()).
		Send()
}
