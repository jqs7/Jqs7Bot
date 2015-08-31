package main

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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
			loger.Warning("Turing Timeout!")
			return "群组娘连接母舰失败，请稍后重试"
		}
	}

	jasonObj, err := jason.NewObjectFromReader(res.Body)
	if err != nil {
		return "群组娘连接母舰失败，请稍后重试"
	}
	errCode, err := jasonObj.GetInt64("code")
	if err != nil {
		return "群组娘连接母舰失败，请稍后重试"
	}
	switch errCode {
	case 100000: //文本类数据
		out, _ := jasonObj.GetString("text")
		isWeather, _ := regexp.MatchString("^.{2,10}:.*,.*-.*°.*;.*$", out)
		if isWeather {
			replacer := strings.NewReplacer(";", "\n", "晴", "☀️", "多云", "☁️")
			out = replacer.Replace(out)
			out = strings.Replace(out, ":", ":\n", 1)
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

func (p *Processor) turing(command ...string) {
	f := func() {
		if len(p.s) == 1 {
			msg := tgbotapi.NewMessage(p.chatid(), "叫奴家是有什么事呢| ω・´)")
			bot.SendMessage(msg)
		} else if len(p.s) >= 2 {
			in := strings.Join(p.s[1:], " ")
			p._turing(in)
		}
	}
	p.hitter(f, command...)
}

func (p *Processor) _turing(text string) {
	if !p.isAuthed() {
		p.sendQuestion()
		return
	}
	msgText := make(chan string)
	chatAction := make(chan bool)
	asGroupMsg := false
	go func() {
		var userid string
		if p.update.Message.IsGroup() &&
			strings.HasPrefix(text, "-") {
			text = strings.TrimPrefix(text, "-")
			asGroupMsg = true
			userid = strconv.Itoa(p.update.Message.Chat.ID)
		} else {
			userid = strconv.Itoa(p.update.Message.From.ID)
		}
		//语言检测，如果不是中文，则使用翻译后的结果
		reZh := regexp.MustCompile(`[\p{Han}]`).
			FindAllString(text, -1)
		if float32(len(reZh))/float32(utf8.RuneCountInString(text)) < 0.4 {
			text = MsTrans(msID, msSecret, text)
		}
		msgText <- TuringBot(turingAPI, userid, text)
	}()

	p.sendTyping(chatAction)
	<-chatAction
	result := <-msgText
	if asGroupMsg {
		result = fmt.Sprintf("- %s", result)
	}

	msg := tgbotapi.NewMessage(p.chatid(), result)
	msg.DisableWebPagePreview = true
	if p.update.Message.IsGroup() {
		msg.ReplyToMessageID = p.update.Message.MessageID
	}
	bot.SendMessage(msg)
	return
}
