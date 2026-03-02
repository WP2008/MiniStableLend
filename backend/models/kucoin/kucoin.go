package kucoin

import (
	"context"
	"minilend/dao"
	"minilend/log"

	"github.com/Kucoin/kucoin-go-sdk"
)

// ApiKeyVersionV2 is v2 api key version
const ApiKeyVersionV2 = "2"

const RedisKeyMUSDPrice = "musd_price"

var MUSDPrice = "0.0"
var MUSDPriceChan = make(chan string, 2)

func GetExchangePrice() {
	ctx := context.Background()
	log.Logger.Sugar().Info("GetExchangePrice ")

	// get plgr price from redis
	price, err := dao.RedisGetString(RedisKeyMUSDPrice)
	if err != nil {
		log.Logger.Sugar().Error("get musd price from redis err ", err)
	} else {
		MUSDPrice = price
	}

	s := kucoin.NewApiService(
		kucoin.ApiKeyOption("key"),
		kucoin.ApiSecretOption("secret"),
		kucoin.ApiPassPhraseOption("passphrase"),
		kucoin.ApiKeyVersionOption(ApiKeyVersionV2),
	)

	rsp, err := s.WebSocketPublicToken(ctx)
	if err != nil {
		log.Logger.Error(err.Error()) // Handle error
		return
	}

	tk := &kucoin.WebSocketTokenModel{}
	if err := rsp.ReadData(tk); err != nil {
		log.Logger.Error(err.Error())
		return
	}

	c := s.NewWebSocketClient(tk)

	mc, ec, err := c.Connect()
	if err != nil {
		log.Logger.Sugar().Errorf("Error: %s", err.Error())
		return
	}

	ch := kucoin.NewSubscribeMessage("/market/ticker:ETH-USDT", false)
	uch := kucoin.NewUnsubscribeMessage("/market/ticker:ETH-USDT", false)

	if err := c.Subscribe(ch); err != nil {
		log.Logger.Error(err.Error()) // Handle error
		return
	}

	for {
		select {
		case err := <-ec:
			c.Stop() // Stop subscribing the WebSocket feed
			log.Logger.Sugar().Errorf("Error: %s", err.Error())
			_ = c.Unsubscribe(uch)
			return
		case msg := <-mc:
			t := &kucoin.TickerLevel1Model{}
			if err := msg.ReadData(t); err != nil {
				log.Logger.Sugar().Errorf("Failure to read: %s", err.Error())
				return
			}
			MUSDPriceChan <- t.Price
			MUSDPrice = t.Price
			log.Logger.Sugar().Info("Price ", t.Price)
			_ = dao.RedisSetString(RedisKeyMUSDPrice, MUSDPrice, 0)
		}
	}
}
