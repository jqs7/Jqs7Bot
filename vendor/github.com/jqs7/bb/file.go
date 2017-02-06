package bb

import "github.com/Syfaro/telegram-bot-api"

func (b *Base) GetLink(fileID string) (string, error) {
	file, err := b.Bot.GetFile(tgbotapi.FileConfig{fileID})
	if err != nil {
		return "", err
	}
	return file.Link(b.Bot.Token), nil
}
