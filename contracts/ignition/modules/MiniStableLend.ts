import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("MiniStableLendModule", (m) => {
  
  // 部署 MiniMUSD 稳定币合约
  const miniMUSD = m.contract("MiniMUSD");

  // 获取Timelock地址
  const timelockAddress = m.getAccount(0);

  // 部署 MiniStETH 质押凭证合约
  const miniStETH = m.contract("MiniStETH", [timelockAddress]);
  

  // 获取当前网络名称（使用环境变量）
  const network = process.env.HARDHAT_NETWORK;
  console.log("Deploying to network:", network);


  // 判断是否为本地网络
  const isLocalNetwork = network === "hardhat" || network === "localhost";

  // 根据网络类型决定部署什么喂价合约
  let miniOracle;
  if (true) {
    // 在本地网络部署 MockAggregatorV3 作为喂价合约
     miniOracle = m.contract("MockAggregatorV3",  [
      3 * 10 * 8,
      8
    ]);
  } else {
    // 在其他网络使用模拟的 Chainlink 喂价合约地址
    miniOracle = m.contract("MiniOracle", [
      "0x5f4eC3Df9cbd43714FE2740f5E3616155c5b84164", // 模拟的 ETH/USD 喂价合约地址
      timelockAddress
    ]);
  }

  // 部署 MiniTreasury 国库合约
  const miniTreasury = m.contract("MiniTreasury", [timelockAddress]);

  // 部署 MiniVault 主合约
  const miniVault = m.contract("MiniVault", [
    miniStETH,
    miniMUSD,
    miniOracle,
    miniTreasury,
    timelockAddress // Timelock 地址
  ]);

  // 设置合约间的引用关系
  m.call(miniMUSD, "setVault", [miniVault]);
  m.call(miniStETH, "setVault", [miniVault]);

  return {
    miniMUSD,
    miniStETH,
    miniOracle,
    miniTreasury,
    miniVault
  };
});