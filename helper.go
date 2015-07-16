package main

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

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

func VimTipsChan(bufferSize int) (out chan Tips) {
	out = make(chan Tips, bufferSize)
	go func() {
		for {
			var tips Tips
			res, err := goreq.Request{
				Uri:     "http://vim-tips.com/random_tips/json",
				Timeout: 7777 * time.Millisecond,
			}.Do()
			if err != nil {
				log.Println("Fail to get vim-tips , retry in 3 seconds")
				time.Sleep(time.Second * 3)
				continue
			}
			res.Body.FromJsonTo(&tips)
			out <- tips
		}
	}()
	return out
}

func BaiduTranslate(apiKey, in string) (out string) {
	in = url.QueryEscape(in)

	res, err := goreq.Request{
		Uri: fmt.Sprintf("http://openapi.baidu.com/public/2.0/bmt/translate?"+
			"client_id=%s&q=%s&from=auto&to=auto",
			apiKey, in),
		Timeout: 7777 * time.Millisecond,
	}.Do()
	if err != nil {
		out = "群组娘连接母舰失败，请稍后重试"
		log.Println("Translation Timeout!")
		return
	}

	jasonObj, _ := jason.NewObjectFromReader(res.Body)
	result, err := jasonObj.GetObjectArray("trans_result")
	if err != nil {
		errCode, _ := jasonObj.GetString("error_code")
		switch errCode {
		case "52001":
			out = "转换失败，母舰大概是快没油了Orz"
		case "52002":
			out = "母舰崩坏中..."
		case "52003":
			out = "大概男盆友用错API Key啦，大家快去蛤他！σ`∀´)`"
		case "52004":
			out = "弹药装填系统泄漏，一定不是奴家的锅(╯‵□′)╯"
		default:
			out = "发生了理论上不可能出现的错误，你是不是穿越了喵？"
		}
		return out
	}

	var outs []string
	for k := range result {
		tmp, _ := result[k].GetString("dst")
		outs = append(outs, tmp)
	}
	out = strings.Join(outs, "\n")
	return out
}
