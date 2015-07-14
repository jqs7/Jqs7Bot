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
	})
}
