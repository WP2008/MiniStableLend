# 精度转换说明文档

## 数据结构

| 数据类型 | 小数位数 | 说明 |
|---------|---------|------|
| BASE | 18 (1e18) | 基础精度常量 |
| ETH 价格 (Chainlink) | 8 | 1 ETH = X USD (8位小数) |
| stETH 兑换率 | 18 | 1 stETH = X ETH (18位小数) |
| mUSD | 18 | 稳定币，USD 表示 (18位小数) |
| stETH 数量 | 18 | 质押代币数量 (18位小数) |

## 核心转换公式

### 1. 抵押品 USD 价值计算

**公式：**
```solidity
collateralValueUSD = (stETH数量 × stETH兑换率 × ETH价格 × 1e10) / BASE
```

**推导过程：**
```
1. stETH → ETH: ETH数量 = stETH数量 / stETH兑换率
   = stETH数量 × 1e18 / stETH兑换率 × 1e18
   = (stETH数量 × 1e18) / stETH兑换率 × 1e18

2. ETH → USD: USD数量 = ETH数量 × ETH价格
   = (stETH数量 × 1e18) / stETH兑换率 × 1e18 × ETH价格 × 1e8

3. 统一为 18 位小数:
   USD数量(18位) = (stETH数量 × stETH兑换率 × ETH价格 × 1e10) / 1e18
```

**关键点：**
- ✅ 先进行所有乘法运算
- ✅ 最后一步才除法
- ❌ 不要先除后乘（会造成精度损失）

### 2. mUSD → stETH 转换

**公式：**
```solidity
stETH数量 = (mUSD数量 × 1e26) / (ETH价格 × stETH兑换率)
```

**推导过程：**
```
1. USD → ETH: ETH数量 = mUSD数量 / ETH价格
   = mUSD数量 × 1e18 / (ETH价格 × 1e8)
   = (mUSD数量 × 1e10) / ETH价格

2. ETH → stETH: stETH数量 = ETH数量 / stETH兑换率
   = [(mUSD数量 × 1e10) / ETH价格] / (stETH兑换率 × 1e18)
   = (mUSD数量 × 1e28) / (ETH价格 × stETH兑换率 × 1e18)
   = (mUSD数量 × 1e26) / (ETH价格 × stETH兑换率)
```

### 3. 清算抵押品计算

**公式（考虑折扣）：**
```solidity
stETH数量 = [(mUSD数量 × 1e26) / (ETH价格 × stETH兑换率) × BASE] / (BASE - LIQUIDATION_DISCOUNT)
```

### 4. 提款限制计算

**公式：**
```solidity
minRequiredStETH = [(debt × 1e26) / (ETH价格 × stETH兑换率) × liquidationThreshold] / BASE
```

## 代码示例

### 正确 ✅
```solidity
// 抵押品价值计算
uint256 collateralValueForUSD = (collateral * stETHRate * ethPrice * 1e10) / BASE;

// mUSD 转 stETH
uint256 stETHAmount = (mUSDAmount * 1e26) / (ethPrice * stETHRate);
```

### 错误 ❌
```solidity
// 错误1：先除后乘，精度丢失
uint256 collateralValueForUSD = (collateral * stETHRate * ethPrice) / BASE;
collateralValueForUSD = collateralValueForUSD * 1e10;  // ❌ 精度已丢失

// 错误2：分母溢出风险
uint256 ethValue = (mUSD * BASE) / (ethPrice * 1e10);  // ❌ 溢出风险
```

## Gas 优化建议

1. **减少 Oracle 调用**：缓存价格，避免重复查询
2. **使用 unchecked**：在确定不会溢出的算术运算中使用 unchecked 块
3. **合并计算**：如有可能，合并多个转换步骤

## 安全检查

- ✅ 所有除法前检查分母不为零
- ✅ 考虑整数溢出（使用 1e26 而非 1e36 避免溢出）
- ✅ 使用 Ray Math (1e27) 或 Wad Math (1e18) 库进行高精度计算

## 测试用例

```solidity
// 测试用例：验证精度
function testPrecision() public {
    uint256 ethPrice = 3000e8;  // $3000
    uint256 stETHRate = 1e18;      // 1:1
    uint256 mUSDAmount = 1000e18; // $1000

    // 预期：1000e18 / (3000e8 × 1e18) × 1e26 = 0.333...e18
    uint256 stETHAmount = (mUSDAmount * 1e26) / (ethPrice * stETHRate);
    assert(stETHAmount > 0);
}
```
