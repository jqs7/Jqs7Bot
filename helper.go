package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/antonholmquist/jason"
	"github.com/fatih/set"
	"github.com/franela/goreq"
	"github.com/kylelemons/go-gypsy/yaml"
)

func YamlList2String(config *yaml.File, text string) string {
	resultGroup := YamlList2Slice(config, text)

	result := strings.Join(resultGroup, "\n")
	result = strings.Replace(result, "\\n", "", -1)

	return result
}

func YamlList2Slice(config *yaml.File, text string) []string {
	count, err := config.Count(text)
	if err != nil {
		log.Println(err)
		return nil
	}

	var result []string
	for i := 0; i < count; i++ {
		v, err := config.Get(text + "[" + strconv.Itoa(i) + "]")
		if err != nil {
			log.Println(err)
			return nil
		}
		result = append(result, v)
	}
	return result
}

type Question struct {
	Q string
	A set.Interface
}

func GetQuestions(config *yaml.File, text string) []Question {
	var result []Question
	questions := YamlList2Slice(config, text)

	for _, v := range questions {
		qs := strings.Split(v, "|")
		question := qs[0]
		answers := strings.Split(qs[1], ";")

		s := set.New(set.ThreadSafe)
		for _, v := range answers {
			s.Add(v)
		}
		result = append(result, Question{question, s})
	}
	return result
}

func To2dSlice(in []string, x, y int) [][]string {
	out := [][]string{}
	var begin, end int
	for i := 0; i < y; i++ {
		end += x
		if end >= len(in) {
			out = append(out, in[begin:])
			break
		}
		out = append(out, in[begin:end])
		begin = end
	}
	return out
}

type Tips struct {
	Content string
	Comment string
}

func (t Tips) GetChan(bufferSize int) (out chan Tips) {
	out = make(chan Tips, bufferSize)
	go func() {
		for {
			var tips Tips
			res, err := goreq.Request{
				Uri:     "http://vim-tips.com/random_tips/json",
				Timeout: 777 * time.Millisecond,
			}.Do()
			if err != nil {
				log.Println("Fail to get vim-tips , retry ...")
				continue
			}
			res.Body.FromJsonTo(&tips)
			out <- tips
		}
	}()
	return out
}

func (t Tips) ToString() string {
	return t.Content + "\n" + t.Comment
}

type Hitokoto struct {
	Hitokoto string
	Source   string
}

func (h Hitokoto) GetChan(bufferSize int) (out chan Hitokoto) {
	out = make(chan Hitokoto, bufferSize)
	go func() {
		for {
			var h Hitokoto
			res, err := goreq.Request{
				Uri:     "http://api.hitokoto.us/rand",
				Timeout: 777 * time.Millisecond,
			}.Do()
			if err != nil {
				log.Println("Fail to get Hitokoto , retry ...")
				continue
			}
			res.Body.FromJsonTo(&h)
			out <- h
		}
	}()
	return out
}

func (h Hitokoto) ToString() string {
	if h.Source == "" {
		return h.Hitokoto
	}
	return "「" + strings.Trim(h.Source, "《》") + "」" + "\n" + h.Hitokoto
}

func BaiduTranslate(apiKey, in string) (out string) {
	in = url.QueryEscape(in)
	retry := 0
Req:
	res, err := goreq.Request{
		Uri: fmt.Sprintf("http://openapi.baidu.com/public/2.0/bmt/translate?"+
			"client_id=%s&q=%s&from=auto&to=auto",
			apiKey, in),
		Timeout: 7777 * time.Millisecond,
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
	result, err := jasonObj.GetObjectArray("trans_result")
	if err != nil {
		errCode, _ := jasonObj.GetString("error_code")
		switch errCode {
		case "52001": //超时
			return "转换失败，母舰大概是快没油了Orz"
		case "52002": //翻译系统错误
			return "母舰崩坏中..."
		case "52003": //未授权用户
			return "大概男盆友用错API Key啦，大家快去蛤他！σ`∀´)`"
		case "52004": //必填参数为空
			return "弹药装填系统泄漏，一定不是奴家的锅(╯‵□′)╯"
		default:
			return "发生了理论上不可能出现的错误，你是不是穿越了喵？"
		}
	}

	var outs []string
	for k := range result {
		tmp, _ := result[k].GetString("dst")
		outs = append(outs, tmp)
	}
	out = strings.Join(outs, "\n")
	return out
}

func Google(query string) string {
	query = url.QueryEscape(query)
	retry := 0
Req:
	res, err := goreq.Request{
		Uri:     "http://ajax.googleapis.com/ajax/services/search/web?v=1.0&rsz=3&q=" + query,
		Timeout: 7777 * time.Millisecond,
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
		buf.WriteString(item.TitleNoFormatting + "\n" + u + "\n")
	}
	return buf.String()
}

func TuringBot(apiKey, in string) string {
	in = url.QueryEscape(in)
	retry := 0
Req:
	res, err := goreq.Request{
		Uri: fmt.Sprintf("http://www.tuling123.com/openapi/api?"+
			"key=%s&info=%s", apiKey, in),
		Timeout: 7777 * time.Millisecond,
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
		return out
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

func E64(in string) string {
	return base64.StdEncoding.EncodeToString([]byte(in))
}

func D64(in string) string {
	out, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return "解码系统出现故障，请查看弹药是否填充无误"
	}
	if utf8.Valid(out) {
		return string(out)
	}
	return "解码结果包含不明物体，群组娘已将之上交国家"
}
