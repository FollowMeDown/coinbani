package provider

import (
	"encoding/json"
	"math"
	"net/http"
	"strings"
	"time"

	"coinbani/cmd/coinbani/options"
	"coinbani/pkg/client"
	"coinbani/pkg/currency"

	"github.com/pkg/errors"
)

const (
	BBResponseExpiration = 10 * time.Minute
	BBResponseCacheKey   = "bb_response"
)

var parseBBResponseFunc = func(r *http.Response) (interface{}, error) {
	var bbResponse *BBResponse
	err := json.NewDecoder(r.Body).Decode(&bbResponse)
	if err != nil || bbResponse.Object == nil {
		return nil, errors.Wrap(err, "decoding BB response json")
	}
	defer r.Body.Close()

	return bbResponse, nil
}

type BBResponse struct {
	Object *BBObject `json:"object"`
}

type BBObject struct {
	DaiARS *BBPrice `json:"daiars"`
	DaiUSD *BBPrice `json:"daiusd"`
	BTCARS *BBPrice `json:"btcars"`
}

type BBPrice struct {
	BidPrice           float64 `json:"purchase_price,string"`
	BidCurrency        string  `json:"bid_currency"`
	AskPrice           float64 `json:"selling_price,string"`
	AskCurrency        string  `json:"ask_currency"`
	PriceChangePercent string  `json:"price_change_percent"`
	Currency           string  `json:"currency"`
	MarketIdentifier   string  `json:"market_identifier"`
}

type bbProvider struct {
	restClient client.Http
	config     *options.ProvidersConfig
}

func NewBBProvider(c *options.ProvidersConfig, r client.Http) *bbProvider {
	return &bbProvider{config: c, restClient: r}
}

func (p *bbProvider) FetchLastPrices() ([]*currency.CurrencyPrice, error) {
	var lastPrices []*currency.CurrencyPrice

	req := &client.GetRequestBuilder{
		Url:             p.config.BBURL,
		CacheKey:        BBResponseCacheKey,
		CacheExpiration: BBResponseExpiration,
		ParseResponse:   parseBBResponseFunc,
	}

	res, err := p.restClient.Get(req)
	if err != nil {
		return nil, errors.Wrap(err, "fetching prices from BB service")
	}

	bbResponse := res.(*BBResponse)

	// DAI ARS
	lastPrices = addCryptocurrencyBBPrice(lastPrices, bbResponse.Object.DaiARS)
	// DAI USD
	lastPrices = addCryptocurrencyBBPrice(lastPrices, bbResponse.Object.DaiUSD)
	// BTC ARS
	lastPrices = addCryptocurrencyBBPrice(lastPrices, bbResponse.Object.BTCARS)
	// ARS USD
	lastPrices = addUSDBPrice(lastPrices, bbResponse)

	return lastPrices, nil
}

func addUSDBPrice(lastPrices []*currency.CurrencyPrice, r *BBResponse) []*currency.CurrencyPrice {
	dai := r.Object.DaiARS.AskPrice
	usd := r.Object.DaiUSD.BidPrice

	lastPrices = append(lastPrices, &currency.CurrencyPrice{
		Desc:     "ARS/USD",
		BidPrice: 0,
		AskPrice: math.Round(dai/usd*100) / 100,
	})

	return lastPrices
}

func addCryptocurrencyBBPrice(lastPrices []*currency.CurrencyPrice, price *BBPrice) []*currency.CurrencyPrice {
	desc := strings.ToUpper(price.BidCurrency) + "/" + strings.ToUpper(price.AskCurrency)

	lastPrices = append(lastPrices, &currency.CurrencyPrice{
		Desc:     desc,
		Currency: strings.Replace(price.Currency, "$", "S", -1),
		BidPrice: price.BidPrice,
		AskPrice: price.AskPrice,
	})

	return lastPrices
}
