package main

import (
	"testing"

	"github.com/fatih/set"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/smartystreets/goconvey/convey"
)

func Test(t *testing.T) {
	convey.Convey("YamlList2String Test", t, func() {
		conf, err := yaml.ReadFile("botconf_test.yaml")
		convey.So(err, convey.ShouldBeNil)
		convey.So(YamlList2String(conf, "listTest"), convey.ShouldEqual, "1\n2\n\n3")
	})

	convey.Convey("Questions Test", t, func() {
		conf, err := yaml.ReadFile("botconf_test.yaml")
		convey.So(err, convey.ShouldBeNil)
		s1 := set.New(set.ThreadSafe)
		s2 := set.New(set.ThreadSafe)
		s1.Add("A1", "A2", "A3")
		s2.Add("A4", "A5", "A6")
		result := []Question{Question{"Q1", s1}, Question{"Q2", s2}}
		convey.So(GetQuestions(conf, "questionsTest"), convey.ShouldHaveSameTypeAs, result)
		convey.So(GetQuestions(conf, "questionsTest"), convey.ShouldResemble, result)
	})

	convey.Convey("To2dSlice Test", t, func() {
		in := []string{"1", "2", "3", "4", "5"}
		out := [][]string{[]string{"1", "2", "3"}, []string{"4", "5"}}
		convey.So(To2dSlice(in, 3, 2), convey.ShouldResemble, out)
	})

	convey.Convey("Vim-Tips Test", t, func() {
		t := <-VimTipsChan(1)
		convey.So(t.Comment, convey.ShouldNotBeBlank)
		convey.So(t.Content, convey.ShouldNotBeBlank)
	})

	convey.Convey("Hitokoto Test", t, func() {
		h := <-HitokotoChan(1)
		convey.So(h.Hitokoto, convey.ShouldNotBeBlank)
		convey.So(h.Source, convey.ShouldNotBeBlank)
	})

	convey.Convey("VH Test", t, func() {
		v := <-VH(1)
		convey.So(v, convey.ShouldNotBeBlank)
	})

	convey.Convey("Base64 Test", t, func() {
		convey.So(E64("Hello"), convey.ShouldEqual, "SGVsbG8=")
		convey.So(D64("sdjaikdbsa"), convey.ShouldEqual,
			"解码系统出现故障，请查看弹药是否填充无误")
		convey.So(D64("SGVsbG8="), convey.ShouldEqual, "Hello")
	})

	convey.Convey("Translate Test", t, func() {
		conf, err := yaml.ReadFile("botconf_test.yaml")
		convey.So(err, convey.ShouldBeNil)

		convey.So(BaiduTranslate("123", "Hello"), convey.ShouldEqual,
			"大概男盆友用错API Key啦，大家快去蛤他！σ`∀´)`")

		key, err := conf.Get("baiduTransKey")
		convey.So(err, convey.ShouldBeNil)
		convey.So(BaiduTranslate(key, "你好"), convey.ShouldEqual, "Hello")
		convey.So(BaiduTranslate(key, "Hello"), convey.ShouldEqual, "你好")
	})
}
