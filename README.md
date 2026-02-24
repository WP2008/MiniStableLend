# MiniLend-Stable
## 项目概述
MiniLend-Stable 是一套**小而精的 ETH 质押 + 借贷 + 稳定币**三合一去中心化金融系统，核心定位为「ETH 专用轻量金融协议」，仅支持 ETH 作为唯一抵押品，发行锚定 USD 的稳定币 mUSD，兼具质押生息、超额抵押借贷、清算兜底能力，全程无复杂治理，以「极简、安全、透明」为核心设计原则。

## 合约逻辑 /contracts

### 如何部署 sepolia

1. .env 中配置 SEPOLIA_RPC_URL， SEPOLIA_PRIVATE_KEY 

2. 部署 npx hardhat ignition 