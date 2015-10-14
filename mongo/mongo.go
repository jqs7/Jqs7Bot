package mongo

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/jqs7/Jqs7Bot/plugin"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/redis.v3"
)

var mc *mgo.Session

func MSession() *mgo.Session {
	if mc == nil {
		var err error
		mgoURL := conf.GetItem("mgoUrl")
		mc, err = mgo.Dial(mgoURL)
		if err != nil {
			log.Println(err)
		}
	}
	return mc.Clone()
}

func M(collection string, f func(*mgo.Collection)) {
	s := MSession()
	defer func() {
		s.Close()

	}()
	f(s.DB("tgBot").C(collection))
}

func MIndex() {
	for _, v := range []string{"dailyTotal", "dailyRank", "dailyUsersCount"} {
		M(v, func(c *mgo.Collection) {
			c.EnsureIndex(mgo.Index{
				Key:    []string{"date"},
				Unique: true,
			})
		})
	}
	M("dailyUser", func(c *mgo.Collection) {
		c.EnsureIndex(mgo.Index{
			Key:    []string{"date", "user"},
			Unique: true,
		})
	})
}

type UserRank struct {
	Name    string
	Count   float64
	Percent float64
}

func DailySave() {
	rc := conf.Redis
	t := time.Now().Add(-24 * time.Hour)
	date := time.Date(t.Year(), t.Month(), t.Day(),
		0, 0, 0, 0, t.Location())

	//每日总发言量统计
	go M("dailyTotal", func(c *mgo.Collection) {
		total := rc.Get("tgTotalAnalytics:" + plugin.GetDate(true, -1)).Val()
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
		result := rc.ZRevRangeByScoreWithScores("tgAnalytics:"+plugin.GetDate(true, -1),
			redis.ZRangeByScore{Min: "-inf", Max: "+inf", Count: 10}).Val()
		//发言总量
		totalTmp := rc.Get("tgTotalAnalytics:" + plugin.GetDate(true, -1)).Val()
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
		count := rc.ZCount("tgAnalytics:"+plugin.GetDate(true, -1), "-inf", "+inf").Val()
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
				score := rc.ZScore("tgAnalytics:"+plugin.GetDate(true, -1), v).Val()
				c.Upsert(bson.M{"date": date, "user": v},
					bson.M{"date": date, "user": v, "count": score})
			}
			if cursor == 0 {
				break
			}
		}
	})
}
