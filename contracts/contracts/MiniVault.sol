// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;


import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "hardhat/console.sol";

import "./interface/IMiniMUSD.sol";
import "./interface/IMiniStETH.sol";
import "./interface/IMiniOracle.sol";
import "./interface/IMiniTreasury.sol";

contract MiniVault is ReentrancyGuard, AccessControl {
    // ============ 常量定义 ============
    bytes32 public constant TIMELOCK_ROLE = keccak256("TIMELOCK_ROLE");
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");
    uint256 public constant BASE = 1e18;                    // 基础精度
    uint256 public constant LIQUIDATION_DISCOUNT = 5e16;    // 清算折扣 5%
    uint256 public constant LIQUIDATION_MAX_PCT = 5e17;     // 单次最大清算比例 50%
    uint256 public constant LIQUIDATION_TREASURY_FEE = 1e16; // 清算国库手续费 1%
    uint256 public constant INTEREST_TREASURY_SHARE = 1e17;  // 利息国库分成 10%
    uint256 public constant RATE_UPDATE_INTERVAL = 24 hours; // stETH 兑换率更新间隔

    uint256 public borrowLimitPct;        // 借贷比例
    uint256 public liquidationThreshold;  // 清算阈值
    uint256 public borrowFeePct;          // 铸币手续费
    uint256 public borrowAPR;             // 借贷年化利率

    uint256 public totalInterestForStETH; // 待分配给 stETH 持有者的利息 （stETH 数量 18位小数）
    uint256 public lastRateUpdate;        // 最后一次更新 ETH -> stETH 兑换率时间

    bool public paused;  // 合约暂停状态


    IMiniStETH public stETH;   // 质押代币
    IMiniMUSD public mUSD;     // 稳定币合约
    IMiniOracle public oracle; // 价格预言机
    IMiniTreasury public treasury;  // 项目方金库

    // 用户的持仓
    struct Position {
        uint256 collateral; // 质押代币 stEth 数量
        uint256 debt;       // 债务 mUSD + 利息
        uint256 initialDebt;// 初始借款本金
        uint256 lastUpdate; // 最后计息时间
    }

    // 用户地址 => 持仓
    mapping (address => Position) public positions;

    // ========= event define =========
    /**
     * @dev 用户存入 ETH 事件
     * @param user deposit ETH 的用户地址
     * @param amount 用户存入的 ETH 数量
     * @param stETHAmount 用户获得的 stETH 数量
     */
    event DepositETH(address indexed user, uint256 amount, uint256 stETHAmount);  

    /**
     * @dev 用户质押 stETH 事件
     * @param user 调用 depositCollateral 的用户地址
     * @param amount 用户质押的 stETH 数量
     */
    event DepositCollateral(address indexed user, uint256 amount);

    /**
     * @dev 用户借出 mUSD
     * @param user 用户地址
     * @param mUSDAmount 借出的 mUSD 数量
     * @param fee 给项目方的 mUSD 费用数量
     */
    event Borrow(address indexed user, uint256 mUSDAmount, uint256 fee);

    /**
     * @dev 用户还回 mUSD
     * @param user 用户地址 
     * @param amount 还回的 mUSD 数量
     */
    event Repay(address indexed user, uint256 amount);

    /**
     * @dev 清算事件
     * @param liquidator 清算人地址
     * @param borrower 被清算的用户地址
     * @param debtRepaid 偿还的债务数量
     * @param collateralSeized 被清算获得的质押物数量 stETH
     * @param fee 清算手续费数量 stETH
     */
    event Liquidate(address indexed liquidator, address indexed borrower, uint256 debtRepaid, uint256 collateralSeized, uint256 fee);

    /**
     * @dev 用户提取质押物事件
     * @param user 用户地址
     * @param amount 提取的 stETH 数量
     * @param ethReceived 用户实际收到的 ETH 数量
     */
    event WithdrawCollateral(address indexed user, uint256 amount, uint256 ethReceived);

    /**
     * @dev 参数更新事件
     * @param paramName 更新的参数名称
     * @param oldValue 旧的参数值
     * @param newValue 新的参数值
     */
    event ParamsUpdated(string paramName, uint256 oldValue, uint256 newValue);

    /**
     * @dev 利息分配事件
     * @param toStETH 分配给 stETH 持有者的利息数量
     * @param toTreasury 分配给国库的利息数量（以 mUSD 计价）
     */
    event InterestDistributed(uint256 toStETH, uint256 toTreasury);
    /**
     * @dev 兑换率更新事件
     * @param oldRate 旧的 stETH 兑换率
     * @param newRate 新的 stETH 兑换率
     */
    event ExchangeRateUpdated(uint256 oldRate, uint256 newRate);

    /**
     * @dev 合约暂停状态变更事件
     */
    event Paused(address indexed account);
    event Unpaused(address indexed account);

    /**
     * @dev 修饰器：合约未暂停时才可执行
     */
    modifier whenNotPaused() {
        require(!paused, "Contract is paused");
        _;
    }

    constructor(
        address _stETH,
        address _mUSD,
        address _oracle,
        address _treasury,
        address _timelock

    ) {
        // 绑定核心合约
        stETH = IMiniStETH(_stETH);
        mUSD = IMiniMUSD(_mUSD);
        oracle = IMiniOracle(_oracle);
        treasury = IMiniTreasury(_treasury);

        // 初始化参数
        borrowLimitPct = 6660e14;          // 借贷比例 66.6%
        liquidationThreshold = 11000e14;   // 清算阈值 110 % 
        borrowFeePct = 5e14;               // 铸币手续费 0.05 %
        borrowAPR = 400e14;                // 借贷年化利率 4%


        // 初始化兑换率更新时间
        lastRateUpdate = block.timestamp;

        // 授予 Timelock 角色权限
        _grantRole(TIMELOCK_ROLE, _timelock);
        _grantRole(PAUSER_ROLE, _timelock);

        // TODO 取消 DEFAULT_ADMIN_ROLE 权限有必要吗，他能做什么
         // 将部署者的权限撤销（仅保留 Timelock 权限）
        // _revokeRole(DEFAULT_ADMIN_ROLE, msg.sender);
    }

    // ========= internal =========

    /**
     * @notice 获取当前价格数据（Gas 优化：一次性获取）
     * @return ethPrice ETH/USD 价格（8位小数）
     * @return stETHRate stETH 兑换率（18位小数）
     */
    function _getPrices() internal view returns (uint256 ethPrice, uint256 stETHRate) {
        ethPrice = oracle.getETHPrice();
        stETHRate = stETH.exchangeRate();
    }

    /**
     * @notice 计算 stETH 数量对应的 USD 价值
     * @param stETHAmount stETH 数量（18位小数）
     * @return USD 价值（8位小数）
     */
    function _stETHToUSD(uint256 stETHAmount) internal view returns (uint256) {
        (uint256 ethPrice, uint256 stETHRate) = _getPrices();
        return (stETHAmount * stETHRate * ethPrice) / (BASE * BASE);
    }

    /**
     * @notice 计算 mUSD 数量对应的 stETH 数量
     * @param mUSDAmount mUSD 数量（8位小数）
     * @return stETH 数量（18位小数）
     */
    function _mUSDToStETH(uint256 mUSDAmount) internal view returns (uint256) {
        (uint256 ethPrice, uint256 stETHRate) = _getPrices();
        return (mUSDAmount * BASE * BASE) / (ethPrice * stETHRate);
    }

    /**
     * @dev 计算并更新用户的利息
     * @param user 需要计息的用户地址
     */
    function _accrueInterest(address user) internal {
        Position storage position = positions[user];

        // 如果没有债务，则不需要计息
        if (position.debt == 0) {
            return;
        }

        uint256 currentTime = block.timestamp;
        // 计算从上次更新到现在的时间差，并根据年化利率计算应计利息
        uint256 timeElapsed = currentTime - position.lastUpdate;
        if (timeElapsed == 0) {
            return;
        }
        
        // 计算利息：债务 * 年化利率 * 时间差 / 一年秒数
        uint256 interest = (position.debt * borrowAPR * timeElapsed) / (365 days * BASE);
        position.debt += interest;
        position.lastUpdate = currentTime;
    }

    /**
     * @notice 计算用户健康因子（抵押率）
     * @dev 健康因子 = (抵押品价值 * 1e18) / 债务价值
     * @param user 用户地址
     * @return 健康因子（1e18 精度）
     */
    function _getHealthFactor(address user) internal view returns (uint256) {
        Position storage position = positions[user];
        if (position.debt == 0) {
            return type(uint256).max; // 无债务
        }

        // 使用公共方法计算抵押品价值
        uint256 collateralValueForUSD = _stETHToUSD(position.collateral);

        // 计算健康因子：质押物价值 / 债务
        return (collateralValueForUSD * BASE) / position.debt;
    }


    /**
     * @dev 计算用户的可借额度
     * @param user 用户地址
     */
    function _getBorrowLimit(address user) internal view returns (uint256) {
        Position storage pos = positions[user];
        if (pos.collateral == 0) {
            return 0;
        }

        // 使用公共方法计算抵押品价值
        uint256 collateralValueForUSD = _stETHToUSD(pos.collateral);

        uint256 maxBorrow = (collateralValueForUSD * borrowLimitPct) / BASE;

        return maxBorrow > pos.debt ? maxBorrow - pos.debt : 0;
    }

    /**
     * @notice 分配实际偿还的利息收益
     * @dev 仅在用户还款/清算时调用，拆分利息为 90%（stETH）+ 10%（国库）
     * @param interestRepaid 本次偿还的利息, mUSD
     */
    function _distributeInterest(uint256 interestRepaid) internal returns (uint256, uint256) {
        if (interestRepaid == 0) return (0, 0);

        // 1. 拆分利息 90% 给 stETH 持有人， 10% 进入国库
        uint256 interestForTreasury = interestRepaid * INTEREST_TREASURY_SHARE / BASE; // 10%
        uint256 interestForStETH = interestRepaid - interestForTreasury; // 90% 

        // 2. 国库收益，直接将 mUSD 转入. （用户偿还的 mUSD 已经转入了 Vault 合约）
        if (interestForTreasury > 0) {
            mUSD.approve(address(treasury), interestForTreasury);
            treasury.receiveFee(address(mUSD), interestForTreasury);
            emit InterestDistributed(interestForStETH, interestForTreasury);
        }

        // 3. stETH 收益，转化为stETH 累计到 totalInterestForStETH
        if (interestForStETH > 0) {
            // 价值转换 mUSD 到 stETH
            uint256 interestStETH = _mUSDToStETH(interestForStETH);
            totalInterestForStETH += interestStETH;
        }

        // 4. 定期更新 stETH 兑换率（每 24 小时一次）， 通过更新兑换比例的方式，分配利息收益
        if (block.timestamp - lastRateUpdate >= RATE_UPDATE_INTERVAL && totalInterestForStETH > 0) {
            _updateStETHExchangeRate();
        }

        return (interestForStETH, interestForTreasury);
    }

    function _updateStETHExchangeRate() internal {
        uint256 totalStETH = stETH.totalSupply();
        if (totalStETH == 0 || totalInterestForStETH == 0) return;

           
        // 计算兑换率增量：累计利息 ETH / stETH 总供应量
        uint256 rateIncrement = (totalInterestForStETH * BASE) / totalStETH;

        uint256 oldRate = stETH.exchangeRate();
        uint256 newExchangeRate = oldRate + rateIncrement;
        
        // 更新 stETH 兑换率
        stETH.updateExchangeRate(newExchangeRate);
        stETH.addStETHInterest(totalInterestForStETH);
        
        // 重置累计利息和更新时间
        totalInterestForStETH = 0;
        lastRateUpdate = block.timestamp;

        emit ExchangeRateUpdated(oldRate, newExchangeRate);
    }

    // ========= external =========

    /**
     * @dev 用户存入 ETH，铸造 stETH 给用户
     * @notice 当前设定 1:1 锚定，后续接入实际质押逻辑
     */
    function depositETH() external payable nonReentrant whenNotPaused {
        require(msg.value > 0, "Deposit eth amount must > 0");

        // 1. 获取 stETH 实时兑换率（1 stETH = X ETH）
        uint256 stETHRate = stETH.exchangeRate();
        require(stETHRate >= 1, "Invalid stETH exchange rate, must be >= 1");
       
        // 2. 计算应铸造的 stETH 数量 = 存入 ETH 数量 × 1e18 / 兑换率
        uint256 stETHToMint = msg.value * BASE / stETHRate;
        require(stETHToMint > 0, "stETH mint amount zero");

        // 铸造 stETH 给用户
        stETH.mint(msg.sender, stETHToMint);
        emit DepositETH(msg.sender, msg.value, stETHToMint);
    }

    /**
    * @notice 用户质押 stETH 作为抵押品
     * @dev 需要用户先授权 stETH 给合约
     * @param amount 用户质押的 stETH 数量
     */
    function depositCollateral(uint256 amount) external nonReentrant whenNotPaused {
        require(amount > 0, "Amount must be > 0");

        // 1. 转移 stETH 到 Vault（Gas 优化：移除冗余检查）
        // transferFrom 会自动检查授权和余额，失败时 revert
        bool success = stETH.transferFrom(msg.sender, address(this), amount);
        require(success, "Transfer stETH failed");

        // 2. 更新用户持仓
        positions[msg.sender].collateral += amount;

        emit DepositCollateral(msg.sender, amount);
    }

    /**
     * @notice 用户铸造 mUSD（借贷）
     * @dev 校验抵押率，收取手续费，更新债务
     * @param mUSDAmount 要铸造的 mUSD 数量
     */
    function borrow(uint256 mUSDAmount) external nonReentrant whenNotPaused {
        require(mUSDAmount > 0, "Borrow amount must be > 0");
        // 1. 累计用户的利息
        _accrueInterest(msg.sender);

        // 2. 检查可借额度
        uint256 borrowLimit = _getBorrowLimit(msg.sender);
        require(borrowLimit >= mUSDAmount, "Exceeds borrow limit");

        // 3. 计算铸币费用
        uint256 fee = mUSDAmount * borrowFeePct / BASE;

        // 4. 铸造 mUSD 用户收到 mUSDAmount - fee，手续费直接进入国库
        mUSD.mint(msg.sender, mUSDAmount - fee);
        mUSD.mint(address(treasury), fee);

        // 5. 更新持仓信息
        Position storage pos = positions[msg.sender];
        pos.initialDebt += mUSDAmount; // 债务的本金
        pos.debt += mUSDAmount;  // 债务 本金+利息
        pos.lastUpdate = block.timestamp;

        emit Borrow(msg.sender, mUSDAmount, fee);
    }

    /**
     * @notice 用户归还 mUSD 债务
     * @dev 用户需要先授权 mUSD 给合约
     * @param amount 归还的 mUSD 数量
     */
    function repay(uint256 amount) external nonReentrant { // 还款不暂停，允许用户还款
        require(amount > 0, "Repay amount must be > 0");

        // 1. 计算利息
        _accrueInterest(msg.sender);
        Position storage pos = positions[msg.sender];

        // 2. 计算本金和利息
        uint256 interestOwed = pos.debt - pos.initialDebt; // 累计未还利息
        uint256 principalToRepay;
        uint256 interestToRepay;
        
        if (amount <= interestOwed) {
            // 只够还利息
            interestToRepay = amount;
            principalToRepay = 0;
        } else {
            interestToRepay = interestOwed;
            principalToRepay = amount - interestOwed;

            if (principalToRepay > pos.initialDebt) {
                principalToRepay = pos.initialDebt;
            }
        }
        
        // 3. 转移用户的 mUSD 到合约
        bool success = mUSD.transferFrom(msg.sender, address(this), amount);
        require(success, "mUSD tranferFrom failed");
        // 4. 分配利息收益, 定期更新兑换比例，结算 stETH 持有者利息收益
        (, uint256 toTreasury) = _distributeInterest(interestToRepay);

        // 5. 销毁 mUSD（扣除国库收益后的剩余部分）
        uint256 amountToBurn = amount - toTreasury;

        if (amountToBurn > 0) {
            mUSD.burn(address(this), amountToBurn);
        }

        // 6. 更新用户持仓
        pos.debt -= (interestToRepay + principalToRepay);
        pos.initialDebt -= principalToRepay;

        emit Repay(msg.sender, amount);
    }

    
    /**
     * @notice 清算用户仓位
     * @dev 清算人用 mUSD 偿还债务，获得折扣 stETH
     * @param borrower 被清算用户地址
     * @param debtRepaid 要偿还的债务金额
     */
    function liquidate(address borrower, uint256 debtRepaid) external nonReentrant {
        require(borrower != address(0), "Invalid borrower");
        require(debtRepaid > 0, "Debt amount must be > 0");

        // 1. 计算借款人债务
        _accrueInterest(borrower);

        // 2. 确认是否可以平仓  健康因子 < 阈值
        uint256 healthFactor =  _getHealthFactor(borrower);
        require(healthFactor < liquidationThreshold, "Not liquidatable");

        // 3. 安比例清算, 最高 LIQUIDATION_MAX_PCT
        Position storage pos = positions[borrower];
        uint256 maxLiquidate = pos.debt * LIQUIDATION_MAX_PCT / BASE;
        if (debtRepaid > maxLiquidate) {
            debtRepaid = maxLiquidate;
        }

        // 4. 计算本次清算中的本金和利息
        uint256 interestOwed = pos.debt - pos.initialDebt;
        uint256 principalToRepay;
        uint256 interestToRepay;

        if (debtRepaid <= interestOwed) {
            interestToRepay = debtRepaid;
            principalToRepay = 0;
        } else {
            interestToRepay = interestOwed;
            principalToRepay = debtRepaid - interestOwed;
            if (principalToRepay > pos.initialDebt) {
                principalToRepay = pos.initialDebt;
            }
        }

        // 5. 计算质押代币的数量
        // 抵押品 stETH = 偿还债务金额 / (ETH 价格 * 兑换率 * (1 - 清算折扣))
        // 价值转化 mUSD 到 stETH
        uint256 collateralValueStETH = _mUSDToStETH(debtRepaid);
        // 应用清算折扣
        uint256 collateralSeized = (collateralValueStETH * BASE) / (BASE - LIQUIDATION_DISCOUNT);

        // 6. 校验 borrower 的质押是否足够
        require(collateralSeized > 0, "Collateral amount must be > 0");
        require(pos.collateral >= collateralSeized, "Collateral exceeds borrower's collateral");

        // 7. 转移 mUSD 并销毁
        bool success = mUSD.transferFrom(msg.sender, address(this), debtRepaid);
        require(success, "mUSD transfer failed");

        // 8. 分配利息收益
        (, uint256 toTreasury) = _distributeInterest(interestToRepay);

        // 9. 销毁 mUSD（扣除国库收益后的剩余部分）
        uint256 amountToBurn = debtRepaid - toTreasury;
        if (amountToBurn > 0) {
            mUSD.burn(address(this), amountToBurn);
        }
        
        // 10. 计算清算手续费 stETH
        uint256 treasuryFee = collateralSeized * LIQUIDATION_TREASURY_FEE / BASE;
        uint256 liquidatorCollateral = collateralSeized - treasuryFee;


        // 11. 转移质押 stETH， 清算人获得 4% 折扣部分，国库 1% 手续费
        stETH.transfer(msg.sender, liquidatorCollateral);

        // 授权给 treasury 转账
        stETH.approve(address(treasury), treasuryFee);
        treasury.receiveFee(address(stETH), treasuryFee);

        // 12. 更新借款人状态
        pos.debt -= debtRepaid;
        pos.initialDebt -= principalToRepay;
        pos.collateral -= collateralSeized;
        pos.lastUpdate = block.timestamp;

        emit Liquidate(msg.sender, borrower, debtRepaid, collateralSeized, treasuryFee);
    }   

    /**
     * @notice 用户提取质押物 ETH
     * @dev 只能提取不导致仓位被清算的 stETH 数量
     * @param amount 用户想要提取的 stETH 数量 (18 位小数)
     */
    function withdrawCollateral(uint256 amount) external nonReentrant whenNotPaused {
        require(amount > 0, "Withdraw amount must be > 0");

        Position storage pos = positions[msg.sender];

        // 1. 计算利息
        _accrueInterest(msg.sender);

        // 2. 校验可取最大 stETH 数量
        if (pos.debt != 0) {
            // 使用公共方法计算债务对应的 stETH 数量（在清算阈值下）
            uint256 debtToStETHValue = _mUSDToStETH(pos.debt);
            uint256 minRequired = (debtToStETHValue * liquidationThreshold) / BASE;

            // 确保提取后不会低于最小要求
            if (pos.collateral - amount < minRequired) {
                amount = pos.collateral - minRequired;
            }
            require(amount > 0, "Cannot withdraw with outstanding debt");
        } else {
            // 最多可提取的数量
            if (pos.collateral < amount) {
                amount = pos.collateral;
            }
        }

        // 3. 更新用户抵押品数量
        pos.collateral -= amount;

        // 4. 销毁 stETH， 按当前兑换率计算应返还的 ETH 数量
        stETH.burn(address(this), amount);

        uint256 exchangeRate = stETH.exchangeRate();
        uint256 ethToWithdraw = amount * exchangeRate / BASE;

        // 5. 向用户发送 ETH
        require(address(this).balance >= ethToWithdraw, "Insufficient ETH in vault");
        (bool success, ) = payable(msg.sender).call{value: ethToWithdraw}("");
        require(success, "ETH transfer failed");

        emit WithdrawCollateral(msg.sender, amount, ethToWithdraw);
    }

    // ============ 权限控制与参数配置函数 ============

    /**
     * @notice 修改全局参数（仅 Timelock 可调用）
     * @dev 用于调整抵押率、费率、利率等核心参数
     * @param paramName 参数名称
     * @param newValue 新参数值
     */
    function setParams(string calldata paramName, uint256 newValue) external onlyRole(TIMELOCK_ROLE) {  
        uint256 oldValue;
        
        if (keccak256(bytes(paramName)) == keccak256(bytes("borrowLimitPct"))) {
            oldValue = borrowLimitPct;
            borrowLimitPct = newValue;
        } else if (keccak256(bytes(paramName)) == keccak256(bytes("liquidationThreshold"))) {
            oldValue = liquidationThreshold;
            liquidationThreshold = newValue;
        } else if (keccak256(bytes(paramName)) == keccak256(bytes("borrowFeePct"))) {
            oldValue = borrowFeePct;
            borrowFeePct = newValue;
        } else if (keccak256(bytes(paramName)) == keccak256(bytes("borrowAPR"))) {
            oldValue = borrowAPR;
            borrowAPR = newValue;
        } else {
            revert("Invalid parameter name");
        }
        
        emit ParamsUpdated(paramName, oldValue, newValue);
    }

    /**
     * @notice 暂停合约（仅 Pauser 可调用）
     * @dev 暂停后，存款、借款、提款功能将停止
     */
    function pause() external onlyRole(PAUSER_ROLE) {
        paused = true;
        emit Paused(msg.sender);
    }

    /**
     * @notice 恢复合约（仅 Pauser 可调用）
     */
    function unpause() external onlyRole(PAUSER_ROLE) {
        paused = false;
        emit Unpaused(msg.sender);
    }

    /**
     * @notice 紧急提取（仅 Timelock 可调用）
     * @dev 在合约暂停或紧急情况下提取锁定资金
     * @param token 代币地址，address(0) 表示 ETH
     * @param amount 提取数量
     * @param to 接收地址
     */
    function emergencyWithdraw(address token, uint256 amount, address to) external onlyRole(TIMELOCK_ROLE) {
        require(paused, "Contract must be paused");
        require(to != address(0), "Invalid recipient");

        if (token == address(0)) {
            // 提取 ETH
            require(address(this).balance >= amount, "Insufficient ETH balance");
            (bool success, ) = payable(to).call{value: amount}("");
            require(success, "ETH transfer failed");
        } else {
            // 提取 ERC20 代币
            IERC20(token).transfer(to, amount);
        }
    }

    /**
     * @notice 更新合约引用（仅 Timelock 可调用）
     * @dev 用于更换 oracle/treasury 等合约地址
     * @param contractName 合约名称
     * @param newAddress 新合约地址
     */
    function updateContract(string calldata contractName, address newAddress) external onlyRole(TIMELOCK_ROLE) {
        require(newAddress != address(0), "Invalid address");
        
        if (keccak256(bytes(contractName)) == keccak256(bytes("oracle"))) {
            oracle = IMiniOracle(newAddress);
        } else if (keccak256(bytes(contractName)) == keccak256(bytes("treasury"))) {
            treasury = IMiniTreasury(newAddress);
        } else {
            revert("Invalid contract name");
        }
    }

    // ----------- view functions -----------
    /**
     * @notice 获取用户的可借额度
     * @param user 用户地址
     * @return 可借额度数量
     */
    function getBorrowLimit(address user) external view returns (uint256) {
        return _getBorrowLimit(user);
    }

    /**
    * @notice 获取用户的健康因子
    * @param user 用户地址
    * @return 健康因子（1e18 精度）
    */
    function getHealthFactor(address user) external view returns (uint256) {
        return _getHealthFactor(user);
    }   

    /**
     * @notice 获取用户的持仓信息
     * @param user 用户地址
     * @return collateral 质押的 stETH 数量 
     * @return debt 债务的 mUSD 数量（包含利息）
     * @return initialDebt 初始借款本金数量
     * @return lastUpdate 上次更新时间
     */
    function getPosition(address user) external view returns (uint256 collateral, uint256 debt, uint256 initialDebt, uint256 lastUpdate) {
        Position storage pos = positions[user];
        return (pos.collateral, pos.debt, pos.initialDebt, pos.lastUpdate);
    }


    // ============ 接收 ETH 函数 ============
    // 拒绝所有直接的 ETH 转账，强制走 depositETH() 流程
    receive() external payable {
        revert("Direct ETH transfer not allowed, use depositETH()");
    }
}




