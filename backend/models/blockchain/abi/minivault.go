package abi

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// MiniVaultABI 包含 MiniVault 合约的 ABI 定义
const MiniVaultABI = `[
    {
        "inputs": [
            {
                "internalType": "address",
                "name": "_stETH",
                "type": "address"
            },
            {
                "internalType": "address",
                "name": "_mUSD",
                "type": "address"
            },
            {
                "internalType": "address",
                "name": "_oracle",
                "type": "address"
            },
            {
                "internalType": "address",
                "name": "_treasury",
                "type": "address"
            },
            {
                "internalType": "address",
                "name": "_timelock",
                "type": "address"
            }
        ],
        "stateMutability": "nonpayable",
        "type": "constructor"
    },
    {
        "inputs": [

        ],
        "name": "AccessControlBadConfirmation",
        "type": "error"
    },
    {
        "inputs": [
            {
                "internalType": "address",
                "name": "account",
                "type": "address"
            },
            {
                "internalType": "bytes32",
                "name": "neededRole",
                "type": "bytes32"
            }
        ],
        "name": "AccessControlUnauthorizedAccount",
        "type": "error"
    },
    {
        "inputs": [

        ],
        "name": "ReentrancyGuardReentrantCall",
        "type": "error"
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
                "name": "mUSDAmount",
                "type": "uint256"
            },
            {
                "indexed": false,
                "internalType": "uint256",
                "name": "fee",
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
            }
        ],
        "name": "DepositCollateral",
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
                "name": "stETHAmount",
                "type": "uint256"
            }
        ],
        "name": "DepositETH",
        "type": "event"
    },
    {
        "anonymous": false,
        "inputs": [
            {
                "indexed": false,
                "internalType": "uint256",
                "name": "oldRate",
                "type": "uint256"
            },
            {
                "indexed": false,
                "internalType": "uint256",
                "name": "newRate",
                "type": "uint256"
            }
        ],
        "name": "ExchangeRateUpdated",
        "type": "event"
    },
    {
        "anonymous": false,
        "inputs": [
            {
                "indexed": false,
                "internalType": "uint256",
                "name": "toStETH",
                "type": "uint256"
            },
            {
                "indexed": false,
                "internalType": "uint256",
                "name": "toTreasury",
                "type": "uint256"
            }
        ],
        "name": "InterestDistributed",
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
                "name": "borrower",
                "type": "address"
            },
            {
                "indexed": false,
                "internalType": "uint256",
                "name": "debtRepaid",
                "type": "uint256"
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
                "name": "fee",
                "type": "uint256"
            }
        ],
        "name": "Liquidate",
        "type": "event"
    },
    {
        "anonymous": false,
        "inputs": [
            {
                "indexed": false,
                "internalType": "string",
                "name": "paramName",
                "type": "string"
            },
            {
                "indexed": false,
                "internalType": "uint256",
                "name": "oldValue",
                "type": "uint256"
            },
            {
                "indexed": false,
                "internalType": "uint256",
                "name": "newValue",
                "type": "uint256"
            }
        ],
        "name": "ParamsUpdated",
        "type": "event"
    },
    {
        "anonymous": false,
        "inputs": [
            {
                "indexed": true,
                "internalType": "address",
                "name": "account",
                "type": "address"
            }
        ],
        "name": "Paused",
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
                "internalType": "bytes32",
                "name": "role",
                "type": "bytes32"
            },
            {
                "indexed": true,
                "internalType": "bytes32",
                "name": "previousAdminRole",
                "type": "bytes32"
            },
            {
                "indexed": true,
                "internalType": "bytes32",
                "name": "newAdminRole",
                "type": "bytes32"
            }
        ],
        "name": "RoleAdminChanged",
        "type": "event"
    },
    {
        "anonymous": false,
        "inputs": [
            {
                "indexed": true,
                "internalType": "bytes32",
                "name": "role",
                "type": "bytes32"
            },
            {
                "indexed": true,
                "internalType": "address",
                "name": "account",
                "type": "address"
            },
            {
                "indexed": true,
                "internalType": "address",
                "name": "sender",
                "type": "address"
            }
        ],
        "name": "RoleGranted",
        "type": "event"
    },
    {
        "anonymous": false,
        "inputs": [
            {
                "indexed": true,
                "internalType": "bytes32",
                "name": "role",
                "type": "bytes32"
            },
            {
                "indexed": true,
                "internalType": "address",
                "name": "account",
                "type": "address"
            },
            {
                "indexed": true,
                "internalType": "address",
                "name": "sender",
                "type": "address"
            }
        ],
        "name": "RoleRevoked",
        "type": "event"
    },
    {
        "anonymous": false,
        "inputs": [
            {
                "indexed": true,
                "internalType": "address",
                "name": "account",
                "type": "address"
            }
        ],
        "name": "Unpaused",
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
                "name": "ethReceived",
                "type": "uint256"
            }
        ],
        "name": "WithdrawCollateral",
        "type": "event"
    },
    {
        "inputs": [

        ],
        "name": "BASE",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "DEFAULT_ADMIN_ROLE",
        "outputs": [
            {
                "internalType": "bytes32",
                "name": "",
                "type": "bytes32"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "INTEREST_TREASURY_SHARE",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "LIQUIDATION_DISCOUNT",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "LIQUIDATION_MAX_PCT",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "LIQUIDATION_TREASURY_FEE",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "PAUSER_ROLE",
        "outputs": [
            {
                "internalType": "bytes32",
                "name": "",
                "type": "bytes32"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "RATE_UPDATE_INTERVAL",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "TIMELOCK_ROLE",
        "outputs": [
            {
                "internalType": "bytes32",
                "name": "",
                "type": "bytes32"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "uint256",
                "name": "mUSDAmount",
                "type": "uint256"
            }
        ],
        "name": "borrow",
        "outputs": [

        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "borrowAPR",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "borrowFeePct",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "borrowLimitPct",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "uint256",
                "name": "amount",
                "type": "uint256"
            }
        ],
        "name": "depositCollateral",
        "outputs": [

        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "depositETH",
        "outputs": [

        ],
        "stateMutability": "payable",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "address",
                "name": "token",
                "type": "address"
            },
            {
                "internalType": "uint256",
                "name": "amount",
                "type": "uint256"
            },
            {
                "internalType": "address",
                "name": "to",
                "type": "address"
            }
        ],
        "name": "emergencyWithdraw",
        "outputs": [

        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "address",
                "name": "user",
                "type": "address"
            }
        ],
        "name": "getBorrowLimit",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "address",
                "name": "user",
                "type": "address"
            }
        ],
        "name": "getHealthFactor",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "address",
                "name": "user",
                "type": "address"
            }
        ],
        "name": "getPosition",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "collateral",
                "type": "uint256"
            },
            {
                "internalType": "uint256",
                "name": "debt",
                "type": "uint256"
            },
            {
                "internalType": "uint256",
                "name": "initialDebt",
                "type": "uint256"
            },
            {
                "internalType": "uint256",
                "name": "lastUpdate",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "bytes32",
                "name": "role",
                "type": "bytes32"
            }
        ],
        "name": "getRoleAdmin",
        "outputs": [
            {
                "internalType": "bytes32",
                "name": "",
                "type": "bytes32"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "bytes32",
                "name": "role",
                "type": "bytes32"
            },
            {
                "internalType": "address",
                "name": "account",
                "type": "address"
            }
        ],
        "name": "grantRole",
        "outputs": [

        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "bytes32",
                "name": "role",
                "type": "bytes32"
            },
            {
                "internalType": "address",
                "name": "account",
                "type": "address"
            }
        ],
        "name": "hasRole",
        "outputs": [
            {
                "internalType": "bool",
                "name": "",
                "type": "bool"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "lastRateUpdate",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "address",
                "name": "borrower",
                "type": "address"
            },
            {
                "internalType": "uint256",
                "name": "debtRepaid",
                "type": "uint256"
            }
        ],
        "name": "liquidate",
        "outputs": [

        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "liquidationThreshold",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "mUSD",
        "outputs": [
            {
                "internalType": "contract IMiniMUSD",
                "name": "",
                "type": "address"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "oracle",
        "outputs": [
            {
                "internalType": "contract IMiniOracle",
                "name": "",
                "type": "address"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "pause",
        "outputs": [

        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "paused",
        "outputs": [
            {
                "internalType": "bool",
                "name": "",
                "type": "bool"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "address",
                "name": "",
                "type": "address"
            }
        ],
        "name": "positions",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "collateral",
                "type": "uint256"
            },
            {
                "internalType": "uint256",
                "name": "debt",
                "type": "uint256"
            },
            {
                "internalType": "uint256",
                "name": "initialDebt",
                "type": "uint256"
            },
            {
                "internalType": "uint256",
                "name": "lastUpdate",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "bytes32",
                "name": "role",
                "type": "bytes32"
            },
            {
                "internalType": "address",
                "name": "callerConfirmation",
                "type": "address"
            }
        ],
        "name": "renounceRole",
        "outputs": [

        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "uint256",
                "name": "amount",
                "type": "uint256"
            }
        ],
        "name": "repay",
        "outputs": [

        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "bytes32",
                "name": "role",
                "type": "bytes32"
            },
            {
                "internalType": "address",
                "name": "account",
                "type": "address"
            }
        ],
        "name": "revokeRole",
        "outputs": [

        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "string",
                "name": "paramName",
                "type": "string"
            },
            {
                "internalType": "uint256",
                "name": "newValue",
                "type": "uint256"
            }
        ],
        "name": "setParams",
        "outputs": [

        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "stETH",
        "outputs": [
            {
                "internalType": "contract IMiniStETH",
                "name": "",
                "type": "address"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "bytes4",
                "name": "interfaceId",
                "type": "bytes4"
            }
        ],
        "name": "supportsInterface",
        "outputs": [
            {
                "internalType": "bool",
                "name": "",
                "type": "bool"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "totalInterestForStETH",
        "outputs": [
            {
                "internalType": "uint256",
                "name": "",
                "type": "uint256"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "treasury",
        "outputs": [
            {
                "internalType": "contract IMiniTreasury",
                "name": "",
                "type": "address"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [

        ],
        "name": "unpause",
        "outputs": [

        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "string",
                "name": "contractName",
                "type": "string"
            },
            {
                "internalType": "address",
                "name": "newAddress",
                "type": "address"
            }
        ],
        "name": "updateContract",
        "outputs": [

        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "uint256",
                "name": "amount",
                "type": "uint256"
            }
        ],
        "name": "withdrawCollateral",
        "outputs": [

        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "stateMutability": "payable",
        "type": "receive"
    }
]`

