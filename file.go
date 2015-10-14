package bb

import "github.com/Syfaro/telegram-bot-api"

type file struct {
	Err    error
	bot    *tgbotapi.BotAPI
	config tgbotapi.FileConfig
	Ret    tgbotapi.File
}

func (b *Base) File(fileID string) *file {
	return &file{
		bot:    b.Bot,
		config: tgbotapi.FileConfig{fileID},
	}
}

func (f *file) Get() *file {
	file, err := f.bot.GetFile(f.config)
	f.Ret = file
	f.Err = err
	return f
}

func (f *file) Link() string {
	if f.Err != nil {
		return ""
	}
	return f.Ret.Link(f.bot.Token)
}
