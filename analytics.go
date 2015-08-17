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

	u.redis.HSet("tgUsersID", strconv.Itoa(u.update.Message.From.ID), u.FromUserName())

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
		return ""
	}
}
