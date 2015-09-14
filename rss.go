package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
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
		if p.update.Message.IsGroup() {
			return
		}
		if len(p.s) < 2 {
			p.rssList()
			return
		}

		feed := rss.New(1, true, rssChan, p.rssItem)
		if err := feed.Fetch(p.s[1], charsetReader); err != nil {
			msg := tgbotapi.NewMessage(p.chatid(),
				"弹药检测失败，请检查后重试")
			bot.SendMessage(msg)
			loge.Warning(err)
			return
		}
		rc.SAdd("tgRssChats", strconv.Itoa(p.chatid()))
		rc.SAdd("tgRss:"+strconv.Itoa(p.chatid()), p.s[1])
		loopFeed(feed, p.s[1], p.chatid())
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
		p.rssList()
	}
	p.hitter(f, command...)
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
			for k := range feeds {
				feed := rss.New(1, true, rssChan, chat.rssItem)
				loopFeed(feed, feeds[k], chat.id)
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
	for k, v := range newitems {
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
		if v.Links[0].Href == rc.Get("tgRssLatest:"+
			strconv.Itoa(chatID)+":"+feed.Url).Val() {
			sendMsg()
			break
		}

		var format string
		if strings.Contains(v.Title, "[") &&
			strings.Contains(v.Title, "]") {
			format = fmt.Sprintf("%s [link](%s)",
				markdownEscape(v.Title), v.Links[0].Href)
		} else {
			format = fmt.Sprintf("[%s](%s)",
				markdownEscape(v.Title), v.Links[0].Href)
		}
		buf.WriteString(format)
		for k, v := range v.Links {
			if k == 0 {
				continue
			}
			buf.WriteString(
				fmt.Sprintf("[link%d](%s) ", k, v))
		}
		buf.WriteString("\n")

		if (k != 0 && k%8 == 0) || k == len(newitems)-1 {
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

func loopFeed(feed *rss.Feed, url string, chatid int) {
	go func() {
		interval := 7
		stopRssLoop[strconv.Itoa(chatid)+":"+url] = make(chan bool)

		time.Sleep(time.Duration(rand.Intn(interval)) * time.Minute)
		t := time.Tick(time.Minute*time.Duration(interval-1) +
			time.Second*time.Duration(rand.Intn(120)))

	Loop:
		for {
			select {
			case <-stopRssLoop[strconv.Itoa(chatid)+":"+url]:
				break Loop
			case <-t:
				if err := feed.Fetch(url, charsetReader); err != nil {
					loge.Warningf("failed to fetch rss, "+
						"retry in 3 seconds... [ %s ]", url)
					time.Sleep(time.Second * 3)
					continue
				}
			}
		}

	}()
}
