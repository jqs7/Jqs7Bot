package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func GinServer() {
	r := gin.Default()
	if runMode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r.LoadHTMLGlob("html/*")
	r.Static("/assets", "./assets")
	r.GET("/", func(c *gin.Context) {
		var total, users []interface{}
		limit := time.Now().AddDate(0, 0, -100)
		M("dailyTotal", func(c *mgo.Collection) {
			c.Find(bson.M{
				"date": bson.M{"$gt": limit},
			}).Sort("date").All(&total)
		})
		M("dailyUsersCount", func(c *mgo.Collection) {
			c.Find(bson.M{
				"date": bson.M{"$gt": limit},
			}).Sort("date").All(&users)
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
			c.Find(
				bson.M{"date": date},
			).One(&result)
		})
		c.JSON(http.StatusOK, result)
	})

	r.GET("/user/:name", func(c *gin.Context) {
		limit := time.Now().AddDate(0, 0, -100)
		s, err := url.QueryUnescape(c.Params.ByName("name"))
		if err != nil {
			return
		}
		var result []interface{}
		userid := rc.HGet("tgUsersName", s).Val()
		M("dailyUser", func(c *mgo.Collection) {
			c.Find(
				bson.M{
					"user": userid,
					"date": bson.M{"$gt": limit},
				},
			).Sort("date").All(&result)
		})
		c.HTML(http.StatusOK, "user.html",
			gin.H{
				"result":   result,
				"userName": s,
			})
	})

	r.GET("/test", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html",
			gin.H{
				"total": []map[string]interface{}{
					{"date": time.Now(), "total": 2024},
					{"date": time.Now(), "total": 3025},
					{"date": time.Now(), "total": 4026},
					{"date": time.Now(), "total": 5027},
				},
				"users": []map[string]interface{}{
					{"date": time.Now(), "userCount": 256},
					{"date": time.Now(), "userCount": 257},
					{"date": time.Now(), "userCount": 258},
					{"date": time.Now(), "userCount": 259},
				},
			})
	})

	r.GET("/test/:date", func(c *gin.Context) {
		c.JSON(http.StatusOK,
			gin.H{
				"date": time.Now(),
				"rank": []map[string]interface{}{
					{"name": time.Now(), "count": 12, "percent": "10"},
					{"name": time.Now(), "count": 22, "percent": "10"},
					{"name": time.Now(), "count": 32, "percent": "10"},
					{"name": time.Now(), "count": 42, "percent": "10"},
					{"name": time.Now(), "count": 52, "percent": "10"},
					{"name": time.Now(), "count": 22, "percent": "10"},
					{"name": time.Now(), "count": 62, "percent": "10"},
					{"name": time.Now(), "count": 72, "percent": "10"},
					{"name": time.Now(), "count": 82, "percent": "10"},
					{"name": time.Now(), "count": 92, "percent": "10"},
				},
			})
	})

	ginpprof.Wrapper(r)
	r.Run(":6060")
}
