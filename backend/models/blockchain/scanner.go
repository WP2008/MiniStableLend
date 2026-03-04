package blockchain

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"minilend/config"
	"minilend/dao"
	"minilend/models/blockchain/abi"
	"minilend/service"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ChainScanner struct {
	client          *ethclient.Client
	contracts       *ContractAddresses
	positionService *service.PositionService
	scanInterval    time.Duration
}

type ContractAddresses struct {
	MiniVault  common.Address
	MiniMUSD   common.Address
	MiniStETH  common.Address
	MiniOracle common.Address
}

// 事件签名哈希
var (
	depositETHEventHash    = abi.GetEventTopic("DepositETH")
	depositCollateralHash  = abi.GetEventTopic("DepositCollateral")
	borrowEventHash        = abi.GetEventTopic("Borrow")
	repayEventHash         = abi.GetEventTopic("Repay")
	liquidateEventHash     = abi.GetEventTopic("Liquidate")
	withdrawCollateralHash = abi.GetEventTopic("WithdrawCollateral")
)

func NewChainScanner(rpcURL string, contracts *ContractAddresses, scanInterval time.Duration) (*ChainScanner, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %v", err)
	}

	return &ChainScanner{
		client:          client,
		contracts:       contracts,
		positionService: &service.PositionService{},
		scanInterval:    scanInterval,
	}, nil
}

func (s *ChainScanner) StartScanning(ctx context.Context) {
	log.Println("Starting chain scanner...")

	// 从数据库加载上次扫描进度
	progress, err := dao.GetOrCreateScannerProgress(s.contracts.MiniVault.Hex())
	if err != nil {
		log.Printf("Failed to get scanner progress: %v, starting from block 0", err)
	} else {
		log.Printf("Loaded scanner progress: last scanned block = %d, total events = %d",
			progress.LastScannedBlock, progress.TotalEventsProcessed)
	}

	// 获取当前最新区块
	header, err := s.client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Printf("Failed to get latest block: %v", err)
		return
	}
	currentBlock := header.Number.Uint64()
	log.Printf("Current block height: %d", currentBlock)

	// 如果没有扫描记录，从上次记录继续扫描；否则从0开始或从配置的区块开始
	fromBlock := progress.LastScannedBlock
	if fromBlock == 0 {
		// 首次扫描，从当前区块前100个区块开始（避免丢失事件）
		if currentBlock > 100 {
			fromBlock = currentBlock - 100
		}
	}
	log.Printf("Starting scan from block: %d", fromBlock)

	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// 保存当前进度再退出
			if err := dao.UpdateLastScannedBlock(s.contracts.MiniVault.Hex(), fromBlock, 0); err != nil {
				log.Printf("Failed to save final progress: %v", err)
			}
			log.Println("Chain scanner stopped")
			return
		case <-ticker.C:
			fromBlock = s.scanNewBlocks(ctx, fromBlock)
		}
	}
}

func (s *ChainScanner) scanNewBlocks(ctx context.Context, fromBlock uint64) uint64 {
	// 获取最新区块
	header, err := s.client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Printf("Failed to get latest block: %v", err)
		return fromBlock
	}

	latestBlock := header.Number.Uint64()
	startBlock := fromBlock + 1

	if startBlock > latestBlock {
		return fromBlock // 没有新区块
	}

	toBlock := latestBlock
	if latestBlock-startBlock > 1000 { // 限制每次扫描的区块数量
		toBlock = startBlock + 1000
	}

	log.Printf("Scanning blocks from %d to %d", startBlock, toBlock)

	// 构建查询条件，只关注MiniVault合约的事件
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(startBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{s.contracts.MiniVault},
		Topics: [][]common.Hash{{
			depositETHEventHash,
			depositCollateralHash,
			borrowEventHash,
			repayEventHash,
			liquidateEventHash,
			withdrawCollateralHash}},
	}

	logs, err := s.client.FilterLogs(ctx, query)
	if err != nil {
		log.Printf("Failed to filter logs: %v", err)
		return fromBlock
	}

	for _, vLog := range logs {
		s.processLog(vLog)
	}

	// 保存扫描进度到数据库
	if err := dao.UpdateLastScannedBlock(s.contracts.MiniVault.Hex(), toBlock, int64(len(logs))); err != nil {
		log.Printf("Failed to save scanner progress: %v", err)
	} else {
		log.Printf("Saved scanner progress: block %d, processed %d events", toBlock, len(logs))
	}

	log.Printf("Finished scanning blocks %d to %d. Processed %d logs", startBlock, toBlock, len(logs))
	return toBlock
}

