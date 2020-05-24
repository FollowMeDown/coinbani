package options

import (
	"log"
	"os"
	"strconv"
)

type config struct {
	Bot       *BotConfig
	Providers *ProvidersConfig
}

func NewConfig() *config {
	debug, err := strconv.ParseBool(os.Getenv("BOT_DEBUG"))
	if err != nil {
		log.Panic(err)
	}

	return &config{
		Bot: &BotConfig{
			Token: os.Getenv("BOT_TOKEN"),
			Debug: debug,
		},
		Providers: &ProvidersConfig{
			BBURL:         os.Getenv("BB_URL"),
			SatoshiARSURL: os.Getenv("SATOSHI_ARS_URL"),
			SatoshiUSDURL: os.Getenv("SATOSHI_USD_URL"),
			DolarURL:      os.Getenv("DOLAR_URL"),
		},
	}
}

type BotConfig struct {
	Token string
	Debug bool
}

type ProvidersConfig struct {
	BBURL         string
	SatoshiARSURL string
	SatoshiUSDURL string
	DolarURL      string
}
