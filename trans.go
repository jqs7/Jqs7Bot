package main

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/antonholmquist/jason"
	"github.com/franela/goreq"
	"github.com/st3v/translator/microsoft"
)

func BaiduTranslate(apiKey, in string) (out, from string) {
	in = url.QueryEscape(in)
	retry := 0
Req:
	res, err := goreq.Request{
		Uri: fmt.Sprintf("http://openapi.baidu.com/public/2.0/bmt/translate?"+
			"client_id=%s&q=%s&from=auto&to=auto",
			apiKey, in),
		Timeout: 17 * time.Second,
	}.Do()
	if err != nil {
		if retry < 2 {
			retry++
			goto Req
		} else {
			log.Println("Translation Timeout!")
			return "群组娘连接母舰失败，请稍后重试", ""
		}
	}

	jasonObj, _ := jason.NewObjectFromReader(res.Body)
	from, _ = jasonObj.GetString("from")
	result, err := jasonObj.GetObjectArray("trans_result")
	if err != nil {
		errCode, _ := jasonObj.GetString("error_code")
		switch errCode {
		case "52001": //超时
			return "转换失败，母舰大概是快没油了Orz", ""
		case "52002": //翻译系统错误
			return "母舰崩坏中...", ""
		case "52003": //未授权用户
			return "大概男盆友用错API Key啦，大家快去蛤他！σ`∀´)`", ""
		case "52004": //必填参数为空
			return "弹药装填系统泄漏，一定不是奴家的锅(╯‵□′)╯", ""
		default:
			return "发生了理论上不可能出现的错误，你是不是穿越了喵？", ""
		}
	}

	var outs []string
	for k := range result {
		tmp, _ := result[k].GetString("dst")
		outs = append(outs, tmp)
	}
	out = strings.Join(outs, "\n")
	return out, from
}

func MsTranslate(clientID, clientSecret, text string) (out, from string, err error) {
	t := microsoft.NewTranslator(clientID, clientSecret)
	from, err = t.Detect(text)
	if err != nil {
		return "", "", err
	}
	switch from {
	case "zh-CHS", "zh-CHT":
		out, err = t.Translate(text, from, "en")
		if err != nil {
			return "", from, err
		}
		return
	default:
		out, err = t.Translate(text, from, "zh-CHS")
		if err != nil {
			return "", from, err
		}
		return
	}
}

func (u *Updater) Trans(in string) (out, from string) {
	sp := strings.Split(in, "\n")

	var w sync.WaitGroup
	var buf bytes.Buffer
	w.Add(2)
	go func() {
		typing := tgbotapi.
			NewChatAction(u.update.Message.Chat.ID, "typing")
		u.bot.SendChatAction(typing)
		w.Done()
	}()
	go func() {
		var err error
		for _, s := range sp {
			out, from, err = MsTranslate(u.configs.msID, u.configs.msSecret, s)
			if err != nil {
				out, from = BaiduTranslate(u.configs.baiduAPI, in)
				return
			}
			buf.WriteString(out + "\n")
		}
		w.Done()
	}()
	w.Wait()
	out = buf.String()
	return
}