// DepositEvent DepositETH事件数据结构
type DepositEvent struct {
	User        common.Address
	Amount      *big.Int
	StETHAmount *big.Int
}

// DepositCollateralEvent 质押事件数据结构
type DepositCollateralEvent struct {
	User   common.Address
	Amount *big.Int
}

// BorrowEvent 借款事件数据结构
type BorrowEvent struct {
	User       common.Address
	MUSDAmount *big.Int
	Fee        *big.Int
}

// RepayEvent 还款事件数据结构
type RepayEvent struct {
	User   common.Address
	Amount *big.Int
}

// LiquidateEvent 清算事件数据结构
type LiquidateEvent struct {
	Liquidator       common.Address
	Borrower         common.Address
	DebtRepaid       *big.Int
	CollateralSeized *big.Int
	Fee              *big.Int
}

// WithdrawCollateralEvent 提取质押物事件数据结构
type WithdrawCollateralEvent struct {
	User        common.Address
	Amount      *big.Int
	ETHReceived *big.Int
}

var miniVaultABI abi.ABI

func init() {
	var err error
	miniVaultABI, err = abi.JSON(strings.NewReader(MiniVaultABI))
	if err != nil {
		panic("Failed to parse MiniVault ABI: " + err.Error())
	}
}

// ParseDepositEvent 解析DepositETH事件日志
func ParseDepositEvent(logData []byte, topics []common.Hash) (*DepositEvent, error) {
	var event DepositEvent

	if len(topics) < 2 {
		return nil, fmt.Errorf("insufficient topics for DepositETH event")
	}

	event.User = common.BytesToAddress(topics[1].Bytes())

	// 解析数据部分 (amount, stETHAmount)
	values, err := miniVaultABI.Unpack("DepositETH", logData)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack DepositETH event: %v", err)
	}

	if len(values) >= 2 {
		event.Amount = values[0].(*big.Int)
		event.StETHAmount = values[1].(*big.Int)
	}

	return &event, nil
}

