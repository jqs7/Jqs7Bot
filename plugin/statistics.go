package plugin

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/redis.v3"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/jqs7/bb"
)

type Rain struct{ bb.Base }

func (r *Rain) Run() {
	msg := tgbotapi.NewMessage(r.ChatID, " ")
	if len(r.Args) >= 2 {
		switch r.Args[1] {
		case "@":
			msg.Text = Statistics("day", true)
		case "m":
			msg.Text = Statistics("month", false)
		case "m@":
			msg.Text = Statistics("month", true)
		case "^":
			msg.Text = Statistics("yesterday", false)
		case "^@":
			msg.Text = Statistics("yesterday", true)
		case "^m":
			msg.Text = Statistics("last_month", false)
		case "^m@":
			msg.Text = Statistics("last_month", true)
		case "me":
			msg.Text = Statistics(FromUserName(r.Message.From), true)
			msg.ParseMode = tgbotapi.ModeMarkdown
			if r.FromGroup || r.FromSuperGroup {
				msg.ReplyToMessageID = r.Message.MessageID
			}
		default:
			name := strings.Join(r.Args[1:], " ")
			msg.Text = Statistics(name, true)
			msg.ParseMode = tgbotapi.ModeMarkdown
			if r.FromGroup || r.FromSuperGroup {
				msg.ReplyToMessageID = r.Message.MessageID
			}
		}
		r.Bot.SendMessage(msg)
	} else {
		if r.Message.ReplyToMessage != nil {
			msg.Text = Statistics(FromUserName(
				r.Message.ReplyToMessage.From), true)
			if r.FromGroup || r.FromSuperGroup {
				msg.ReplyToMessageID = r.Message.ReplyToMessage.MessageID
			}
			msg.ParseMode = tgbotapi.ModeMarkdown
			r.Bot.SendMessage(msg)
		} else {
			msg.Text = Statistics("day", false)
			r.Bot.SendMessage(msg)
		}
	}
}

func Statistics(s string, withAt bool) string {
	rc := conf.Redis
	day, month := true, false
	key := func(getDay bool, offset int) string {
		return "tgAnalytics:" + GetDate(getDay, offset)
	}
	totalKey := func(getDay bool, offset int) string {
		return "tgTotalAnalytics:" + GetDate(getDay, offset)
	}

	report := func(getDay bool, offset int) string {
		//前10个活跃用户
		result := rc.ZRevRangeByScoreWithScores(key(getDay, offset),
			redis.ZRangeByScore{Min: "-inf", Max: "+inf", Count: 10}).Val()
		//发言总量
		totalTmp := rc.Get(totalKey(getDay, offset)).Val()
		total, _ := strconv.ParseFloat(totalTmp, 64)

		//活跃用户数
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

		//输出格式
		s := fmt.Sprintf("%s大水比💦 Total: %.0f / %d\n",
			title, total, count)
		buf.WriteString(s)
		for k := range result {
			score := result[k].Score
			member := fmt.Sprintf("%s", result[k].Member)
			user := rc.HGet("tgUsersID", member).Val()
			if !withAt {
				user = strings.TrimPrefix(user, "@")
			}
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

		s = fmt.Sprintf("平均每人: %.2f \n更多: http://bot.jqs7.com\n",
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
		//指定用户日|月发言量
		userName := s
		userid := rc.HGet("tgUsersName", s).Val()
		if userid == "" {
			return "舰队阵列手册中查无此人呢喵ˋ( ° ▽、°   )"
		}
		dayCount := rc.ZScore(key(day, 0), userid).Val()
		monthCount := rc.ZScore(key(month, 0), userid).Val()

		//所有用户日|月总发言量
		totalTmp := rc.Get(totalKey(day, 0)).Val()
		dayTotal, _ := strconv.ParseFloat(totalTmp, 64)
		totalTmp = rc.Get(totalKey(month, 0)).Val()
		monthTotal, _ := strconv.ParseFloat(totalTmp, 64)

		//日|月排名
		dayRank := rc.ZRevRank(key(day, 0), userid).Val()
		monthRank := rc.ZRevRank(key(month, 0), userid).Val()

		//日|月总活跃人数
		countDay := rc.ZCount(key(day, 0), "-inf", "+inf").Val()
		countMonth := rc.ZCount(key(month, 0), "-inf", "+inf").Val()
		if dayCount == 0 {
			dayRank = countDay + 1
		}
		if monthCount == 0 {
			monthRank = countMonth + 1
		}

		rank := (2.0 / float64(dayRank+1+monthRank+1)) * 100

		//输出格式
		s := fmt.Sprintf("ID: %s\n今日: %.0f / %.2f%% 排名: %d\n"+
			"本月: %.0f / %.2f%% 排名: %d\n"+
			"水值: %.2f%% [更多](http://bot.jqs7.com/user/%s)\n",
			userid, dayCount, dayCount/dayTotal*100, dayRank+1,
			monthCount, monthCount/monthTotal*100, monthRank+1,
			rank, userName,
		)
		if rank > 10 {
			s += "是个十足的大水比喵！💦"
		}
		return s
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

func GetDate(day bool, offset int) string {
	if day {
		t := time.Now().AddDate(0, 0, offset)
		return strconv.Itoa(t.Year()) +
			t.Month().String() + strconv.Itoa(t.Day())
	} else {
		t := time.Now().AddDate(0, offset, 0)
		return strconv.Itoa(t.Year()) +
			t.Month().String()
	}
}
