package main

import "gopkg.in/mgo.v2"

func MSession() *mgo.Session {
	if mc == nil {
		var err error
		mc, err = mgo.Dial(mgoUrl)
		if err != nil {
			loge.Error(err)
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
