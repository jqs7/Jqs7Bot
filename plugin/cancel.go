package plugin

type Cancel struct{ Default }

func (c *Cancel) Run() {
	if !c.FromGroup {
		c.NewMessage(c.ChatID,
			"群组娘已完成零态重置").
			ReplyMarkup(c.NewReplyKeyboardHide().Done()).
			Send()
		c.setStatus("")
	}
}