// ParseBorrowEvent 解析借款事件日志
func ParseBorrowEvent(logData []byte, topics []common.Hash) (*BorrowEvent, error) {
	var event BorrowEvent

	if len(topics) < 2 {
		return nil, fmt.Errorf("insufficient topics for Borrow event")
	}

	event.User = common.BytesToAddress(topics[1].Bytes())

	// 解析数据部分 (mUSDAmount, fee)
	values, err := miniVaultABI.Unpack("Borrow", logData)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack Borrow event: %v", err)
	}

	if len(values) >= 2 {
		event.MUSDAmount = values[0].(*big.Int)
		event.Fee = values[1].(*big.Int)
	}

	return &event, nil
}

// ParseRepayEvent 解析还款事件日志
func ParseRepayEvent(logData []byte, topics []common.Hash) (*RepayEvent, error) {
	var event RepayEvent

	if len(topics) < 2 {
		return nil, fmt.Errorf("insufficient topics for Repay event")
	}

	event.User = common.BytesToAddress(topics[1].Bytes())

	// 解析数据部分 (amount)
	values, err := miniVaultABI.Unpack("Repay", logData)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack Repay event: %v", err)
	}

	if len(values) >= 1 {
		event.Amount = values[0].(*big.Int)
	}

	return &event, nil
}

