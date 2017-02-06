package webServer

import (
	"net/http"
	"net/url"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/jqs7/Jqs7Bot/mongo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/unrolled/render.v1"
)

func GinServer() {
	r := gin.Default()
	if conf.GetItem("runMode") == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	render := render.New(render.Options{
		Directory:  "html",
		Delims:     render.Delims{"<%", "%>"},
		Extensions: []string{".html"},
	})

	r.Static("/dist", "./html/dist")

	r.GET("/", func(c *gin.Context) {
		render.HTML(c.Writer, http.StatusOK, "index", nil)
	})

	r.GET("/api", func(c *gin.Context) {
		var total, users []interface{}
		limit := time.Now().AddDate(0, 0, -100)
		mongo.M("dailyTotal", func(c *mgo.Collection) {
			c.Find(bson.M{
				"date": bson.M{"$gt": limit},
			}).Sort("date").All(&total)
		})
		mongo.M("dailyUsersCount", func(c *mgo.Collection) {
			c.Find(bson.M{
				"date": bson.M{"$gt": limit},
			}).Sort("date").All(&users)
		})
		render.JSON(c.Writer, http.StatusOK,
			gin.H{"total": total, "users": users})
	})

	r.GET("/api/rank/:date", func(c *gin.Context) {
		s := c.Params.ByName("date")
		loc, _ := time.LoadLocation("Asia/Shanghai")
		date, err := time.ParseInLocation("2006-01-02", s, loc)
		if err != nil {
			return
		}
		var result interface{}
		mongo.M("dailyRank", func(c *mgo.Collection) {
			c.Find(
				bson.M{"date": date},
			).One(&result)
		})
		render.JSON(c.Writer, http.StatusOK, result)
	})

	r.GET("/api/user/:name", func(c *gin.Context) {
		limit := time.Now().AddDate(0, 0, -100)
		s, err := url.QueryUnescape(c.Params.ByName("name"))
		if err != nil {
			return
		}
		var result []interface{}
		userid := conf.Redis.HGet("tgUsersName", s).Val()
		mongo.M("dailyUser", func(c *mgo.Collection) {
			c.Find(
				bson.M{
					"user": userid,
					"date": bson.M{"$gt": limit},
				},
			).Sort("date").All(&result)
		})
		render.JSON(c.Writer, http.StatusOK,
			gin.H{"result": result, "userName": s})
	})

	r.GET("/user/:name", func(c *gin.Context) {
		render.HTML(c.Writer, http.StatusOK, "user", nil)
	})

	ginpprof.Wrapper(r)
	r.Run(":6060")
}
