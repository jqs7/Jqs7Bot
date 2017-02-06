package bb

import (
	"log"
	"net/http"
	"strings"

	"github.com/Syfaro/telegram-bot-api"
)

var plugins []plugin

type bb struct {
	bot *tgbotapi.BotAPI
	Err error
}

func LoadBot(token string) *bb {
	var err error
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return &bb{nil, err}
	}
	return &bb{bot, nil}
}

func (b *bb) SetWebhook(domain, port, crt, key string) *bb {
	if b.Err != nil {
		return b
	}
	hook := tgbotapi.NewWebhookWithCert("https://"+
		domain+":"+port+"/"+b.bot.Token, crt)
	_, err := b.bot.SetWebhook(hook)
	b.bot.ListenForWebhook("/" + b.bot.Token)
	go http.ListenAndServeTLS(":"+port, crt, key, nil)
	return &bb{b.bot, err}
}

func (b *bb) SetUpdate(timeout int) *bb {
	if b.Err != nil {
		return b
	}
	hook := tgbotapi.NewWebhook("")
	_, err := b.bot.SetWebhook(hook)
	if err != nil {
		return &bb{b.bot, err}
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout
	err = b.bot.UpdatesChan(u)
	return &bb{b.bot, err}
}

func (b *bb) Plugin(e PluginInterface, commands ...string) *bb {
	plugin := plugin{
		commands,
		base{
			run:     e.Run,
			handler: e.handler,
		},
	}
	plugins = append(plugins, plugin)
	return &bb{b.bot, nil}
}

func (b *bb) GetBot() *tgbotapi.BotAPI {
	return b.bot
}

var prepare struct {
	base
}

func (b *bb) Prepare(e PluginInterface) *bb {
	prepare.run = e.Run
	prepare.handler = e.handler
	return &bb{b.bot, nil}
}

var finish struct {
	base
}

func (b *bb) Finish(e PluginInterface) *bb {
	finish.run = e.Run
	finish.handler = e.handler
	return &bb{b.bot, nil}
}

var _default struct {
	base
}

func (b *bb) Default(e PluginInterface) *bb {
	_default.run = e.Run
	_default.handler = e.handler
	return &bb{b.bot, nil}
}

func (b *bb) Start() {
	if b.Err != nil {
		log.Panicln(b.Err)
		return
	}
Update:
	for update := range b.bot.Updates {
		args := strings.FieldsFunc(update.Message.Text,
			func(r rune) bool {
				switch r {
				case '\t', '\v', '\f', '\r', ' ', 0xA0:
					return true
				}
				return false
			})

		if prepare.run != nil {
			prepare.handler(b.bot, update, args)
			isPanic := false
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Println(err)
						isPanic = true
					}
				}()
				prepare.run()
			}()
			if isPanic {
				continue
			}
		}

		match := false
		if len(args) > 0 {
		RangePlugins:
			for _, plugin := range plugins {
				for _, command := range plugin.commands {
					if command == args[0] {
						plugin.handler(b.bot, update, args)
						isPanic := false
						func() {
							defer func() {
								if err := recover(); err != nil {
									log.Println(err)
									isPanic = true
								}
							}()
							plugin.run()
						}()
						if isPanic {
							continue Update
						}
						match = true
						break RangePlugins
					}
				}
			}
		}

		if !match && _default.run != nil {
			_default.handler(b.bot, update, args)
			isPanic := false
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Println(err)
						isPanic = true
					}
				}()
				_default.run()
			}()
			if isPanic {
				continue
			}
		}
		if finish.run != nil {
			finish.handler(b.bot, update, args)
			isPanic := false
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Println(err)
						isPanic = true
					}
				}()
				finish.run()
			}()
			if isPanic {
				continue
			}
		}
	}
}

type Base struct {
	UpdateID       int
	ChatID         int
	FromGroup      bool
	FromChannel    bool
	FromPrivate    bool
	FromSuperGroup bool
	Args           []string
	Bot            *tgbotapi.BotAPI
	Message        tgbotapi.Message
}

func (b *Base) handler(bot *tgbotapi.BotAPI, update tgbotapi.Update, args []string) {
	b.Bot = bot
	b.UpdateID = update.UpdateID
	b.Message = update.Message
	b.Args = args
	b.ChatID = update.Message.Chat.ID
	b.FromGroup = update.Message.Chat.IsGroup()
	b.FromChannel = update.Message.Chat.IsChannel()
	b.FromPrivate = update.Message.Chat.IsPrivate()
	b.FromSuperGroup = (update.Message.Chat.Type == "supergroup")
}

func (b *Base) Run() {
	log.Println("default run func")
}

type plugin struct {
	commands []string
	base
}

type PluginInterface interface {
	Run()
	handler(*tgbotapi.BotAPI, tgbotapi.Update, []string)
}

type base struct {
	run     func()
	handler func(*tgbotapi.BotAPI, tgbotapi.Update, []string)
}
