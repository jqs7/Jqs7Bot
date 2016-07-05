package bb

import "github.com/Syfaro/telegram-bot-api"

type doc struct {
	Err    error
	bot    *tgbotapi.BotAPI
	config tgbotapi.DocumentConfig
	Ret    tgbotapi.Message
}

func (b *Base) NewDocumentShare(chatID int, fileID string) *doc {
	return &doc{
		bot:    b.Bot,
		config: tgbotapi.NewDocumentShare(chatID, fileID),
	}
}

func (b *Base) NewDocumentUpload(chatID int, file interface{}) *doc {
	return &doc{
		config: tgbotapi.NewDocumentUpload(chatID, file),
	}
}

func (d *doc) FilePath(path string) *doc {
	d.config.FilePath = path
	return d
}

func (d *doc) ReplyMarkup(markup interface{}) *doc {
	d.config.ReplyMarkup = markup
	return d
}

func (d *doc) ReplyToMessageID(id int) *doc {
	d.config.ReplyToMessageID = id
	return d
}

func (d *doc) Send() *doc {
	msg, err := d.bot.SendDocument(d.config)
	d.Ret = msg
	d.Err = err
	return d
}
