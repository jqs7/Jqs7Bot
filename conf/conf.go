package conf

import (
	"log"
	"regexp"
	"strings"

	"github.com/fatih/set"
	"github.com/fsnotify/fsnotify"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/spf13/viper"
	"gopkg.in/redis.v3"
)

var (
	conf          *yaml.File
	Redis         *redis.Client
	Categories    []string
	CategoriesSet set.Interface
	Groups        []Group
)

type Group struct {
	GroupName string
	GroupURL  string
}

func init() {
	viper.SetConfigName("botconf")
	viper.AddConfigPath(".")
	viper.WatchConfig()
	if err := viper.ReadInConfig(); err != nil {
		log.Println(err.Error())
	}
	viper.OnConfigChange(func(e fsnotify.Event) {
		loadGroups()
	})
	loadGroups()

	Redis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: viper.GetString("redisPass"),
	})
}

func loadGroups() {
	Categories = viper.GetStringSlice("catagoris")
	CategoriesSet = set.New(set.NonThreadSafe)
	Groups = []Group{}
	for _, v := range Categories {
		CategoriesSet.Add(v)
		for _, i := range viper.GetStringSlice(v) {
			reg := regexp.MustCompile("^(.+) (http(s)?://(.*))$")
			strs := reg.FindAllStringSubmatch(i, -1)
			if !reg.MatchString(i) {
				Groups = append(Groups, Group{GroupName: i, GroupURL: ""})
				continue
			}
			if len(strs) > 0 {
				Groups = append(Groups, Group{GroupName: strs[0][1], GroupURL: strs[0][2]})
			}
		}
	}
}

func GetItem(i string) string {
	return viper.GetString(i)
}

func List2StringInConf(text string) string {
	resultGroup := List2SliceInConf(text)

	result := strings.Join(resultGroup, "\n")
	result = strings.Replace(result, "\\n", "", -1)

	return result
}

func List2SliceInConf(text string) []string {
	return viper.GetStringSlice(text)
}

type Question struct {
	Q string
	A set.Interface
}

func GetQuestions() []Question {
	var result []Question
	questions := viper.GetStringSlice("questions")
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
