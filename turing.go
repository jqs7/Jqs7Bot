package main

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/antonholmquist/jason"
	"github.com/franela/goreq"
)

func TuringBot(apiKey, userid, in string) string {
	in = url.QueryEscape(in)
	retry := 0
Req:
	res, err := goreq.Request{
		Uri: fmt.Sprintf("http://www.tuling123.com/openapi/api?"+
			"key=%s&info=%s&userid=%s", apiKey, in, userid),
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

	jasonObj, _ := jason.NewObjectFromReader(res.Body)
	errCode, _ := jasonObj.GetInt64("code")
	switch errCode {
	case 100000: //文本类数据
		out, _ := jasonObj.GetString("text")
		if strings.Contains(in, url.QueryEscape("天气")) ||
			strings.Contains(in, url.QueryEscape("天氣")) {
			log.Println("ok")
			out = strings.Replace(out, ";", "\n", -1)
		}
		out = strings.Replace(out, "<br>", "\n", -1)
		return out
	case 200000: //网址
		url, _ := jasonObj.GetString("url")
		return url
	case 302000: //新闻
		list, _ := jasonObj.GetObjectArray("list")
		var buf bytes.Buffer
		for _, v := range list {
			article, _ := v.GetString("article")
			url, _ := v.GetString("detailurl")
			buf.WriteString(fmt.Sprintf("%s\n%s\n",
				article, url))
		}
		return buf.String()
	case 305000: //列车
		list, _ := jasonObj.GetObjectArray("list")
		var buf bytes.Buffer
		for _, v := range list {
			trainNum, _ := v.GetString("trainnum")
			start, _ := v.GetString("start")
			terminal, _ := v.GetString("terminal")
			startTime, _ := v.GetString("starttime")
			endTime, _ := v.GetString("endtime")

			buf.WriteString(fmt.Sprintf("%s|%s->%s|%s->%s\n",
				trainNum, start, terminal, startTime, endTime))
		}
		return buf.String()
	case 306000: //航班
		list, _ := jasonObj.GetObjectArray("list")
		var buf bytes.Buffer
		for _, v := range list {
			flight, _ := v.GetString("flight")
			startTime, _ := v.GetString("starttime")
			endTime, _ := v.GetString("endtime")

			buf.WriteString(fmt.Sprintf("%s|%s->%s\n",
				flight, startTime, endTime))
		}
		return buf.String()
	case 308000: //菜谱、视频、小说
		list, _ := jasonObj.GetObjectArray("list")
		var buf bytes.Buffer
		for _, v := range list {
			name, _ := v.GetString("name")
			url, _ := v.GetString("detailurl")
			buf.WriteString(fmt.Sprintf("%s\n%s\n",
				name, url))
		}
		return buf.String()
	case 40001: //key长度错误
		return "大概男盆友用错API Key啦，大家快去蛤他！σ`∀´)`"
	case 40002: //请求内容为空
		return "弹药装填系统泄漏，一定不是奴家的锅(╯‵□′)╯"
	case 40003: //key错误或帐号未激活
		return "大概男盆友用错API Key啦，大家快去蛤他！σ`∀´)`"
	case 40004: //请求次数已用完
		return "今天弹药不足，明天再来吧(＃°Д°)"
	case 40005: //暂不支持该功能
		return "恭喜你触发了母舰的迷之G点"
	case 40006: //服务器升级中
		return "母舰升级中..."
	case 40007: //服务器数据格式异常
		return "转换失败，母舰大概是快没油了Orz"
	default:
		return "发生了理论上不可能出现的错误，你是不是穿越了喵？"
	}
}

func (u *Updater) Turing(turingAPI, text string) {
	if !u.IsAuthed() {
		u.SendQuestion()
		return
	}
	typing := tgbotapi.NewChatAction(u.update.Message.Chat.ID, "typing")
	msgText := make(chan string)
	chatAction := make(chan bool)
	go func() {
		msgText <- TuringBot(turingAPI,
			strconv.Itoa(u.update.Message.Chat.ID), text)
	}()
	go func() {
		u.bot.SendChatAction(typing)
		chatAction <- true
	}()
	<-chatAction
	msg := tgbotapi.NewMessage(u.update.Message.Chat.ID, <-msgText)
	msg.DisableWebPagePreview = true
	if u.update.Message.Chat.ID < 0 {
		msg.ReplyToMessageID = u.update.Message.MessageID
	}
	u.bot.SendMessage(msg)
	return
}
