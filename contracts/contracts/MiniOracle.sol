// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

/**
 * @title MiniOracle
 * @dev 处理 Chainlink ETH/USD 喂价
 * @notice 
 */
contract MiniOracle is AccessControl {
    // 角色定义：仅Timelock可修改喂价地址
    bytes32 public constant TIMELOCK_ROLE = keccak256("TIMELOCK_ROLE");

    // Chainlink ETH/USD喂价合约
    AggregatorV3Interface public ethUsdFeed;

    // 价格小数位数 8位
    uint8 public constant DECIMALS = 8; 

    // 价格有效期  1小时
    uint256 public constant PRICE_VALID_WINDOW = 3600;


    constructor(address _ethUsdFeed, address _timelock) {
        require(_ethUsdFeed != address(0), "Feed address cannot be zero");
        require(_timelock != address(0), "Timelock address cannot be zero");
        
        ethUsdFeed = AggregatorV3Interface(_ethUsdFeed);
        _grantRole(TIMELOCK_ROLE, _timelock);
    }

    /**
     * @dev 设置新的ETH/USD喂价地址（仅Timelock可调用）
     * @param _ethUsdFeed 新的Chainlink喂价合约地址
     */
    function setFeed(address _ethUsdFeed) external onlyRole(TIMELOCK_ROLE) {
        require(_ethUsdFeed != address(0), "New feed address cannot be zero");
        
        ethUsdFeed = AggregatorV3Interface(_ethUsdFeed);
    }

    /**
     * @dev 获取最新的ETH/USD价格，包含有效性检查
     * @return price 当前ETH/USD价格，8位小数
     */
    function getETHPrice() public view returns (uint256) {
        (, int256 price, , uint256 updatedAt,) = ethUsdFeed.latestRoundData();
        require(price > 0, "Invalid price");
        require(updatedAt >= block.timestamp - PRICE_VALID_WINDOW, "Price expired");
        return uint256(price);
    }

    /**
     * @dev 将ETH数量转换为USD金额，基于当前ETH/USD价格
     */
    function ethToUsd(uint256 ethAmount) external view returns (uint256) {

        uint256 ethPrice = getETHPrice();
        //（10 ** 8）*（10 ** 18）/（10 ** 8） = 10 ** 18
        // ethPrice * ethPrice / (10 ** DECIMALS)  多乘了8个零要除掉
        return (ethAmount * ethPrice) / (10 ** DECIMALS);
    }
}