package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/gin-gonic/gin"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
		rc.Expire(key(day), time.Hour*(24*3+3))
	case rc.TTL(key(month)).Val() < 0:
		rc.Expire(key(month), time.Hour*(24*30*3+3))
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
			case "@":
				msg = tgbotapi.NewMessage(p.chatid(), Statistics("day", true))
			case "m":
				msg = tgbotapi.NewMessage(p.chatid(), Statistics("month", false))
			case "m@":
				msg = tgbotapi.NewMessage(p.chatid(), Statistics("month", true))
			case "^":
				msg = tgbotapi.NewMessage(p.chatid(), Statistics("yesterday", false))
			case "^@":
				msg = tgbotapi.NewMessage(p.chatid(), Statistics("yesterday", true))
			case "^m":
				msg = tgbotapi.NewMessage(p.chatid(), Statistics("last_month", false))
			case "^m@":
				msg = tgbotapi.NewMessage(p.chatid(), Statistics("last_month", true))
			case "me":
				msg = tgbotapi.NewMessage(p.chatid(),
					Statistics(FromUserName(p.update.Message.From), true))
				if p.update.Message.IsGroup() {
					msg.ReplyToMessageID = p.update.Message.MessageID
				}
			default:
				name := strings.Join(p.s[1:], " ")
				msg = tgbotapi.NewMessage(p.chatid(), Statistics(name, true))
				if p.update.Message.IsGroup() {
					msg.ReplyToMessageID = p.update.Message.MessageID
				}
			}
			bot.SendMessage(msg)
		} else {
			if p.update.Message.ReplyToMessage != nil {
				msg = tgbotapi.NewMessage(p.chatid(),
					Statistics(FromUserName(
						p.update.Message.ReplyToMessage.From), true),
				)
				bot.SendMessage(msg)
			} else {
				msg = tgbotapi.NewMessage(p.chatid(), Statistics("day", false))
				bot.SendMessage(msg)
			}
		}
	}
	p.hitter(f, command...)
}

func Statistics(s string, withAt bool) string {
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
		//指定用户日|月发言量
		userid := rc.HGet("tgUsersName", s).Val()
		if userid == "" {
			return "舰队阵列手册中查无此人呢喵ˋ( ° ▽、°  )"
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
			"水值: %.2f%%\n",
			userid, dayCount, dayCount/dayTotal*100, dayRank+1,
			monthCount, monthCount/monthTotal*100, monthRank+1,
			rank,
		)
		if rank > 10 {
			s += "是个十足的大水比喵！💦"
		}
		return s
	}
}

type UserRank struct {
	Name    string
	Count   float64
	Percent float64
}

func dailySave() {
	t := time.Now().AddDate(0, 0, -1)
	date := time.Date(t.Year(), t.Month(), t.Day(),
		0, 0, 0, 0, t.Location())

	//每日总发言量统计
	go M("dailyTotal", func(c *mgo.Collection) {
		total := rc.Get("tgTotalAnalytics:" + GetDate(true, -1)).Val()
		if total == "" {
			total = "0"
		}
		c.Upsert(bson.M{"date": date},
			bson.M{
				"date":  date,
				"total": total,
			})
	})

	//每日前10名用户
	go M("dailyRank", func(c *mgo.Collection) {
		//前10个活跃用户
		result := rc.ZRevRangeByScoreWithScores("tgAnalytics:"+GetDate(true, -1),
			redis.ZRangeByScore{Min: "-inf", Max: "+inf", Count: 10}).Val()
		//发言总量
		totalTmp := rc.Get("tgTotalAnalytics:" + GetDate(true, -1)).Val()
		total, _ := strconv.ParseFloat(totalTmp, 64)

		var u []UserRank
		for _, v := range result {
			id := fmt.Sprintf("%s", v.Member)
			name := rc.HGet("tgUsersID", id).Val()
			user := UserRank{
				Name:    name,
				Count:   v.Score,
				Percent: v.Score / total * 100,
			}
			u = append(u, user)
		}
		c.Upsert(bson.M{"date": date},
			bson.M{"date": date, "rank": u})
	})

	//每日活跃用户量
	go M("dailyUsersCount", func(c *mgo.Collection) {
		count := rc.ZCount("tgAnalytics:"+GetDate(true, -1), "-inf", "+inf").Val()
		c.Upsert(bson.M{"date": date},
			bson.M{"date": date, "userCount": count})
	})

	//每个用户每日发言量
	go M("dailyUser", func(c *mgo.Collection) {
		var cursor int64
		for {
			var result []string
			cursor, result = rc.HScan("tgUsersID", cursor, "", 10).Val()
			for _, v := range result {
				score := rc.ZScore("tgAnalytics:"+GetDate(true, -1), v).Val()
				c.Upsert(bson.M{"date": date, "user": v},
					bson.M{"date": date, "user": v, "count": score})
			}
			if cursor == 0 {
				break
			}
		}
	})
}

func MIndex() {
	for _, v := range []string{"dailyTotal", "dailyRank", "dailyUsersCount"} {
		M(v, func(c *mgo.Collection) {
			c.EnsureIndex(mgo.Index{
				Key:    []string{"-date"},
				Unique: true,
			})
		})
	}
	M("dailyUser", func(c *mgo.Collection) {
		c.EnsureIndex(mgo.Index{
			Key:    []string{"-date", "user"},
			Unique: true,
		})
	})
}

func GinServer() {
	r := gin.Default()
	r.LoadHTMLGlob("html/*")
	r.GET("/", func(c *gin.Context) {
		var total []interface{}
		M("dailyTotal", func(c *mgo.Collection) {
			c.Find(nil).All(&total)
		})
		var users []interface{}
		M("dailyUsersCount", func(c *mgo.Collection) {
			c.Find(nil).All(&users)
		})
		c.HTML(http.StatusOK, "index.html",
			gin.H{"total": total, "users": users})
	})

	r.GET("/rank/:date", func(c *gin.Context) {
		s := c.Params.ByName("date")
		loc, _ := time.LoadLocation("Asia/Shanghai")
		date, err := time.ParseInLocation("2006-01-02", s, loc)
		if err != nil {
			return
		}
		var result interface{}
		M("dailyRank", func(c *mgo.Collection) {
			c.Find(gin.H{"date": date}).One(&result)
		})
		c.JSON(http.StatusOK, result)
	})

	r.GET("/user/:name", func(c *gin.Context) {
		s, err := url.QueryUnescape(c.Params.ByName("name"))
		if err != nil {
			return
		}
		var result []interface{}
		userid := rc.HGet("tgUsersName", s).Val()
		M("dailyUser", func(c *mgo.Collection) {
			c.Find(gin.H{"user": userid}).All(&result)
		})
		c.HTML(http.StatusOK, "user.html",
			result)
	})

	ginpprof.Wrapper(r)
	r.Run(":6060")
}
