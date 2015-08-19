package main

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"gopkg.in/redis.v3"
)

func (u *Updater) Analytics() {
	day, month := true, false
	key := func(getDay bool) string {
		return "tgAnalytics:" + GetDate(getDay, 0)
	}
	totalKey := func(getDay bool) string {
		return "tgTotalAnalytics:" + GetDate(getDay, 0)
	}

	u.redis.HSet("tgUsersID", strconv.Itoa(u.update.Message.From.ID),
		FromUserName(u.update.Message.From))
	u.redis.HSet("tgUsersName", FromUserName(u.update.Message.From),
		strconv.Itoa(u.update.Message.From.ID))

	switch {
	case u.redis.TTL(key(day)).Val() < 0:
		u.redis.Expire(key(day), time.Hour*26*2)
	case u.redis.TTL(key(month)).Val() < 0:
		u.redis.Expire(key(month), time.Hour*24*63)
	}

	if u.update.Message.IsGroup() {
		u.redis.Incr(totalKey(day))
		u.redis.ZIncrBy(key(day), 1, strconv.Itoa(u.update.Message.From.ID))
		u.redis.Incr(totalKey(month))
		u.redis.ZIncrBy(key(month), 1, strconv.Itoa(u.update.Message.From.ID))
	}
}

func (u *Updater) Statistics(s string) string {
	day, month := true, false
	key := func(getDay bool, offset int) string {
		return "tgAnalytics:" + GetDate(getDay, offset)
	}
	totalKey := func(getDay bool, offset int) string {
		return "tgTotalAnalytics:" + GetDate(getDay, offset)
	}

	report := func(getDay bool, offset int) string {
		result := u.redis.ZRevRangeByScoreWithScores(key(getDay, offset),
			redis.ZRangeByScore{Min: "-inf", Max: "+inf", Count: 10}).Val()

		totalS := u.redis.Get(totalKey(getDay, offset)).Val()
		total, _ := strconv.ParseFloat(totalS, 64)

		count := u.redis.ZCount(key(getDay, offset), "-inf", "+inf").Val()
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
			user := u.redis.HGet("tgUsersID", member).Val()
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
		userid := u.redis.HGet("tgUsersName", s).Val()
		if userid == "" {
			return "舰队阵列手册中查无此人呢喵ˋ( ° ▽、°  )"
		}
		dayCount := u.redis.ZScore(key(day, 0), userid).Val()
		monthCount := u.redis.ZScore(key(month, 0), userid).Val()

		totalTmp := u.redis.Get(totalKey(day, 0)).Val()
		dayTotal, _ := strconv.ParseFloat(totalTmp, 64)

		totalTmp = u.redis.Get(totalKey(month, 0)).Val()
		monthTotal, _ := strconv.ParseFloat(totalTmp, 64)

		dayRank := u.redis.ZRevRank(key(day, 0), userid).Val()
		monthRank := u.redis.ZRevRank(key(month, 0), userid).Val()
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
