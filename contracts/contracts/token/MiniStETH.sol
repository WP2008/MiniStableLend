// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

/**
 * @title MiniStETH
 * @dev ETH 质押凭证，仅MiniValut 可以 mint/burn，Timelock 可以设置Vault地址
 * 
 */
contract MiniStETH is ERC20, AccessControl {

    // 角色定义：仅Timelock可设置Vault地址
    bytes32 public constant TIMELOCK_ROLE = keccak256("TIMELOCK_ROLE");

    // 业务合约地址：仅MiniVault可操作mint/burn
    address public vault; 

    // stETH 兑换比例  stETH = exchangeRate * ETH，18位小数
    uint256 public exchangeRate;

    // 累计分配给 stETH 持有者的利息（ETH 数量）
    uint256 public totalStETHInterest;
    
    
    event ExchangeRateUpdated(uint256 oldRate, uint256 newRate);


    constructor(address _timelock) ERC20("Mini Staked ETH", "stETH") {
        require(_timelock != address(0), "Timelock address cannot be zero");
        _grantRole(TIMELOCK_ROLE, _timelock);
        exchangeRate = 1e18;
    }

    modifier onlyVault() {
        require(msg.sender == vault, "Only Vault");
        _;
    }

    // 更新兑换率（仅 Vault 可调用，用于分配利息）
    function updateExchangeRate(uint256 newRate) external onlyVault {
        require(newRate >= exchangeRate, "Rate cannot decrease");
        emit ExchangeRateUpdated(exchangeRate, newRate);
        exchangeRate = newRate;
    }

    // 累计 stETH 利息（仅 Vault 可调用）
    function addStETHInterest(uint256 amount) external onlyVault {
        totalStETHInterest += amount;
    }


    function setVault(address _vault) external onlyRole(TIMELOCK_ROLE) {
        require(_vault != address(0), "Vault address cannot be zero");
        vault = _vault;
    }
    
    function mint(address to, uint256 amount) external onlyVault {
        require(to != address(0), "Recipient cannot be zero");
        _mint(to, amount);
    }

    function burn(address from, uint256 amount) external onlyVault {
        require(from != address(0), "From cannot be zero");
        _burn(from, amount);
    }
}