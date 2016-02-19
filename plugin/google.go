package plugin

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/franela/goreq"
	"github.com/jqs7/Jqs7Bot/helper"
	"github.com/jqs7/bb"
)

type Google struct{ bb.Base }

func (g *Google) Run() {
	if len(g.Args) >= 2 {
		q := strings.Join(g.Args[1:], " ")
		g.NewMessage(g.ChatID, g.G(q)).
			MarkdownMode().
			DisableWebPagePreview().
			Send()
	}
}

func (g *Google) G(query string) string {
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
			log.Println("Google Timeout!")
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
		buf.WriteString(helper.ToMarkdown(t, u) + "\n")
	}
	return buf.String()
}
