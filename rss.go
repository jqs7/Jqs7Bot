package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Syfaro/telegram-bot-api"
	rss "github.com/jteeuwen/go-pkg-rss"
	"github.com/qiniu/iconv"
)

var stopRssLoop = make(map[string]chan bool)

func (p *Processor) rss(command ...string) {
	f := func() {
		if len(p.s) < 2 {
			p.rssList()
			return
		}

		if len(p.s) > 2 {
			if err := newRss(p, p.s[2]); err != nil {
				msg := tgbotapi.NewMessage(p.chatid(), err.Error())
				bot.SendMessage(msg)
			}
			return
		}

		if err := newRss(p); err != nil {
			msg := tgbotapi.NewMessage(p.chatid(), err.Error())
			bot.SendMessage(msg)
		}
	}
	p.hitter(f, command...)
}

func (p *Processor) rmrss(command ...string) {
	f := func() {
		if len(p.s) < 2 {
			return
		}
		rc.Del("tgRssLatest:" + strconv.Itoa(p.chatid()) + ":" + p.s[1])
		stopRssLoop[strconv.Itoa(p.chatid())+":"+p.s[1]] <- true
		rc.SRem("tgRss:"+strconv.Itoa(p.chatid()), p.s[1])
		rc.Del("tgRssInterval:" + strconv.Itoa(p.chatid()) + ":" + p.s[1])
		p.rssList()
	}
	p.hitter(f, command...)
}

func newRss(p *Processor, interval ...string) error {
	feed := rss.New(1, true, rssChan, p.rssItem)
	if err := feed.Fetch(p.s[1], charsetReader); err != nil {
		loge.Warning(err)
		return errors.New("弹药检测失败，请检查后重试")
	}
	rc.SAdd("tgRssChats", strconv.Itoa(p.chatid()))
	rc.SAdd("tgRss:"+strconv.Itoa(p.chatid()), p.s[1])
	if len(interval) > 0 {
		in, err := strconv.Atoi(interval[0])
		if err != nil {
			return errors.New("哔哔！时空坐标参数设置错误！")
		}
		rc.Set("tgRssInterval:"+
			strconv.Itoa(p.chatid())+":"+p.s[1], interval[0], -1)
		loopFeed(feed, p.s[1], p.chatid(), in)
		return nil
	}
	loopFeed(feed, p.s[1], p.chatid(), -1)
	return nil
}

func (p *Processor) rssItem(feed *rss.Feed,
	ch *rss.Channel, newitems []*rss.Item) {
	rssItem(feed, ch, newitems, p.chatid())
}

func (p *Processor) rssList() {
	r := rc.SMembers("tgRss:" + strconv.Itoa(p.chatid())).Val()
	sort.Strings(r)
	s := strings.Join(r, "\n")
	msg := tgbotapi.NewMessage(p.chatid(), s)
	bot.SendMessage(msg)
}

func initRss() {
	chats := rc.SMembers("tgRssChats").Val()
	for k := range chats {
		feeds := rc.SMembers("tgRss:" + chats[k]).Val()
		id, _ := strconv.Atoi(chats[k])
		chat := &chat{id}
		go func(feeds []string) {
			for u := range feeds {
				feed := rss.New(1, true, rssChan, chat.rssItem)
				interval, _ := strconv.Atoi(rc.Get("tgRssInterval:" + chats[k] + ":" + feeds[u]).Val())
				loopFeed(feed, feeds[u], chat.id, interval)
			}
		}(feeds)
	}
}

type chat struct{ id int }

func (c *chat) rssItem(feed *rss.Feed,
	ch *rss.Channel, newitems []*rss.Item) {
	rssItem(feed, ch, newitems, c.id)
}

func rssItem(feed *rss.Feed,
	ch *rss.Channel, newitems []*rss.Item, chatID int) {
	loge.Infof("%d new item(s) in %s\n", len(newitems), feed.Url)
	var buf bytes.Buffer
	for k, item := range newitems {

		sendMsg := func() {
			if buf.String() != "" {
				msg := tgbotapi.NewMessage(chatID,
					"*"+markdownEscape(ch.Title)+"*\n"+
						buf.String())
				msg.DisableWebPagePreview = true
				msg.ParseMode = tgbotapi.ModeMarkdown
				go func(msg tgbotapi.MessageConfig) {
					bot.SendMessage(msg)
				}(msg)
			}
		}

		if item.Links[0].Href == rc.Get("tgRssLatest:"+
			strconv.Itoa(chatID)+":"+feed.Url).Val() {
			sendMsg()
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
			sendMsg()
			buf.Reset()
		}
	}
	rc.Set("tgRssLatest:"+strconv.Itoa(chatID)+":"+feed.Url,
		newitems[0].Links[0].Href, -1)
}

func rssChan(feed *rss.Feed, newchannels []*rss.Channel) {
	loge.Infof("%d new channel(s) in %s\n", len(newchannels), feed.Url)
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

func loopFeed(feed *rss.Feed, url string, chatid int, interval int) {
	go func() {
		if interval < 7 {
			interval = 7
		}
		stopRssLoop[strconv.Itoa(chatid)+":"+url] = make(chan bool)

		firstLoop := true
		retryTimes := 0
		t := time.Tick(time.Minute*time.Duration(interval-1) +
			time.Second*time.Duration(rand.Intn(120)))

	Loop:
		for {
			select {
			case <-stopRssLoop[strconv.Itoa(chatid)+":"+url]:
				break Loop
			case <-t:
				if firstLoop {
					time.Sleep(time.Duration(rand.Intn(interval)) * time.Minute)
					firstLoop = false
				}
				if err := feed.Fetch(url, charsetReader); err != nil {
					if retryTimes > 30 {
						loge.Warningf("Retry in 30 Minutes...[ %s ]", url)
						time.Sleep(time.Minute * 30)
						retryTimes = 0
						continue
					}
					loge.Warningf("failed to fetch rss, "+
						"retry in 3 seconds... [ %s ]", url)
					time.Sleep(time.Second * 3)
					retryTimes++
					continue
				}
			}
		}

	}()
}
