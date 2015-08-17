package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"gopkg.in/redis.v3"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/antonholmquist/jason"
	"github.com/fatih/set"
	"github.com/franela/goreq"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/pyk/byten"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/st3v/translator/microsoft"
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

func (u *Updater) Analytics() {
	dayKey := "tgAnalytics:" + GetDate(true)
	monthKey := "tgAnalytics:" + GetDate(false)
	dayTotalKey := "tgTotalAnalytics:" + GetDate(true)
	monthTotalKey := "tgTotalAnalytics:" + GetDate(false)

	u.redis.HSet("tgUsersID", strconv.Itoa(u.update.Message.From.ID), u.FromUserName())

	switch {
	case u.redis.TTL(dayKey).Val() < 0:
		u.redis.Expire(dayKey, time.Hour*24*2)
	case u.redis.TTL(monthKey).Val() < 0:
		u.redis.Expire(monthKey, time.Hour*24*60)
	}

	u.redis.Incr(dayTotalKey)
	u.redis.ZIncrBy(dayKey, 1, strconv.Itoa(u.update.Message.From.ID))
	u.redis.Incr(monthTotalKey)
	u.redis.ZIncrBy(monthKey, 1, strconv.Itoa(u.update.Message.From.ID))
}

func (u *Updater) Statistics(s string) string {
	dayKey := "tgAnalytics:" + GetDate(true)
	monthKey := "tgAnalytics:" + GetDate(false)
	dayTotalKey := "tgTotalAnalytics:" + GetDate(true)
	monthTotalKey := "tgTotalAnalytics:" + GetDate(false)
	switch s {
	case "day":
		result := u.redis.ZRevRangeByScoreWithScores(dayKey,
			redis.ZRangeByScore{Min: "-inf", Max: "+inf", Count: 10}).Val()
		totalS := u.redis.Get(dayTotalKey).Val()
		total, _ := strconv.ParseFloat(totalS, 64)
		otherCount := u.redis.ZCount(dayTotalKey, "-inf", "+inf").Val() - 10
		otherUser := total
		var buf bytes.Buffer
		s := fmt.Sprintf("今日大水比💦Total: %.0f\n", total)
		buf.WriteString(s)
		for k := range result {
			score := result[k].Score
			member := fmt.Sprintf("%s", result[k].Member)
			user := u.redis.HGet("tgUsersID", member).Val()
			s := fmt.Sprintf("%s -- %.0f / %.2f%%\n",
				user, score, score/total*100)
			buf.WriteString(s)
			otherUser -= score
		}
		s = fmt.Sprintf("其他用户:%.0f / %.2f%% 人均:%.0f\n",
			otherUser, otherUser/total*100, otherUser/float64(otherCount))
		buf.WriteString(s)
		return buf.String()
	case "month":
		result := u.redis.ZRevRangeByScoreWithScores(monthKey,
			redis.ZRangeByScore{Min: "-inf", Max: "+inf", Count: 10}).Val()
		totalS := u.redis.Get(monthTotalKey).Val()
		total, _ := strconv.ParseFloat(totalS, 64)
		otherCount := u.redis.ZCount(dayTotalKey, "-inf", "+inf").Val() - 10
		otherUser := total
		var buf bytes.Buffer
		s := fmt.Sprintf("本月大水比:💦Total: %.0f\n", total)
		buf.WriteString(s)
		for k := range result {
			score := result[k].Score
			member := fmt.Sprintf("%s", result[k].Member)
			user := u.redis.HGet("tgUsersID", member).Val()
			s := fmt.Sprintf("%s -- %.0f / %.2f%%\n",
				user, score, score/total*100)
			buf.WriteString(s)
			otherUser -= score
		}
		s = fmt.Sprintf("其他用户:%.0f / %.2f%% 人均:%.0f\n",
			otherUser, otherUser/total*100, otherUser/float64(otherCount))
		buf.WriteString(s)
		return buf.String()
	default:
		return ""
	}
}

func GetDate(day bool) string {
	now := time.Now()
	year := strconv.Itoa(now.Year())
	month := now.Month().String()
	if day {
		day := strconv.Itoa(now.Day())
		return year + month + day
	}
	return year + month
}

func (u *Updater) FromUserName() string {
	userName := u.update.Message.From.UserName
	if userName != "" {
		return "@" + userName
	}
	name := u.update.Message.From.FirstName
	lastName := u.update.Message.From.LastName
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

func Stat(t string) string {
	checkErr := func(err error) string {
		return "系统酱正在食用作死药丸中..."
	}
	switch t {
	case "free":
		m, err := mem.VirtualMemory()
		checkErr(err)
		s, err := mem.SwapMemory()
		checkErr(err)
		mem := new(runtime.MemStats)
		runtime.ReadMemStats(mem)
		return fmt.Sprintf(
			"全局:\n"+
				"Total: %s Free: %s\nUsed: %s %s%%\nCache: %s\n"+
				"Swap:\nTotal: %s Free: %s\n Used: %s %s%%\n"+
				"群组娘:\n"+
				"Allocated: %s\nTotal Allocated: %s\nSystem: %s\n",
			humanByte(m.Total, m.Free, m.Used, m.UsedPercent, m.Cached,
				s.Total, s.Free, s.Used, s.UsedPercent,
				mem.Alloc, mem.TotalAlloc, mem.Sys)...,
		)
	case "df":
		fs, err := disk.DiskPartitions(false)
		checkErr(err)
		var buf bytes.Buffer
		for k := range fs {
			du, err := disk.DiskUsage(fs[k].Mountpoint)
			switch {
			case err != nil, du.UsedPercent == 0, du.Free == 0:
				continue
			}
			f := fmt.Sprintf("Mountpoint: %s Type: %s \n"+
				"Total: %s Free: %s \nUsed: %s %s%%\n",
				humanByte(fs[k].Mountpoint, fs[k].Fstype,
					du.Total, du.Free, du.Used, du.UsedPercent)...,
			)
			buf.WriteString(f)
		}
		return buf.String()
	case "os":
		h, err := host.HostInfo()
		checkErr(err)
		l, err := load.LoadAvg()
		checkErr(err)
		c, err := cpu.CPUPercent(time.Second, false)
		checkErr(err)
		return fmt.Sprintf(
			"OSRelease: %s\nHostName: %s\nLoadAdv: %.2f %.2f %.2f\n"+
				"Goroutine: %d\nCPU: %.2f%%",
			h.Platform, h.Hostname, l.Load1, l.Load5, l.Load15,
			runtime.NumGoroutine(), c[0],
		)
	default:
		return "欢迎来到未知领域(ゝ∀･)"
	}
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
