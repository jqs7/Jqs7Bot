package bb

import "github.com/Syfaro/telegram-bot-api"

type video struct {
	Err    error
	bot    *tgbotapi.BotAPI
	config tgbotapi.VideoConfig
	Ret    tgbotapi.Message
}

func (b *Base) NewVideoShare(chatID int, fileID string) *video {
	return &video{
		bot:    b.Bot,
		config: tgbotapi.NewVideoShare(chatID, fileID),
	}
}

func (b *Base) NewVideoUploadv(chatID int, file interface{}) *video {
	return &video{
		bot:    b.Bot,
		config: tgbotapi.NewVideoUpload(chatID, file),
	}
}

func (v *video) FilePath(path string) *video {
	v.config.FilePath = path
	return v
}

func (v *video) Duration(duration int) *video {
	v.config.Duration = duration
	return v
}

func (v *video) Caption(caption string) *video {
	v.config.Caption = caption
	return v
}

func (v *video) ReplyMarkup(markup interface{}) *video {
	v.config.ReplyMarkup = markup
	return v
}

func (v *video) ReplyToMessageID(id int) *video {
	v.config.ReplyToMessageID = id
	return v
}

func (v *video) Send() *video {
	msg, err := v.bot.SendVideo(v.config)
	v.Ret = msg
	v.Err = err
	return v
}