func (s *ChainScanner) processLog(vLog types.Log) {
	switch vLog.Topics[0] {
	case depositETHEventHash:
		s.processDepositEvent(vLog)
	case depositCollateralHash:
		s.processDepositCollateralEvent(vLog)
	case borrowEventHash:
		s.processBorrowEvent(vLog)
	case repayEventHash:
		s.processRepayEvent(vLog)
	case liquidateEventHash:
		s.processLiquidateEvent(vLog)
	case withdrawCollateralHash:
		s.processWithdrawCollateralEvent(vLog)
	}
}

func (s *ChainScanner) processDepositEvent(vLog types.Log) {
	event, err := abi.ParseDepositEvent(vLog.Data, vLog.Topics)
	if err != nil {
		log.Printf("Failed to parse deposit event: %v", err)
		return
	}

	log.Printf("DepositETH event: user=%s, amount=%s, stETHAmount=%s",
		event.User.Hex(), event.Amount.String(), event.StETHAmount.String())

	// TODO: 实现存款ETH事件的处理逻辑
}

func (s *ChainScanner) processDepositCollateralEvent(vLog types.Log) {
	event, err := abi.ParseDepositCollateralEvent(vLog.Data, vLog.Topics)
	if err != nil {
		log.Printf("Failed to parse deposit collateral event: %v", err)
		return
	}

	log.Printf("DepositCollateral event: user=%s, amount=%s",
		event.User.Hex(), event.Amount.String())

	// 更新数据库中的质押信息
	position, err := s.positionService.GetUserPosition(event.User.Hex())
	if err != nil {
		log.Printf("Position not found for user: %s, creating new position", event.User.Hex())
		position = &dao.Position{
			UserAddress:        event.User.Hex(),
			CollateralAmount:   convertWeiToFloat(event.Amount),
			DebtAmount:         0,
			InitialDebt:        0,
			HealthFactor:       0,
			InterestRate:       0,
			LastInterestUpdate: time.Now(),
			IsActive:           true,
			LiquidationPrice:   0,
		}
		if err := dao.CreatePosition(position); err != nil {
			log.Printf("Failed to create position: %v", err)
		}
	} else {
		position.CollateralAmount += convertWeiToFloat(event.Amount)
		if err := dao.UpdatePosition(position); err != nil {
			log.Printf("Failed to update position: %v", err)
		}
	}
}

func (s *ChainScanner) processBorrowEvent(vLog types.Log) {
	event, err := abi.ParseBorrowEvent(vLog.Data, vLog.Topics)
	if err != nil {
		log.Printf("Failed to parse borrow event: %v", err)
		return
	}

	log.Printf("Borrow event: user=%s, mUSDAmount=%s, fee=%s",
		event.User.Hex(), event.MUSDAmount.String(), event.Fee.String())

	// 更新数据库中的借款信息
	position, err := s.positionService.GetUserPosition(event.User.Hex())
	if err != nil {
		log.Printf("Position not found for user: %s", event.User.Hex())
		return
	}

	// Borrow事件包含借款金额和手续费，实际借款净额 = mUSDAmount - fee
	borrowAmount := new(big.Int).Sub(event.MUSDAmount, event.Fee)
	position.DebtAmount = convertWeiToFloat(borrowAmount)
	position.InitialDebt = convertWeiToFloat(borrowAmount)
	position.LastInterestUpdate = time.Now()

	if err := dao.UpdatePosition(position); err != nil {
		log.Printf("Failed to update position: %v", err)
	}
}

