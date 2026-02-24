// SPDX-License-Identifier: MIT

pragma solidity ^0.8.20;

interface IMiniOracle {
    function getETHPrice() external view returns (uint256);
}