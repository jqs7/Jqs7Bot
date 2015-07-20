package main

import "strconv"

func (u *Updater) ListMan() {
	if u.update.Message.Chat.ID < 0 {
		var result string
		fields := u.redis.HGetAllMap("tgMan:" +
			strconv.Itoa(u.update.Message.Chat.ID)).Val()
		for k := range fields {
			result += k + "\n"
		}
		u.BotReply(result)
	}
}

func (u *Updater) Man(field string) {
	if u.update.Message.Chat.ID < 0 {
		if field == "man" && !u.redis.HExists("tgMan:"+
			strconv.Itoa(u.update.Message.Chat.ID), "man").Val() {
			u.BotReply("你在慢慢个什么鬼啦！(σﾟ∀ﾟ)σ")
			return
		}
		u.BotReply(u.redis.HGet("tgMan:"+
			strconv.Itoa(u.update.Message.Chat.ID), field).Val())
	}
}

func (u *Updater) SetMan(field, value string) {
	if !u.IsAuthed() {
		u.SendQuestion()
		return
	}

	if u.update.Message.Chat.ID < 0 {
		u.redis.HSet("tgMan:"+strconv.Itoa(u.update.Message.Chat.ID),
			field, value)
		u.BotReply(field + ":\n" + value)
	}
}

func (u *Updater) RmMan(fields ...string) {
	for k := range fields {
		u.redis.HDel("tgMan:"+strconv.Itoa(u.update.Message.Chat.ID), fields[k])
	}
	u.ListMan()
}
