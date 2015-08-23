package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/antonholmquist/jason"
	"github.com/franela/goreq"
)

func (p *Processor) trans(command ...string) {
	f := func() {
		if p.update.Message.ReplyToMessage != nil &&
			p.update.Message.ReplyToMessage.Text != "" &&
			len(p.s) < 2 {
			in := p.update.Message.ReplyToMessage.Text
			result := p.translator(in)
			msg := tgbotapi.NewMessage(p.chatid(), result)
			bot.SendMessage(msg)
		} else if len(p.s) >= 2 {
			in := strings.Join(p.s[1:], " ")
			result := p.translator(in)
			msg := tgbotapi.NewMessage(p.chatid(), result)
			bot.SendMessage(msg)
		}
	}
	p.hitter(f, command...)
}

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

func YandexTrans(yandexID, in string) string {
	escapedIn := url.QueryEscape(in)
	retry := 0
Req:
	var to string
	from := YandexDetect(ydTransAPI, in)
	if from == "zh" {
		to = "en"
	} else {
		to = "zh"
	}
	res, err := goreq.Request{
		Uri: fmt.Sprintf("https://translate.yandex.net"+
			"/api/v1.5/tr.json/translate?"+
			"key=%s&lang=%s&text=%s",
			yandexID, to, escapedIn),
		Timeout: 17 * time.Second,
	}.Do()
	if err != nil {
		if retry < 2 {
			retry++
			goto Req
		} else {
			log.Println("Translation Timeout!")
			return "群组娘连接母舰失败，请稍后重试"
		}
	}

	jasonObj, _ := jason.NewObjectFromReader(res.Body)
	code, _ := jasonObj.GetInt64("code")
	switch code {
	case 200:
		text, _ := jasonObj.GetStringArray("text")
		return strings.Join(text, "\n")
	case 401: //未授权用户
		return "大概男盆友用错API Key啦，大家快去蛤他！σ`∀´)`"
	case 402: //API被屏蔽
		return "可怜的群组娘被母舰放逐了X﹏X"
	case 403, 404: //请求次数已用完
		return "今天弹药不足，明天再来吧(＃°Д°)"
	case 413: //文本太长
		return "警报！弹药系统过载！请放宽后重试"
	case 422: //文本不可翻译
		return "咦？这是外星语喵？"
	case 501: //不支持的语种
		return "恭喜你触发了母舰的迷之G点"
	default:
		return "发生了理论上不可能出现的错误，你是不是穿越了喵？"
	}
}

func YandexDetect(yandexID, in string) string {
	in = url.QueryEscape(in)
	retry := 0
Req:
	res, err := goreq.Request{
		Uri: fmt.Sprintf("https://translate.yandex.net/api/v1.5/tr.json/detect?"+
			"key=%s&text=%s",
			yandexID, in),
		Timeout: 17 * time.Second,
	}.Do()
	if err != nil {
		if retry < 2 {
			retry++
			goto Req
		} else {
			log.Println("Translation Timeout!")
			return "群组娘连接母舰失败，请稍后重试"
		}
	}

	jasonObj, _ := jason.NewObjectFromReader(res.Body)
	code, _ := jasonObj.GetInt64("code")
	switch code {
	case 200:
		lang, _ := jasonObj.GetString("lang")
		return lang
	case 401: //未授权用户
		return "大概男盆友用错API Key啦，大家快去蛤他！σ`∀´)`"
	case 402:
		return "可怜的群组娘被母舰放逐了X﹏X"
	case 403, 404: //请求次数已用完
		return "今天弹药不足，明天再来吧(＃°Д°)"
	default:
		return "发生了理论上不可能出现的错误，你是不是穿越了喵？"
	}
}

func (p *Processor) translator(in string) string {
	result := make(chan string)
	typingChan := make(chan bool)
	p.sendTyping(typingChan)
	go func() {
		result <- YandexTrans(ydTransAPI, in)
	}()
	<-typingChan
	return <-result
}
