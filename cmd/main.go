package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"coinbani/cmd/options"
	"coinbani/pkg/crypto"
	"coinbani/pkg/crypto/exchange"
	"coinbani/pkg/reply"

	tb "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	logger.Info("initializing coinbani bot")

	cfg := options.NewConfig()
	bot, err := tb.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = cfg.Bot.Debug
	logger.Info(fmt.Sprintf("authorized on account %s", bot.Self.UserName))

	u := tb.NewUpdate(0)
	u.Timeout = 60

	logger.Info("starting channel for getting bot updates")
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	bbExchange := exchange.NewBBExchange(cfg.Exchange, httpClient)

	cryptoService := crypto.NewService(bbExchange, logger)
	replyHandler := reply.NewHandler(bot, cryptoService, logger)

	logger.Info("coinbani bot successfully started!")
	for {
		select {
		case update, ok := <-updates:
			if !ok {
				logger.Info("channel closed, exiting")
				break
			}

			go replyHandler.HandleReply(update)
		}
	}

}
