package main

import (
	"log"

	"github.com/carlescere/scheduler"
	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/jqs7/Jqs7Bot/mongo"
	"github.com/jqs7/Jqs7Bot/plugin"
	"github.com/jqs7/Jqs7Bot/webServer"
	"github.com/jqs7/bb"
)

func main() {
	bot := bb.LoadBot(conf.GetItem("botapi"))
	if conf.GetItem("runMode") == "debug" {
		bot.SetUpdate(10)
	} else {
		bot.SetWebhook("jqs7.com", "8443",
			"crt.pem", "key.pem")
	}
	if bot.Err != nil {
		log.Println("bot connection failed")
		log.Println(bot.Err)
		return
	}

	go plugin.InitRss(bot.GetBot())
	go func() {
		mongo.MIndex()
		mongo.DailySave()
		scheduler.Every().Day().At("00:05").Run(mongo.DailySave)
	}()
	go webServer.GinServer()

	botName := bot.GetBot().Self.UserName
	bot.Prepare(&plugin.Prepare{}).
		Plugin(new(plugin.Start), "/help", "/start", "/help@"+botName, "/start@"+botName).
		Plugin(new(plugin.Rule), "/rule", "/rule@"+botName).
		Plugin(new(plugin.SetRule), "/setrule").
		Plugin(new(plugin.RmRule), "/rmrule").
		Plugin(new(plugin.AutoRule), "/autorule").
		Plugin(new(plugin.About), "/about", "/about@"+botName).
		Plugin(new(plugin.OtherResources), "/other_resources", "/other_resources@"+botName).
		Plugin(new(plugin.Subscribe), "/subscribe", "/subscribe@"+botName).
		Plugin(new(plugin.UnSubscribe), "/unsubscribe", "/unsubscribe@"+botName).
		Plugin(new(plugin.Broadcast), "/broadcast").
		Plugin(new(plugin.Groups), "/groups", "/groups@"+botName).
		Plugin(new(plugin.Cancel), "/cancel").
		Plugin(new(plugin.Base64), "/e64", "/d64").
		Plugin(new(plugin.Google), "/gg").
		Plugin(new(plugin.Trans), "/trans").
		Plugin(new(plugin.Man), "/man", "/setman", "/rmman").
		Plugin(new(plugin.Reload), "/reload").
		Plugin(new(plugin.Stat), "/os", "/df", "/free", "/redis").
		Plugin(new(plugin.Rain), "/rain").
		Plugin(new(plugin.Rss), "/rss", "/rmrss").
		Plugin(new(plugin.Markdown), "/md").
		Plugin(new(plugin.Search), "/search").
		Plugin(new(plugin.Turing), "@"+botName).
		Default(&plugin.Default{}).
		Start()
}
