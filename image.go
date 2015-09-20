package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/franela/goreq"
	"github.com/mozillazg/request"
)

func imageLink(photo tgbotapi.PhotoSize) string {
	file, err := bot.GetFile(tgbotapi.FileConfig{photo.FileID})
	if err != nil {
		return "群组娘连接母舰失败，请稍后重试"
	}
	link := file.Link(bot.Token)
	resp, err := goreq.Request{
		Method: "GET",
		Uri:    link,
	}.Do()

	filePath := "/tmp/" + photo.FileID
	f, err := os.Create(filePath)
	if err != nil {
		return "飞船冷却系统遭到严重虫子干扰，这是药丸？"
	}
	io.Copy(f, resp.Body)
	f, err = os.Open(filePath)
	if err != nil {
		return "飞船冷却系统遭到严重虫子干扰，这是药丸？"
	}
	defer f.Close()

	return vim_cn_Uploader(f)
}

func vim_cn_Uploader(f *os.File) string {
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Files = []request.FileField{
		request.FileField{"name", f.Name(), f},
	}
	res, err := req.Post("https://img.vim-cn.com/")
	if err != nil {
		return "嘟嘟！希望号已失联~( ＞﹏＜ )"
	}

	s, err := res.Text()
	if err != nil {
		return "生命转换系统发生了[叮咚~]"
	}
	s = strings.TrimSpace(s)

	return fmt.Sprintf("[img.vim-cn.com/%s](%s)", s[len(s)-9:], s)
}
