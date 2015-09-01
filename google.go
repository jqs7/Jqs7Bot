package main

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/franela/goreq"
)

func Google(query string) string {
	query = url.QueryEscape(query)
	retry := 0
Req:
	res, err := goreq.Request{
		Uri: fmt.Sprintf("http://ajax.googleapis.com/"+
			"ajax/services/search/web?v=1.0&rsz=3&q=%s", query),
		Timeout: 17 * time.Second,
	}.Do()
	if err != nil {
		if retry < 2 {
			retry++
			goto Req
		} else {
			loge.Warning("Google Timeout!")
			return "群组娘连接母舰失败，请稍后重试"
		}
	}

	var google struct {
		ResponseData struct {
			Results []struct {
				URL               string
				TitleNoFormatting string
			}
		}
	}

	err = res.Body.FromJsonTo(&google)
	if err != nil {
		return "转换失败，母舰大概是快没油了Orz"
	}

	var buf bytes.Buffer
	for _, item := range google.ResponseData.Results {
		u, _ := url.QueryUnescape(item.URL)
		t, _ := url.QueryUnescape(item.TitleNoFormatting)
		buf.WriteString(t + "\n" + u + "\n")
	}
	return buf.String()
}

func (p *Processor) google(command ...string) {
	f := func() {
		if len(p.s) >= 2 {
			q := strings.Join(p.s[1:], " ")
			msg := tgbotapi.NewMessage(p.chatid(), Google(q))
			msg.DisableWebPagePreview = true
			bot.SendMessage(msg)
		}
	}
	p.hitter(f, command...)
}
