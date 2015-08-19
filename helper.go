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

	"github.com/Syfaro/telegram-bot-api"
	"github.com/fatih/set"
	"github.com/franela/goreq"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/pyk/byten"
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

func GetDate(day bool, offset int) string {
	now := time.Now()
	year := strconv.Itoa(now.Year())
	var month string
	if !day {
		month = (now.Month() + time.Month(offset)).String()
		return year + month
	} else {
		month = now.Month().String()
		day := strconv.Itoa(now.Day() + offset)
		return year + month + day
	}
}

func FromUserName(user tgbotapi.User) string {
	userName := user.UserName
	if userName != "" {
		return "@" + userName
	}
	name := user.FirstName
	lastName := user.LastName
	if lastName != "" {
		name += " " + lastName
	}
	return name
}

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
		buf.WriteString(t + "\n" + u + "\n")
	}
	return buf.String()
}

func humanByte(in ...interface{}) (out []interface{}) {
	for _, v := range in {
		switch v.(type) {
		case int, uint64:
			s := fmt.Sprintf("%d", v)
			i, _ := strconv.ParseInt(s, 10, 64)
			out = append(out, byten.Size(i))
		case float64:
			s := fmt.Sprintf("%.2f", v)
			out = append(out, s)
		default:
			out = append(out, v)
		}
	}
	return out
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
