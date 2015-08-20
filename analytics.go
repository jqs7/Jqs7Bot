package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Syfaro/telegram-bot-api"

	"gopkg.in/redis.v3"
)

func (p *Processor) analytics() {
	day, month := true, false
	key := func(getDay bool) string {
		return "tgAnalytics:" + GetDate(getDay, 0)
	}
	totalKey := func(getDay bool) string {
		return "tgTotalAnalytics:" + GetDate(getDay, 0)
	}

	rc.HSet("tgUsersID", strconv.Itoa(p.update.Message.From.ID),
		FromUserName(p.update.Message.From))
	rc.HSet("tgUsersName", FromUserName(p.update.Message.From),
		strconv.Itoa(p.update.Message.From.ID))

	switch {
	case rc.TTL(key(day)).Val() < 0:
		rc.Expire(key(day), time.Hour*26*2)
	case rc.TTL(key(month)).Val() < 0:
		rc.Expire(key(month), time.Hour*24*63)
	}

	if p.update.Message.IsGroup() {
		rc.Incr(totalKey(day))
		rc.ZIncrBy(key(day), 1, strconv.Itoa(p.update.Message.From.ID))
		rc.Incr(totalKey(month))
		rc.ZIncrBy(key(month), 1, strconv.Itoa(p.update.Message.From.ID))
	}
}

func (p *Processor) statistics(command ...string) {
	f := func() {
		msg := tgbotapi.NewMessage(p.chatid(), " ")
		if len(p.s) >= 2 {
			switch p.s[1] {
			case "m":
				msg = tgbotapi.NewMessage(p.chatid(), Statistics("month"))
			case "^":
				msg = tgbotapi.NewMessage(p.chatid(), Statistics("yesterday"))
			case "^m":
				msg = tgbotapi.NewMessage(p.chatid(), Statistics("last_month"))
			default:
				name := strings.Join(p.s[1:], " ")
				msg = tgbotapi.NewMessage(p.chatid(), Statistics(name))
			}
			bot.SendMessage(msg)
		} else {
			if p.update.Message.ReplyToMessage != nil {
				msg = tgbotapi.NewMessage(p.chatid(),
					Statistics(FromUserName(
						p.update.Message.ReplyToMessage.From)),
				)
				bot.SendMessage(msg)
			} else {
				msg = tgbotapi.NewMessage(p.chatid(), Statistics("day"))
				bot.SendMessage(msg)
			}
		}
	}
	p.hitter(f, command...)
}

func Statistics(s string) string {
	day, month := true, false
	key := func(getDay bool, offset int) string {
		return "tgAnalytics:" + GetDate(getDay, offset)
	}
	totalKey := func(getDay bool, offset int) string {
		return "tgTotalAnalytics:" + GetDate(getDay, offset)
	}

	report := func(getDay bool, offset int) string {
		result := rc.ZRevRangeByScoreWithScores(key(getDay, offset),
			redis.ZRangeByScore{Min: "-inf", Max: "+inf", Count: 10}).Val()

		totalS := rc.Get(totalKey(getDay, offset)).Val()
		total, _ := strconv.ParseFloat(totalS, 64)

		count := rc.ZCount(key(getDay, offset), "-inf", "+inf").Val()
		otherUser := total
		var buf bytes.Buffer
		title := GetDate(getDay, offset) + " "
		if getDay && offset == 0 {
			title = "今日"
		}
		if !getDay && offset == 0 {
			title = "本月"
		}
		s := fmt.Sprintf("%s大水比💦 Total: %.0f / %d\n",
			title, total, count)
		buf.WriteString(s)
		for k := range result {
			score := result[k].Score
			member := fmt.Sprintf("%s", result[k].Member)
			user := rc.HGet("tgUsersID", member).Val()
			s := fmt.Sprintf("%s : %.0f / %.2f%%\n",
				user, score, score/total*100)
			buf.WriteString(s)
			otherUser -= score
		}
		if otherUser > 0 {
			s = fmt.Sprintf("其他用户: %.0f / %.2f%%\n",
				otherUser, otherUser/total*100)
			buf.WriteString(s)
		}
		s = fmt.Sprintf("平均每人: %.2f\n",
			total/float64(count))
		buf.WriteString(s)
		return buf.String()
	}

	switch s {
	case "day":
		return report(true, 0)
	case "month":
		return report(false, 0)
	case "yesterday":
		return report(true, -1)
	case "last_month":
		return report(false, -1)
	default:
		userid := rc.HGet("tgUsersName", s).Val()
		if userid == "" {
			return "舰队阵列手册中查无此人呢喵ˋ( ° ▽、°  )"
		}
		dayCount := rc.ZScore(key(day, 0), userid).Val()
		monthCount := rc.ZScore(key(month, 0), userid).Val()

		totalTmp := rc.Get(totalKey(day, 0)).Val()
		dayTotal, _ := strconv.ParseFloat(totalTmp, 64)

		totalTmp = rc.Get(totalKey(month, 0)).Val()
		monthTotal, _ := strconv.ParseFloat(totalTmp, 64)

		dayRank := rc.ZRevRank(key(day, 0), userid).Val()
		monthRank := rc.ZRevRank(key(month, 0), userid).Val()
		s := fmt.Sprintf("ID: %s\n今日: %.0f / %.2f%% 排名: %d\n"+
			"本月: %.0f / %.2f%% 排名: %d\n",
			userid, dayCount, dayCount/dayTotal*100, dayRank+1,
			monthCount, monthCount/monthTotal*100, monthRank+1)
		if dayRank < 10 && monthRank < 10 {
			s += "是个十足的大水比喵！💦"
		} else if monthRank < 10 {
			s += "今天水的不够多呢！💦"
		}
		return s
	}
}
