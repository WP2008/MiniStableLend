// SPDX-License-Identifier: MIT

pragma solidity ^0.8.20;

interface IMiniStETH {
    function mint(address to, uint256 amount) external;
    function burn(address from, uint256 amount) external;

    function exchangeRate() external view returns (uint256);
    function updateExchangeRate(uint256 newRate) external;
    function addStETHInterest(uint256 amount) external;

    function totalSupply() external view returns (uint256);
    function transfer(address to, uint256 amount) external returns (bool);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);
    function allowance(address owner, address spender) external view returns (uint256);
    function balanceOf(address account) external view returns (uint256); 
}