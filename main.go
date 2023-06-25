package main

import (
	"os"
	"os/signal"

	bot2 "github.com/mbeoliero/bot/bot"
	"github.com/mbeoliero/bot/conf"
)

func main() {
	conf.InitConfig()

	bot := bot2.NewQQBot(conf.Get().Account.Uid, conf.Get().Account.Password,
		conf.Get().WhiteList.UserList, conf.Get().WhiteList.GroupList)
	bot.Start()
	defer bot.Stop()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