func (s *ChainScanner) processRepayEvent(vLog types.Log) {
	event, err := abi.ParseRepayEvent(vLog.Data, vLog.Topics)
	if err != nil {
		log.Printf("Failed to parse repay event: %v", err)
		return
	}

	log.Printf("Repay event: user=%s, amount=%s",
		event.User.Hex(), event.Amount.String())

	// 更新数据库中的还款信息
	position, err := s.positionService.GetUserPosition(event.User.Hex())
	if err != nil {
		log.Printf("Position not found for user: %s", event.User.Hex())
		return
	}

	// 还款减少债务
	position.DebtAmount -= convertWeiToFloat(event.Amount)
	if position.DebtAmount < 0 {
		position.DebtAmount = 0
	}
	position.LastInterestUpdate = time.Now()

	if err := dao.UpdatePosition(position); err != nil {
		log.Printf("Failed to update position: %v", err)
	}
}

func (s *ChainScanner) processLiquidateEvent(vLog types.Log) {
	event, err := abi.ParseLiquidateEvent(vLog.Data, vLog.Topics)
	if err != nil {
		log.Printf("Failed to parse liquidate event: %v", err)
		return
	}

	log.Printf("Liquidate event: liquidator=%s, borrower=%s, debtRepaid=%s, collateralSeized=%s, fee=%s",
		event.Liquidator.Hex(), event.Borrower.Hex(), event.DebtRepaid.String(),
		event.CollateralSeized.String(), event.Fee.String())

	// 更新数据库中的清算信息
	position, err := s.positionService.GetUserPosition(event.Borrower.Hex())
	if err != nil {
		log.Printf("Position not found for borrower: %s", event.Borrower.Hex())
		return
	}

	// 清算后仓位状态更新
	position.CollateralAmount -= convertWeiToFloat(event.CollateralSeized)
	position.DebtAmount = 0   // 债务清零
	position.IsActive = false // 仓位不再活跃

	if err := dao.UpdatePosition(position); err != nil {
		log.Printf("Failed to update position: %v", err)
	}
}

func (s *ChainScanner) processWithdrawCollateralEvent(vLog types.Log) {
	event, err := abi.ParseWithdrawCollateralEvent(vLog.Data, vLog.Topics)
	if err != nil {
		log.Printf("Failed to parse withdraw collateral event: %v", err)
		return
	}

	log.Printf("WithdrawCollateral event: user=%s, amount=%s, ethReceived=%s",
		event.User.Hex(), event.Amount.String(), event.ETHReceived.String())

	// 更新数据库中的提取质押物信息
	position, err := s.positionService.GetUserPosition(event.User.Hex())
	if err != nil {
		log.Printf("Position not found for user: %s", event.User.Hex())
		return
	}

	position.CollateralAmount -= convertWeiToFloat(event.Amount)
	if position.CollateralAmount < 0 {
		position.CollateralAmount = 0
	}

	if err := dao.UpdatePosition(position); err != nil {
		log.Printf("Failed to update position: %v", err)
	}
}

// convertWeiToFloat 将wei单位的数值转换为float64（18位小数）
func convertWeiToFloat(wei *big.Int) float64 {
	// 转换为以太单位（除以10^18）
	eth := new(big.Float).SetInt(wei)
	divisor := new(big.Float).SetFloat64(1e18)
	result, _ := new(big.Float).Quo(eth, divisor).Float64()
	return result
}

// 获取默认合约地址（本地开发网络）
func GetDefaultContractAddresses() *ContractAddresses {
	return &ContractAddresses{
		MiniVault:  common.HexToAddress(config.Config.Blockchain.MiniVaultAddress),
		MiniMUSD:   common.HexToAddress(config.Config.Blockchain.MiniMUSDAddress),
		MiniStETH:  common.HexToAddress(config.Config.Blockchain.MiniStETHAddress),
		MiniOracle: common.HexToAddress(config.Config.Blockchain.MiniOracleAddress),
	}
}
