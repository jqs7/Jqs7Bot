package main

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"gopkg.in/redis.v3"
)

func (u *Updater) Analytics() {
	dayKey := "tgAnalytics:" + GetDate(true)
	monthKey := "tgAnalytics:" + GetDate(false)
	dayTotalKey := "tgTotalAnalytics:" + GetDate(true)
	monthTotalKey := "tgTotalAnalytics:" + GetDate(false)

	u.redis.HSet("tgUsersID", strconv.Itoa(u.update.Message.From.ID),
		FromUserName(u.update.Message.From))
	u.redis.HSet("tgUsersName", FromUserName(u.update.Message.From),
		strconv.Itoa(u.update.Message.From.ID))

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
		count := u.redis.ZCount(monthKey, "-inf", "+inf").Val()
		otherUser := total
		var buf bytes.Buffer
		s := fmt.Sprintf("今日大水比💦 Total: %.0f\n", total)
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
	case "month":
		result := u.redis.ZRevRangeByScoreWithScores(monthKey,
			redis.ZRangeByScore{Min: "-inf", Max: "+inf", Count: 10}).Val()
		totalS := u.redis.Get(monthTotalKey).Val()
		total, _ := strconv.ParseFloat(totalS, 64)
		count := u.redis.ZCount(monthKey, "-inf", "+inf").Val()
		otherUser := total
		var buf bytes.Buffer
		s := fmt.Sprintf("本月大水比:💦 Total: %.0f\n", total)
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
	default:
		userid := u.redis.HGet("tgUsersName", s).Val()
		if userid == "" {
			return "舰队阵列手册中查无此人呢喵ˋ( ° ▽、°  )"
		}
		dayCount := u.redis.ZScore(dayKey, userid).Val()
		monthCount := u.redis.ZScore(monthKey, userid).Val()
		totalTmp := u.redis.Get(dayTotalKey).Val()
		dayTotal, _ := strconv.ParseFloat(totalTmp, 64)
		totalTmp = u.redis.Get(monthTotalKey).Val()
		monthTotal, _ := strconv.ParseFloat(totalTmp, 64)
		dayRank := u.redis.ZRevRank(dayKey, userid).Val()
		monthRank := u.redis.ZRevRank(monthKey, userid).Val()
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
