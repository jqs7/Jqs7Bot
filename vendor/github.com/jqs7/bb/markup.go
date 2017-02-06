package bb

import "github.com/Syfaro/telegram-bot-api"

type replyKeyboardMarkup struct {
	config tgbotapi.ReplyKeyboardMarkup
}

func (b *Base) NewReplyKeyboardMarkup(keyboard [][]string) *replyKeyboardMarkup {
	return &replyKeyboardMarkup{
		tgbotapi.ReplyKeyboardMarkup{Keyboard: keyboard},
	}
}

func (r *replyKeyboardMarkup) ResizeKeyboard() *replyKeyboardMarkup {
	r.config.ResizeKeyboard = true
	return r
}

func (r *replyKeyboardMarkup) OneTimeKeyboard() *replyKeyboardMarkup {
	r.config.OneTimeKeyboard = true
	return r
}

func (r *replyKeyboardMarkup) Selective() *replyKeyboardMarkup {
	r.config.Selective = true
	return r
}

func (r *replyKeyboardMarkup) Done() tgbotapi.ReplyKeyboardMarkup {
	return r.config
}

type replyKeyboardHide struct {
	config tgbotapi.ReplyKeyboardHide
}

func (b *Base) NewReplyKeyboardHide() *replyKeyboardHide {
	return &replyKeyboardHide{
		tgbotapi.ReplyKeyboardHide{HideKeyboard: true},
	}
}

func (r *replyKeyboardHide) Selective() *replyKeyboardHide {
	r.config.Selective = true
	return r
}

func (r *replyKeyboardHide) Done() tgbotapi.ReplyKeyboardHide {
	return r.config
}

type forceReply struct {
	config tgbotapi.ForceReply
}

func (b *Base) NewForceReply() *forceReply {
	return &forceReply{
		tgbotapi.ForceReply{ForceReply: true},
	}
}

func (r *forceReply) Selective() *forceReply {
	r.config.Selective = true
	return r
}

func (r *forceReply) Done() tgbotapi.ForceReply {
	return r.config
}
