package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Syfaro/telegram-bot-api"
	rss "github.com/jteeuwen/go-pkg-rss"
	"github.com/qiniu/iconv"
)

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
			log.Println(err)
			return
		}
		rc.SAdd("tgRssChats", strconv.Itoa(p.chatid()))
		rc.SAdd("tgRss:"+strconv.Itoa(p.chatid()), p.s[1])
		loopFeed(feed, p.s[1])
	}
	p.hitter(f, command...)
}

func (p *Processor) rmrss(command ...string) {
	f := func() {
		if len(p.s) < 2 {
			return
		}
		rc.SRem("tgRss:"+strconv.Itoa(p.chatid()), p.s[1:]...)
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
		for k := range feeds {
			feed := rss.New(1, true, rssChan, chat.rssItem)
			loopFeed(feed, feeds[k])
		}
	}
}

type chat struct{ id int }

func (c *chat) rssItem(feed *rss.Feed,
	ch *rss.Channel, newitems []*rss.Item) {
	rssItem(feed, ch, newitems, c.id)
}

func rssItem(feed *rss.Feed,
	ch *rss.Channel, newitems []*rss.Item, chatID int) {
	fmt.Printf("%d new item(s) in %s\n", len(newitems), feed.Url)
	var buf bytes.Buffer
	buf.WriteString(ch.Title + "\n")
	for k, v := range newitems {
		if v.Links[0].Href == rc.Get("tgRssLatest:"+
			strconv.Itoa(chatID)+":"+feed.Url).Val() {
			break
		}
		if k < 25 {
			buf.WriteString(v.Title + "\n" + v.Links[0].Href + "\n")
		}
	}
	rc.Set("tgRssLatest:"+strconv.Itoa(chatID)+":"+feed.Url,
		newitems[0].Links[0].Href, -1)
	if buf.String() != ch.Title+"\n" {
		msg := tgbotapi.NewMessage(chatID, buf.String())
		msg.DisableWebPagePreview = true
		bot.SendMessage(msg)
	}
}

func rssChan(feed *rss.Feed, newchannels []*rss.Channel) {
	fmt.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
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

func loopFeed(feed *rss.Feed, url string) {
	go func() {
		for {
			if err := feed.Fetch(url, charsetReader); err != nil {
				log.Println("failed to fetch rss, " +
					"retry in 3 seconds...")
				time.Sleep(time.Second * 3)
				continue
			}
			<-time.After(time.Minute * 5)
		}
	}()
}