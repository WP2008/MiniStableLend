// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/**
 * @title MiniTreasury
 * @dev 项目方金库，用于收集手续费和清算收益
 */
contract MiniTreasury is AccessControl {
    // 角色定义：仅Timelock可提取资金
    bytes32 public constant TIMELOCK_ROLE = keccak256("TIMELOCK_ROLE");

    // 收集的费用历史记录
    event FeeReceived(address indexed token, uint256 amount);
    event FeeWithdrawn(address indexed to, address indexed token, uint256 amount);

    constructor(address _timelock) {
        require(_timelock != address(0), "Timelock address cannot be zero");
        _grantRole(TIMELOCK_ROLE, _timelock);
    }

    /**
     * @notice 接收费用
     * @dev 可被 Vault 调用，直接通过 transfer 实现
     */
    receive() external payable {
        if (msg.value > 0) {
            emit FeeReceived(address(0), msg.value);
        }
    }

    /**
     * @notice 接收 ERC20 代币费用
     * @param token 代币地址
     * @param amount 代币数量
     */
    function receiveFee(address token, uint256 amount) external {
        require(amount > 0, "Amount must be > 0");
        require(token != address(0), "Token address cannot be zero");

        bool success = IERC20(token).transferFrom(msg.sender, address(this), amount);
        require(success, "Transfer failed");

        emit FeeReceived(token, amount);
    }

    /**
     * @notice 提取资金（仅 Timelock 可调用）
     * @param to 接收地址
     * @param token 代币地址，address(0) 表示 ETH
     * @param amount 提取数量
     */
    function withdraw(address to, address token, uint256 amount) external onlyRole(TIMELOCK_ROLE) {
        require(to != address(0), "Recipient cannot be zero");
        require(amount > 0, "Amount must be > 0");

        if (token == address(0)) {
            // 提取 ETH
            require(address(this).balance >= amount, "Insufficient ETH balance");
            (bool success, ) = payable(to).call{value: amount}("");
            require(success, "ETH transfer failed");
        } else {
            // 提取 ERC20 代币
            uint256 balance = IERC20(token).balanceOf(address(this));
            require(balance >= amount, "Insufficient token balance");
            bool success = IERC20(token).transfer(to, amount);
            require(success, "Token transfer failed");
        }

        emit FeeWithdrawn(to, token, amount);
    }

    /**
     * @notice 查询指定代币的余额
     * @param token 代币地址，address(0) 表示 ETH
     * @return 余额
     */
    function getBalance(address token) external view returns (uint256) {
        if (token == address(0)) {
            return address(this).balance;
        } else {
            return IERC20(token).balanceOf(address(this));
        }
    }
}
