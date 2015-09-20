package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/franela/goreq"
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
	client := new(http.Client)
	req, err := fileUploadReq("https://img.vim-cn.com/", "name", f)
	resp, err := client.Do(req)
	if err != nil {
		return "嘟嘟！希望号已失联~( ＞﹏＜ )"
	}
	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return "生命转换系统发生了[叮咚~]"
	}
	defer resp.Body.Close()
	s := strings.TrimSpace(body.String())

	return fmt.Sprintf("[img.vim-cn.com/%s](%s)", s[len(s)-9:], s)
}

func fileUploadReq(uri, paramName string, file *os.File) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, file.Name())
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	//for k, v := range params {
	//writer.WriteField(k, v)
	//}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	return request, nil
}
