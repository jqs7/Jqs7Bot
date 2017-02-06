package bb

import "github.com/Syfaro/telegram-bot-api"

type location struct {
	Err    error
	bot    *tgbotapi.BotAPI
	config tgbotapi.LocationConfig
	Ret    tgbotapi.Message
}

func (b *Base) NewLocation(chatID int, latitude float64, longitude float64) *location {
	return &location{
		bot:    b.Bot,
		config: tgbotapi.NewLocation(chatID, latitude, longitude),
	}
}

func (l *location) ReplyMarkup(markup interface{}) *location {
	l.config.ReplyMarkup = markup
	return l
}

func (l *location) ReplyToMessageID(id int) *location {
	l.config.ReplyToMessageID = id
	return l
}

func (l *location) Send() *location {
	msg, err := l.bot.SendLocation(l.config)
	l.Ret = msg
	l.Err = err
	return l
}
