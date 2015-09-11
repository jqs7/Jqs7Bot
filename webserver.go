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
	r.LoadHTMLGlob("html/*")
	r.GET("/", func(c *gin.Context) {
		//limit := time.Now().AddDate(0, 0, -100)
		var total []interface{}
		M("dailyTotal", func(c *mgo.Collection) {
			c.Find(nil).All(&total)
			//c.Find(bson.M{
			//"date": bson.M{"$gt": limit}}).
			//Sort("date").All(&total)
		})
		var users []interface{}
		M("dailyUsersCount", func(c *mgo.Collection) {
			c.Find(nil).All(&users)
			//c.Find(bson.M{
			//"date": bson.M{"$gt": limit}}).
			//Sort("date").All(&users)
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
			c.Find(bson.M{"date": date}).One(&result)
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
			c.Find(bson.M{
				"user": userid,
				"date": bson.M{"$gt": limit}}).
				All(&result)
		})
		c.HTML(http.StatusOK, "user.html",
			result)
	})

	ginpprof.Wrapper(r)
	r.Run(":6060")
}
