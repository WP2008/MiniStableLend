package main

import (
	"context"
	"log"
	"minilend/api/router"
	"minilend/config"
	"minilend/dao"
	"minilend/models/blockchain"
	"minilend/models/kucoin"
	"minilend/models/ws"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库
	dao.NewDao()

	go ws.StartServer()

	go kucoin.GetExchangePrice()

	// 启动链扫描服务
	go startBlockchainScanner()

	// 创建Gin引擎
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 注册路由
	router.RegisterRoutes(r)

	// 启动HTTP服务器
	if err := r.Run(":" + config.Config.Env.Port); err != nil {
		log.Fatalln(err)
	}
}

func startBlockchainScanner() {
	contracts := blockchain.GetDefaultContractAddresses()
	scanner, err := blockchain.NewChainScanner(
		config.Config.Blockchain.LocalRPCURL,
		contracts,
		time.Duration(config.Config.Blockchain.ScanIntervalSecs)*time.Second,
	)
	if err != nil {
		log.Printf("Failed to create blockchain scanner: %v", err)
		return
	}

	scanner.StartScanning(context.Background())
}
