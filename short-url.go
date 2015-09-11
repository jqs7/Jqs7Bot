package main

import (
	"fmt"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/franela/goreq"
)

func ShortUrl(url string) string {
	res, err := goreq.Request{
		Uri: fmt.Sprintf(
			"http://v.gd/create.php?format=json&url=",
			url),
		Timeout: 17 * time.Second,
	}.Do()
	if err != nil {
		loge.Warning("Short Failed!")
		return url
	}
	jasonObj, err := jason.NewObjectFromReader(res.Body)
	if err != nil {
		return url
	}
	s, err := jasonObj.GetString("shorturl")
	if err != nil {
		e, _ := jasonObj.GetString("errormessage")
		loge.Warning(e)
		return url
	}
	return s
}
