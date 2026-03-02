package abi

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// MiniVaultABI 包含 MiniVault 合约的 ABI 定义
const MiniVaultABI = `[
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "user",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "collateralAmount",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "mintedAmount",
				"type": "uint256"
			}
		],
		"name": "Deposit",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "user",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "interestRate",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "newDebt",
				"type": "uint256"
			}
		],
		"name": "Borrow",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "user",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "remainingDebt",
				"type": "uint256"
			}
		],
		"name": "Repay",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "liquidator",
				"type": "address"
			},
			{
				"indexed": true,
				"internalType": "address",
				"name": "user",
				"type": "address"
		},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "collateralSeized",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "debtRepaid",
				"type": "uint256"
			}
		],
		"name": "Liquidate",
		"type": "event"
	}`

// DepositEvent 存款事件数据结构
type DepositEvent struct {
	User             common.Address
	CollateralAmount *big.Int
	MintedAmount     *big.Int
}

// BorrowEvent 借款事件数据结构
type BorrowEvent struct {
	User         common.Address
	Amount       *big.Int
	InterestRate *big.Int
	NewDebt      *big.Int
}

// RepayEvent 还款事件数据结构
type RepayEvent struct {
	User          common.Address
	Amount        *big.Int
	RemainingDebt *big.Int
}

// LiquidateEvent 清算事件数据结构
type LiquidateEvent struct {
	Liquidator       common.Address
	User             common.Address
	CollateralSeized *big.Int
	DebtRepaid       *big.Int
}

var miniVaultABI abi.ABI

func init() {
	var err error
	miniVaultABI, err = abi.JSON(strings.NewReader(MiniVaultABI))
	if err != nil {
		panic("Failed to parse MiniVault ABI: " + err.Error())
	}
}

// ParseDepositEvent 解析存款事件日志
func ParseDepositEvent(logData []byte, topics []common.Hash) (*DepositEvent, error) {
	var event DepositEvent

	// 第一个topic是事件签名，第二个topic是user地址
	if len(topics) < 2 {
		return nil, fmt.Errorf("insufficient topics for deposit event")
	}

	event.User = common.BytesToAddress(topics[1].Bytes())

	// 解析数据部分
	err := miniVaultABI.UnpackIntoInterface(&event, "Deposit", logData)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack deposit event: %v", err)
	}

	return &event, nil
}

// ParseBorrowEvent 解析借款事件日志
func ParseBorrowEvent(logData []byte, topics []common.Hash) (*BorrowEvent, error) {
	var event BorrowEvent

	if len(topics) < 2 {
		return nil, fmt.Errorf("insufficient topics for borrow event")
	}

	event.User = common.BytesToAddress(topics[1].Bytes())

	err := miniVaultABI.UnpackIntoInterface(&event, "Borrow", logData)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack borrow event: %v", err)
	}

	return &event, nil
}

// ParseRepayEvent 解析还款事件日志
func ParseRepayEvent(logData []byte, topics []common.Hash) (*RepayEvent, error) {
	var event RepayEvent

	if len(topics) < 2 {
		return nil, fmt.Errorf("insufficient topics for repay event")
	}

	event.User = common.BytesToAddress(topics[1].Bytes())

	err := miniVaultABI.UnpackIntoInterface(&event, "Repay", logData)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack repay event: %v", err)
	}

	return &event, nil
}

// ParseLiquidateEvent 解析清算事件日志
func ParseLiquidateEvent(logData []byte, topics []common.Hash) (*LiquidateEvent, error) {
	var event LiquidateEvent

	if len(topics) < 3 {
		return nil, fmt.Errorf("insufficient topics for liquidate event")
	}

	event.Liquidator = common.BytesToAddress(topics[1].Bytes())
	event.User = common.BytesToAddress(topics[2].Bytes())

	err := miniVaultABI.UnpackIntoInterface(&event, "Liquidate", logData)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack liquidate event: %v", err)
	}

	return &event, nil
}

// GetEventTopic 获取事件签名对应的topic
func GetEventTopic(eventName string) common.Hash {
	return miniVaultABI.Events[eventName].ID
}

// LogToHex 将日志数据转换为十六进制字符串（用于调试）
func LogToHex(data []byte) string {
	return "0x" + hex.EncodeToString(data)
}
