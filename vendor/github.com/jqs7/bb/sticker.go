package bb

import "github.com/Syfaro/telegram-bot-api"

type sticker struct {
	Err    error
	bot    *tgbotapi.BotAPI
	config tgbotapi.StickerConfig
	Ret    tgbotapi.Message
}

func (b *Base) NewStickerShare(chatID int, fileID string) *sticker {
	return &sticker{
		bot:    b.Bot,
		config: tgbotapi.NewStickerShare(chatID, fileID),
	}
}

func (b *Base) NewStickerUpload(chatID int, file interface{}) *sticker {
	return &sticker{
		bot:    b.Bot,
		config: tgbotapi.NewStickerUpload(chatID, file),
	}
}

func (s *sticker) FilePath(path string) *sticker {
	s.config.FilePath = path
	return s
}

func (s *sticker) ReplyMarkup(markup interface{}) *sticker {
	s.config.ReplyMarkup = markup
	return s
}

func (s *sticker) ReplyToMessageID(id int) *sticker {
	s.config.ReplyToMessageID = id
	return s
}

func (s *sticker) Send() *sticker {
	msg, err := s.bot.SendSticker(s.config)
	s.Ret = msg
	s.Err = err
	return s
}
