package handler

import (
	"minilend/models/ws"
	"minilend/utils"
	"net/http"
	"strings"
	"time"

	"minilend/log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type PriceHander struct {
}

func (c PriceHander) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/price", c.NewPrice)
}

func (c *PriceHander) NewPrice(ctx *gin.Context) {

	defer func() {
		recoverRes := recover()
		if recoverRes != nil {
			log.Logger.Sugar().Error("new price recover ", recoverRes)
		}
	}()

	conn, err := (&websocket.Upgrader{
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 5 * time.Second,
		CheckOrigin: func(r *http.Request) bool { //Cross domain
			return true
		},
	}).Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Logger.Sugar().Error("websocket request err:", err)
		return
	}

	randomId := ""
	remoteIP := ctx.RemoteIP()
	if len(remoteIP) > 0 {
		randomId = strings.Replace(remoteIP, ".", "_", -1) + "_" + utils.GetRandomString(23)
	} else {
		randomId = utils.GetRandomString(32)
	}
	server := &ws.Server{
		Id:       randomId,
		Socket:   conn,
		Send:     make(chan []byte, 800),
		LastTime: time.Now().Unix(),
	}

	go server.ReadAndWrite()
}
