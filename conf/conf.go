package conf

import (
	"log"
	"strconv"
	"strings"

	"gopkg.in/redis.v3"

	"github.com/fatih/set"
	"github.com/kylelemons/go-gypsy/yaml"
)

var (
	conf          *yaml.File
	Redis         *redis.Client
	Categories    []string
	CategoriesSet set.Interface
)

func init() {
	LoadConf()
	redisPass, _ := conf.Get("redisPass")
	Redis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: redisPass,
	})

	Categories = []string{
		"Linux", "Programming", "Software",
		"影音", "科幻", "ACG", "IT", "社区",
		"闲聊", "资源", "同城", "Others",
	}
	CategoriesSet = set.New(set.NonThreadSafe)
	for _, v := range Categories {
		CategoriesSet.Add(v)
	}
}

func LoadConf() {
	var err error
	conf, err = yaml.ReadFile("botconf.yaml")
	if err != nil {
		log.Panic(err)
	}
}

func GetItem(i string) string {
	item, _ := conf.Get(i)
	return item
}

func List2StringInConf(text string) string {
	resultGroup := List2SliceInConf(text)

	result := strings.Join(resultGroup, "\n")
	result = strings.Replace(result, "\\n", "", -1)

	return result
}

func List2SliceInConf(text string) []string {
	count, err := conf.Count(text)
	if err != nil {
		log.Println(err)
		return nil
	}

	var result []string
	for i := 0; i < count; i++ {
		v, err := conf.Get(text + "[" + strconv.Itoa(i) + "]")
		if err != nil {
			log.Println(err)
			return nil
		}
		result = append(result, v)
	}
	return result
}

type Question struct {
	Q string
	A set.Interface
}

func GetQuestions() []Question {
	var result []Question
	questions := List2SliceInConf("questions")

	for _, v := range questions {
		qs := strings.Split(v, "|")
		question := qs[0]
		answers := strings.Split(qs[1], ";")

		s := set.New(set.ThreadSafe)
		for _, v := range answers {
			s.Add(v)
		}
		result = append(result, Question{question, s})
	}
	return result
}
