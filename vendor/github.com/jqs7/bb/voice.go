package bb

import "github.com/Syfaro/telegram-bot-api"

type voice struct {
	Err    error
	bot    *tgbotapi.BotAPI
	config tgbotapi.VoiceConfig
	Ret    tgbotapi.Message
}

func (b *Base) NewVoiceShare(chatID int, fileID string) *voice {
	return &voice{
		bot:    b.Bot,
		config: tgbotapi.NewVoiceShare(chatID, fileID),
	}
}

func (b *Base) NewVoiceUpload(chatID int, file interface{}) *voice {
	return &voice{
		bot:    b.Bot,
		config: tgbotapi.NewVoiceUpload(chatID, file),
	}
}

func (v *voice) FilePath(path string) *voice {
	v.config.FilePath = path
	return v
}

func (v *voice) Duration(duration int) *voice {
	v.config.Duration = duration
	return v
}

func (v *voice) ReplyMarkup(markup interface{}) *voice {
	v.config.ReplyMarkup = markup
	return v
}

func (v *voice) ReplyToMessageID(id int) *voice {
	v.config.ReplyToMessageID = id
	return v
}

func (v *voice) Send() *voice {
	msg, err := v.bot.SendVoice(v.config)
	v.Ret = msg
	v.Err = err
	return v
}
