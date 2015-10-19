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

	"github.com/Syfaro/telegram-bot-api"
	"github.com/carlescere/scheduler"
	"github.com/jqs7/Jqs7Bot/conf"
	rss "github.com/jteeuwen/go-pkg-rss"
	"github.com/qiniu/iconv"
)

type Rss struct{ Default }

var jobMap map[string]*scheduler.Job = make(map[string]*scheduler.Job)

//var stopRssLoop = newMap()

//type mMap struct {
//m map[string]chan bool
//sync.Mutex
//}

//func (m *mMap) Put(key string) {
//m.Lock()
//defer m.Unlock()
//if v, exist := m.m[key]; exist {
//v <- true
//}
//}

//func (m *mMap) Init(key string) {
//m.Lock()
//defer m.Unlock()
//if _, exist := m.m[key]; !exist {
//m.m[key] = make(chan bool)
//}
//}

//func (m *mMap) Get(key string) chan bool {
//m.Lock()
//defer m.Unlock()
//if v, exist := m.m[key]; exist {
//return v
//}
//return nil
//}

//func newMap() *mMap {
//return &mMap{m: make(map[string]chan bool)}
//}

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
		//stopRssLoop.Put(strconv.Itoa(r.ChatID) + ":" + r.Args[1])
		jobMap[strconv.Itoa(r.ChatID)+":"+r.Args[1]].Quit <- true
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
		jobMap[strconv.Itoa(r.ChatID)+":"+r.Args[1]], err =
			scheduler.Every(getInterval(in)).Seconds().
				NotImmediately().Run(genLoop(feed, r.Args[1]))
		if err != nil {
			log.Println(err.Error)
		}
		return nil
	}
	var err error
	jobMap[strconv.Itoa(r.ChatID)+":"+r.Args[1]], err =
		scheduler.Every(getInterval(-1)).Seconds().
			NotImmediately().Run(genLoop(feed, r.Args[1]))
	if err != nil {
		log.Println(err.Error())
	}

	return nil
}

func (r *Rss) rssItem(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	rssItem(feed, ch, newitems, r.Bot, r.ChatID)
}

func rssChan(feed *rss.Feed, newchannels []*rss.Channel) {
	log.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
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

//func loopFeed(feed *rss.Feed, url string, chatid int, interval int) {
//if interval < 7 {
//interval = 7
//}
//stopRssLoop.Init(strconv.Itoa(chatid) + ":" + url)

//firstLoop := true
//retryTimes := 0
//t := time.Tick(time.Minute*time.Duration(interval-1) +
//time.Second*time.Duration(rand.Intn(120)))

//Loop:
//for {
//select {
//case <-stopRssLoop.Get(strconv.Itoa(chatid) + ":" + url):
//break Loop
//case <-t:
//if firstLoop {
//time.Sleep(time.Duration(rand.Intn(interval)) * time.Minute)
//firstLoop = false
//}
//if err := feed.Fetch(url, charsetReader); err != nil {
//if retryTimes > 30 {
//log.Printf("Retry in 30 Minutes...[ %s ]\n", url)
//time.Sleep(time.Minute * 30)
//retryTimes = 0
//continue
//}
//log.Printf("failed to fetch rss, "+
//"retry in 3 seconds... [ %s ]\n", url)
//time.Sleep(time.Second * 3)
//retryTimes++
//continue
//}
//}
//}
//}

//accept minute and return seconds
func getInterval(minute int) (second int) {
	interval := 7
	if minute > 7 {
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
	log.Printf("%d new item(s) in %s\n", len(newitems), feed.Url)
	var buf bytes.Buffer
	rc := conf.Redis
	for k, item := range newitems {

		if item.Links[0].Href == rc.Get("tgRssLatest:"+
			strconv.Itoa(chatid)+":"+feed.Url).Val() &&
			buf.String() != "" {
			msg := tgbotapi.NewMessage(chatid,
				"*"+markdownEscape(ch.Title)+"*\n"+buf.String())
			msg.DisableWebPagePreview = true
			msg.ParseMode = tgbotapi.ModeMarkdown
			bot.SendMessage(msg)

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
		if (k != 0 && k%itemNumsInMessage == 0) || k == len(newitems)-1 &&
			buf.String() != "" {
			msg := tgbotapi.NewMessage(chatid,
				"*"+markdownEscape(ch.Title)+"*\n"+buf.String())
			msg.DisableWebPagePreview = true
			msg.ParseMode = tgbotapi.ModeMarkdown
			bot.SendMessage(msg)
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
			var err error
			jobMap[strconv.Itoa(chat.id)+":"+f], err =
				scheduler.Every(getInterval(interval)).Seconds().
					NotImmediately().Run(genLoop(feed, f))
			if err != nil {
				log.Println(strconv.Itoa(chat.id) + ":" + f + " init fail")
			}
		}
	}
}

func genLoop(feed *rss.Feed, url string) func() {
	return func() {
		if err := feed.Fetch(url, charsetReader); err != nil {
			log.Println(err.Error() + " " + url)
		}
		log.Println("loop " + url)
	}
}
