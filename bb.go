package bb

import (
	"log"
	"net/http"
	"strings"

	"github.com/Syfaro/telegram-bot-api"
)

var bot *tgbotapi.BotAPI

var plugins []plugin

type bb struct {
	Err error
}

func LoadBot(token string) *bb {
	var err error
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		return &bb{err}
	}
	return &bb{nil}
}

func (b *bb) SetWebhook(domain, port, crt, key string) *bb {
	if b.Err != nil {
		return b
	}
	hook := tgbotapi.NewWebhookWithCert("https://"+
		domain+":"+port+"/"+bot.Token, crt)
	_, err := bot.SetWebhook(hook)
	bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServeTLS(":"+port, crt, key, nil)
	return &bb{err}
}

func (b *bb) SetUpdate() *bb {
	hook := tgbotapi.NewWebhook("")
	_, err := bot.SetWebhook(hook)
	if err != nil {
		return &bb{err}
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	err = bot.UpdatesChan(u)
	return &bb{err}
}

func (b *bb) Plugin(e pluginInterface, commands ...string) *bb {
	plugin := plugin{
		commands,
		e.Run,
		e.init,
	}
	plugins = append(plugins, plugin)
	return &bb{nil}
}

var prepare struct {
	run func()
}

func (b *bb) Prepare(e pluginInterface) *bb {
	prepare.run = e.Run
	return &bb{nil}
}

var finish struct {
	run func()
}

func (b *bb) Finish(e pluginInterface) *bb {
	finish.run = e.Run
	return &bb{nil}
}

var _default struct {
	run func()
}

func (b *bb) Default(e pluginInterface) *bb {
	_default.run = e.Run
	return &bb{nil}
}

func (b *bb) Start() {
	if b.Err != nil {
		log.Panicln(b.Err)
		return
	}
	for update := range bot.Updates {
		go func(update tgbotapi.Update) {
			if prepare.run != nil {
				prepare.run()
			}
			args := strings.FieldsFunc(update.Message.Text,
				func(r rune) bool {
					switch r {
					case '\t', '\v', '\f', '\r', ' ', 0xA0:
						return true
					}
					return false
				})

			match := false
		RangePlugins:
			for _, plugin := range plugins {
				for _, command := range plugin.commands {
					if command == args[0] {
						plugin.init(bot, update.UpdateID, update.Message, args)
						plugin.run()
						match = true
						break RangePlugins
					}
				}
			}
			if !match && _default.run != nil {
				_default.run()
			}
			if finish.run != nil {
				finish.run()
			}
		}(update)
	}
}

type Base struct {
	Bot       *tgbotapi.BotAPI
	UpdateID  int
	FromGroup bool
	Message   tgbotapi.Message
	Args      []string
}

func (b *Base) init(bot *tgbotapi.BotAPI, updateID int,
	message tgbotapi.Message, args []string) {
	b.Bot = bot
	b.UpdateID = updateID
	b.Message = message
	b.Args = args
	if message.IsGroup() {
		b.FromGroup = true
	} else {
		b.FromGroup = false
	}
}

func (b *Base) Run() {
	log.Println("default run func")
}

type plugin struct {
	commands []string
	run      func()
	init     func(*tgbotapi.BotAPI, int, tgbotapi.Message, []string)
}

type pluginInterface interface {
	Run()
	init(*tgbotapi.BotAPI, int, tgbotapi.Message, []string)
}
