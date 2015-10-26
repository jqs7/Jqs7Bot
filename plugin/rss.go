package plugin

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/carlescere/scheduler"
	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/m3ng9i/feedreader"
)

type Rss struct{ Default }

var jobs = newJMap()

type jMap struct {
	m map[string]*scheduler.Job
	sync.Mutex
}

func (j *jMap) NewJob(key string, job *scheduler.Job) {
	j.Lock()
	defer j.Unlock()
	if v, ok := j.m[key]; ok {
		v.Quit <- true
		delete(j.m, key)
	}
	j.m[key] = job
}

func (j *jMap) StopJob(key string) {
	j.Lock()
	defer j.Unlock()
	if v, ok := j.m[key]; ok {
		v.Quit <- true
		delete(j.m, key)
	}
}

func (j *jMap) Length() int {
	j.Lock()
	defer j.Unlock()
	return len(j.m)
}

func newJMap() *jMap {
	return &jMap{m: make(map[string]*scheduler.Job)}
}

func (r *Rss) Run() {
	switch r.Args[0] {
	case "/rss":
		if len(r.Args) < 2 {
			r.rssList()
			return
		}

		if len(r.Args) > 2 {
			if err := r.newRss(r.Args[2]); err != nil {
				r.NewMessage(r.ChatID, err.Error()).Send()
			}
			return
		}

		if err := r.newRss(); err != nil {
			r.NewMessage(r.ChatID, err.Error()).Send()
		}

	case "/rmrss":
		if len(r.Args) < 2 {
			return
		}
		rc := conf.Redis
		rc.Del("tgRssLatest:" + strconv.Itoa(r.ChatID) + ":" + r.Args[1])
		jobs.StopJob(strconv.Itoa(r.ChatID) + ":" + r.Args[1])
		rc.SRem("tgRss:"+strconv.Itoa(r.ChatID), r.Args[1])
		rc.Del("tgRssInterval:" + strconv.Itoa(r.ChatID) + ":" + r.Args[1])
		r.rssList()
	}
}

func (r *Rss) rssList() {
	rs := conf.Redis.SMembers("tgRss:" + strconv.Itoa(r.ChatID)).Val()
	if len(rs) > 0 {
		sort.Strings(rs)
		s := strings.Join(rs, "\n")
		r.NewMessage(r.ChatID, s).Send()
	} else {
		r.NewMessage(r.ChatID, "然而此会话并无rss订阅呢ˊ_>ˋ").Send()
	}
}

func (r *Rss) newRss(interval ...string) error {
	rc := conf.Redis
	_, err := feedreader.Fetch(r.Args[1])
	if err != nil {
		log.Println(err)
		return errors.New("弹药检测失败，请检查后重试")
	}
	rc.SAdd("tgRssChats", strconv.Itoa(r.ChatID))
	rc.SAdd("tgRss:"+strconv.Itoa(r.ChatID), r.Args[1])
	if len(interval) > 0 {
		in, err := strconv.Atoi(interval[0])
		if err != nil {
			return errors.New("哔哔！时空坐标参数设置错误！")
		}
		rc.Set("tgRssInterval:"+
			strconv.Itoa(r.ChatID)+":"+r.Args[1], interval[0], -1)
		j, err := scheduler.Every(getInterval(in)).Seconds().
			Run(rssJob(r.Args[1], r.ChatID, r.Bot))
		if err != nil {
			log.Println(err.Error())
			return nil
		}
		jobs.NewJob(strconv.Itoa(r.ChatID)+":"+r.Args[1], j)
		return nil
	}
	j, err := scheduler.Every(getInterval(-1)).Seconds().
		Run(rssJob(r.Args[1], r.ChatID, r.Bot))
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	jobs.NewJob(strconv.Itoa(r.ChatID)+":"+r.Args[1], j)

	return nil
}

func rssJob(feedURL string, chatid int, bot *tgbotapi.BotAPI) func() {
	return func() {
		feed, err := feedreader.Fetch(feedURL)
		if err != nil {
			log.Printf("Error: %s %s", feedURL, err.Error())
			return
		}
		var buf bytes.Buffer
		rc := conf.Redis
		for k, item := range feed.Items {
			if strings.Contains(conf.Redis.Get("tgRssLatest:"+
				strconv.Itoa(chatid)+":"+feedURL).Val(), item.Link) {
				if buf.String() != "" {
					msg := tgbotapi.NewMessage(chatid,
						"*"+markdownEscape(feed.Title)+"*\n"+buf.String())
					msg.DisableWebPagePreview = true
					msg.ParseMode = tgbotapi.ModeMarkdown
					bot.SendMessage(msg)
				}
				break
			}

			if strings.ContainsAny(item.Title, "[]()") ||
				strings.Contains(item.Title, "://") {
				str := fmt.Sprintf("%s [link](%s)\n",
					markdownEscape(item.Title), item.Link)
				buf.WriteString(str)
			} else {
				str := fmt.Sprintf("[%s](%s)\n",
					item.Title, item.Link)
				buf.WriteString(str)
			}

			itemNumsInMessage := 9
			if k != 0 && k%itemNumsInMessage == 0 || k == len(feed.Items)-1 {
				if buf.String() != "" {
					msg := tgbotapi.NewMessage(chatid,
						"*"+markdownEscape(feed.Title)+"*\n"+buf.String())
					msg.DisableWebPagePreview = true
					msg.ParseMode = tgbotapi.ModeMarkdown
					bot.SendMessage(msg)
				}
				buf.Reset()
			}
		}

		if feed.Items != nil && len(feed.Items) > 0 {
			var buf []string
			for _, v := range feed.Items {
				buf = append(buf, v.Link)
			}
			rc.Set("tgRssLatest:"+strconv.Itoa(chatid)+":"+feedURL,
				strings.Join(buf, ":"), -1)
		}
	}
}

//accept minute and return seconds
func getInterval(minute int) (second int) {
	interval := 7
	if minute > interval {
		interval = minute
	}
	return (interval-1)*60 + rand.Intn(120)
}

func markdownEscape(s string) string {
	return strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"`", "\\`",
	).Replace(s)
}

func InitRss(bot *tgbotapi.BotAPI) {
	rc := conf.Redis
	chats := rc.SMembers("tgRssChats").Val()
	for _, c := range chats {
		feeds := rc.SMembers("tgRss:" + c).Val()
		id, _ := strconv.Atoi(c)
		for _, f := range feeds {
			interval, _ := strconv.Atoi(rc.Get("tgRssInterval:" + c + ":" + f).Val())
			j, err := scheduler.Every(getInterval(interval)).Seconds().
				NotImmediately().Run(rssJob(f, id, bot))
			if err != nil {
				log.Println(strconv.Itoa(id) + ":" + f + " init fail")
				continue
			}
			jobs.NewJob(strconv.Itoa(id)+":"+f, j)
		}
	}
	log.Printf("%d jobs init complete!\n", jobs.Length())
}
