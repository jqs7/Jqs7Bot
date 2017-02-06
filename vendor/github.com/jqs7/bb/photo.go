package bb

import "github.com/Syfaro/telegram-bot-api"

type photo struct {
	Err    error
	bot    *tgbotapi.BotAPI
	config tgbotapi.PhotoConfig
	Ret    tgbotapi.Message
}

func (b *Base) NewPhotoShare(chatID int, fileID string) *photo {
	return &photo{
		bot:    b.Bot,
		config: tgbotapi.NewPhotoShare(chatID, fileID),
	}
}

func (b *Base) NewPhotoUpload(chatID int, file interface{}) *photo {
	return &photo{
		config: tgbotapi.NewPhotoUpload(chatID, file),
	}
}

func (p *photo) FilePath(path string) *photo {
	p.config.FilePath = path
	return p
}

func (p *photo) Caption(caption string) *photo {
	p.config.Caption = caption
	return p
}

func (p *photo) ReplyMarkup(markup interface{}) *photo {
	p.config.ReplyMarkup = markup
	return p
}

func (p *photo) ReplyToMessageID(id int) *photo {
	p.config.ReplyToMessageID = id
	return p
}

func (p *photo) UseExistingPhoto() *photo {
	p.config.UseExistingPhoto = true
	return p
}

func (p *photo) Send() *photo {
	msg, err := p.bot.SendPhoto(p.config)
	p.Ret = msg
	p.Err = err
	return p
}

type userPhoto struct {
	Err    error
	bot    *tgbotapi.BotAPI
	config tgbotapi.UserProfilePhotosConfig
	Ret    tgbotapi.UserProfilePhotos
}

func (b *Base) UserProfilePhotos(userID int) *userPhoto {
	return &userPhoto{
		bot:    b.Bot,
		config: tgbotapi.NewUserProfilePhotos(userID),
	}
}

func (u *userPhoto) Limit(limit int) *userPhoto {
	u.config.Limit = limit
	return u
}

func (u *userPhoto) Offset(offset int) *userPhoto {
	u.config.Offset = offset
	return u
}

func (u *userPhoto) Get() *userPhoto {
	photos, err := u.bot.GetUserProfilePhotos(u.config)
	u.Ret = photos
	u.Err = err
	return u
}