// ParseLiquidateEvent 解析清算事件日志
func ParseLiquidateEvent(logData []byte, topics []common.Hash) (*LiquidateEvent, error) {
	var event LiquidateEvent

	if len(topics) < 3 {
		return nil, fmt.Errorf("insufficient topics for Liquidate event")
	}

	event.Liquidator = common.BytesToAddress(topics[1].Bytes())
	event.Borrower = common.BytesToAddress(topics[2].Bytes())

	// 解析数据部分 (debtRepaid, collateralSeized, fee)
	values, err := miniVaultABI.Unpack("Liquidate", logData)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack Liquidate event: %v", err)
	}

	if len(values) >= 3 {
		event.DebtRepaid = values[0].(*big.Int)
		event.CollateralSeized = values[1].(*big.Int)
		event.Fee = values[2].(*big.Int)
	}

	return &event, nil
}

// ParseDepositCollateralEvent 解析质押事件日志
func ParseDepositCollateralEvent(logData []byte, topics []common.Hash) (*DepositCollateralEvent, error) {
	var event DepositCollateralEvent

	if len(topics) < 2 {
		return nil, fmt.Errorf("insufficient topics for DepositCollateral event")
	}

	event.User = common.BytesToAddress(topics[1].Bytes())

	// 解析数据部分 (amount)
	values, err := miniVaultABI.Unpack("DepositCollateral", logData)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack DepositCollateral event: %v", err)
	}

	if len(values) >= 1 {
		event.Amount = values[0].(*big.Int)
	}

	return &event, nil
}

// ParseWithdrawCollateralEvent 解析提取质押物事件日志
func ParseWithdrawCollateralEvent(logData []byte, topics []common.Hash) (*WithdrawCollateralEvent, error) {
	var event WithdrawCollateralEvent

	if len(topics) < 2 {
		return nil, fmt.Errorf("insufficient topics for WithdrawCollateral event")
	}

	event.User = common.BytesToAddress(topics[1].Bytes())

	// 解析数据部分 (amount, ethReceived)
	values, err := miniVaultABI.Unpack("WithdrawCollateral", logData)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack WithdrawCollateral event: %v", err)
	}

	if len(values) >= 2 {
		event.Amount = values[0].(*big.Int)
		event.ETHReceived = values[1].(*big.Int)
	}

	return &event, nil
}

// GetEventTopic 获取事件签名对应的topic
func GetEventTopic(eventName string) common.Hash {
	return miniVaultABI.Events[eventName].ID
}
