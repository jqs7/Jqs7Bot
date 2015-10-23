package plugin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/carlescere/scheduler"
	"github.com/jqs7/Jqs7Bot/conf"
	rss "github.com/jteeuwen/go-pkg-rss"
	"github.com/qiniu/iconv"
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

		if r.isMaster() {
			if len(r.Args) > 2 {
				if err := r.newRss(r.Args[2]); err != nil {
					r.NewMessage(r.ChatID, err.Error()).Send()
				}
				return
			}

			if err := r.newRss(); err != nil {
				r.NewMessage(r.ChatID, err.Error()).Send()
			}
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
	feed := rss.New(1, true, rssChan, r.rssItem)
	if err := feed.Fetch(r.Args[1], charsetReader); err != nil {
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
			NotImmediately().Run(genLoop(feed, r.Args[1]))
		if err != nil {
			log.Println(err.Error())
			return nil
		}
		jobs.NewJob(strconv.Itoa(r.ChatID)+":"+r.Args[1], j)
		return nil
	}
	j, err := scheduler.Every(getInterval(-1)).Seconds().
		NotImmediately().Run(genLoop(feed, r.Args[1]))
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	jobs.NewJob(strconv.Itoa(r.ChatID)+":"+r.Args[1], j)

	return nil
}

func (r *Rss) rssItem(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	rssItem(feed, ch, newitems, r.Bot, r.ChatID)
}

func rssChan(feed *rss.Feed, newchannels []*rss.Channel) {
	//log.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
}

func charsetReader(charset string, r io.Reader) (io.Reader, error) {
	switch charset {
	case "ISO-8859-1", "iso-8859-1":
		return r, nil
	default:
		cd, err := iconv.Open("utf-8", charset)
		if err != nil {
			break
		}
		r := iconv.NewReader(cd, r, 1024)
		return r, nil
	}
	return nil, errors.New("Unsupported character set encoding: " + charset)
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
	).Replace(s)
}

func rssItem(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item, bot *tgbotapi.BotAPI, chatid int) {
	var buf bytes.Buffer
	rc := conf.Redis
	for k, item := range newitems {

		if item.Links[0].Href == rc.Get("tgRssLatest:"+
			strconv.Itoa(chatid)+":"+feed.Url).Val() {
			if buf.String() != "" {
				msg := tgbotapi.NewMessage(chatid,
					"*"+markdownEscape(ch.Title)+"*\n"+buf.String())
				msg.DisableWebPagePreview = true
				msg.ParseMode = tgbotapi.ModeMarkdown
				bot.SendMessage(msg)
			}
			break
		}

		if len(item.Links) == 0 {
			buf.WriteString(item.Title)
		} else {
			for i, link := range item.Links {
				href := link.Href

				//fix uncomplete link
				if u, e := url.Parse(href); e == nil {
					fu, _ := url.Parse(feed.Url)
					if u.Host == "" {
						u.Host = fu.Host
					}
					if u.Scheme == "" {
						u.Scheme = fu.Scheme
					}
					href = u.String()
				}

				if i == 0 {
					var format string
					if strings.ContainsAny(item.Title, "[]()") {
						format = fmt.Sprintf("%s [link](%s) ",
							markdownEscape(item.Title), href)
					} else {
						format = fmt.Sprintf("[%s](%s) ",
							item.Title, href)
					}
					buf.WriteString(format)
					continue
				}
				buf.WriteString(fmt.Sprintf("[link %d](%s) ",
					i, href))
			}
		}

		buf.WriteString("\n")

		itemNumsInMessage := 9
		if (k != 0 && k%itemNumsInMessage == 0) || k == len(newitems)-1 {
			if buf.String() != "" {
				msg := tgbotapi.NewMessage(chatid,
					"*"+markdownEscape(ch.Title)+"*\n"+buf.String())
				msg.DisableWebPagePreview = true
				msg.ParseMode = tgbotapi.ModeMarkdown
				bot.SendMessage(msg)
			}
			buf.Reset()
		}
	}
	rc.Set("tgRssLatest:"+strconv.Itoa(chatid)+":"+feed.Url,
		newitems[0].Links[0].Href, -1)
}

type chat struct {
	id  int
	bot *tgbotapi.BotAPI
}

func (c *chat) rssItem(feed *rss.Feed,
	ch *rss.Channel, newitems []*rss.Item) {
	rssItem(feed, ch, newitems, c.bot, c.id)
}

func InitRss(bot *tgbotapi.BotAPI) {
	rc := conf.Redis
	chats := rc.SMembers("tgRssChats").Val()
	for _, c := range chats {
		feeds := rc.SMembers("tgRss:" + c).Val()
		id, _ := strconv.Atoi(c)
		chat := &chat{id, bot}
		for _, f := range feeds {
			feed := rss.New(1, true, rssChan, chat.rssItem)
			interval, _ := strconv.Atoi(rc.Get("tgRssInterval:" + c + ":" + f).Val())
			j, err := scheduler.Every(getInterval(interval)).Seconds().
				NotImmediately().Run(genLoop(feed, f))
			if err != nil {
				log.Println(strconv.Itoa(chat.id) + ":" + f + " init fail")
				continue
			}
			jobs.NewJob(strconv.Itoa(chat.id)+":"+f, j)
		}
	}
}

func genLoop(feed *rss.Feed, url string) func() {
	return func() {
		if err := feed.Fetch(url, charsetReader); err != nil {
			log.Println(err.Error() + " " + url)
		}
	}
}
