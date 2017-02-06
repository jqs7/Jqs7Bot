package bb

import "github.com/Syfaro/telegram-bot-api"

type message struct {
	Err    error
	bot    *tgbotapi.BotAPI
	config tgbotapi.MessageConfig
	Ret    tgbotapi.Message
}

func (b *Base) NewMessage(chatID int, text string) *message {
	return &message{
		bot:    b.Bot,
		config: tgbotapi.NewMessage(chatID, text),
	}
}

func (m *message) DisableWebPagePreview() *message {
	m.config.DisableWebPagePreview = true
	return m
}

func (m *message) MarkdownMode() *message {
	m.config.ParseMode = tgbotapi.ModeMarkdown
	return m
}

func (m *message) ReplyToMessageID(ID int) *message {
	m.config.ReplyToMessageID = ID
	return m
}

func (m *message) ReplyMarkup(markup interface{}) *message {
	m.config.ReplyMarkup = markup
	return m
}

func (m *message) Send() *message {
	msg, err := m.bot.SendMessage(m.config)
	m.Ret = msg
	m.Err = err
	return m
}

type forward struct {
	Err    error
	bot    *tgbotapi.BotAPI
	config tgbotapi.ForwardConfig
	Ret    tgbotapi.Message
}

func (b *Base) NewForward(chatID, fromChatID, messageID int) *forward {
	return &forward{
		bot:    b.Bot,
		config: tgbotapi.NewForward(chatID, fromChatID, messageID),
	}
}

func (f *forward) Send() *forward {
	msg, err := f.bot.ForwardMessage(f.config)
	f.Ret = msg
	f.Err = err
	return f
}
