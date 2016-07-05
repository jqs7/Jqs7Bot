package bb

import "github.com/Syfaro/telegram-bot-api"

type action struct {
	Err    error
	bot    *tgbotapi.BotAPI
	config tgbotapi.ChatActionConfig
}

func (b *Base) NewChatAction(chatID int) *action {
	return &action{
		bot:    b.Bot,
		config: tgbotapi.NewChatAction(chatID, ""),
	}
}

func (a *action) Typing() *action {
	a.config.Action = tgbotapi.ChatTyping
	return a
}

func (a *action) RecordAudio() *action {
	a.config.Action = tgbotapi.ChatRecordAudio
	return a
}

func (a *action) RecordVideo() *action {
	a.config.Action = tgbotapi.ChatRecordVideo
	return a
}

func (a *action) UploadAudio() *action {
	a.config.Action = tgbotapi.ChatUploadAudio
	return a
}

func (a *action) UploadDocument() *action {
	a.config.Action = tgbotapi.ChatUploadDocument
	return a
}

func (a *action) UploadPhoto() *action {
	a.config.Action = tgbotapi.ChatUploadPhoto
	return a
}

func (a *action) UploadVideo() *action {
	a.config.Action = tgbotapi.ChatUploadVideo
	return a
}

func (a *action) Send() *action {
	err := a.bot.SendChatAction(a.config)
	a.Err = err
	return a
}
