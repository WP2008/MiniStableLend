// 简单测试脚本，使用预部署的合约地址
import { network } from "hardhat";
import {expect} from "chai";

// 连接网络并获取 ethers 对象
const provider = await network.connect();
const { ethers } = provider;

async function main() {
  console.log("Starting simple test generator...");

  // 使用Hardhat本地网络默认地址
  const miniVaultAddress = "0x5FC8d32690cc91D4c39d9d3abcBD16989F875707";
  const miniMUSDAddress = "0x5FbDB2315678afecb367f032d93F642f64180aa3";
  const miniStETHAddress = "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512";
  const MiniOracleAddress = "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9";

  // 获取合约实例
  const MiniVault = await ethers.getContractAt("MiniVault", miniVaultAddress);
  const MiniMUSD = await ethers.getContractAt("MiniMUSD", miniMUSDAddress);
  const MiniStETH = await ethers.getContractAt("MiniStETH", miniStETHAddress);
  const MiniOracle = await ethers.getContractAt("MiniOracle", MiniOracleAddress);


  const code = await ethers.provider.getCode(miniVaultAddress);

  // 获取测试账户
  const [deployer, user1] = await ethers.getSigners();

  console.log("1. User deposits ETH");
  await (await MiniVault.connect(user1).depositETH({value: ethers.parseEther("1")})).wait();
  const stETHAmount = await MiniStETH.connect(user1).balanceOf(user1.address);
  console.log("stETH Amount: ", stETHAmount);

  console.log("2. User deposits stETH as collateral");
  await MiniStETH.connect(user1).approve(miniVaultAddress, stETHAmount);
  await MiniVault.connect(user1).depositCollateral(stETHAmount);
  var position = await MiniVault.getPosition(user1.address);
  console.log("position: ", position);

  console.log("3. User borrows mUSD");
  const borrowLimit = await MiniVault.getBorrowLimit(user1.address);
  console.log("borrowLimit: ", borrowLimit);

  await MiniVault.connect(user1).borrow(borrowLimit);
  const mUSDAmount = await MiniMUSD.balanceOf(user1.address);
  console.log("mUSDAmount: ", mUSDAmount);

  position = await MiniVault.getPosition(user1.address);
  console.log("position: ", position);

}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});