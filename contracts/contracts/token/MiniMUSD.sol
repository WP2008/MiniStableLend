// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title MiniMUSD 
 * @dev 稳定币合约, 仅 MiniVault 可以 mint/burn
 */
contract MiniMUSD is ERC20 {

    address public immutable vault;

    constructor(address _vault) ERC20("Mini MUSD", "mUSD") {
        require(_vault != address(0), "Vault address cannot be zero");
        vault = _vault;
    }

    modifier onlyVault() {
        require(msg.sender == vault, "Only Vault");
        _;
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