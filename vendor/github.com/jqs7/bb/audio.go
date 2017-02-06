package bb

import "github.com/Syfaro/telegram-bot-api"

type audio struct {
	Err    error
	bot    *tgbotapi.BotAPI
	config tgbotapi.AudioConfig
	Ret    tgbotapi.Message
}

func (b *Base) NewAudioShare(chatID int, fileID string) *audio {
	return &audio{
		bot:    b.Bot,
		config: tgbotapi.NewAudioShare(chatID, fileID),
	}
}

func (b *Base) NewAudioUpload(chatID int, file interface{}) *audio {
	return &audio{
		config: tgbotapi.NewAudioUpload(chatID, file),
	}
}

func (a *audio) FilePath(path string) *audio {
	a.config.FilePath = path
	return a
}

func (a *audio) Duration(duration int) *audio {
	a.config.Duration = duration
	return a
}

func (a *audio) Performer(performer string) *audio {
	a.config.Performer = performer
	return a
}

func (a *audio) Title(title string) *audio {
	a.config.Title = title
	return a
}

func (a *audio) ReplyMarkup(markup interface{}) *audio {
	a.config.ReplyMarkup = markup
	return a
}

func (a *audio) ReplyToMessageID(id int) *audio {
	a.config.ReplyToMessageID = id
	return a
}

func (a *audio) UseExistingAudio() *audio {
	a.config.UseExistingAudio = true
	return a
}

func (p *audio) Send() *audio {
	msg, err := p.bot.SendAudio(p.config)
	p.Ret = msg
	p.Err = err
	return p
}
