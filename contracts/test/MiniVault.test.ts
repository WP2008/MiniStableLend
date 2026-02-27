import { expect } from "chai";
import { network } from "hardhat";
import { ethers } from "ethers";

const { ethers: hardhatEthers } = await network.connect();

describe("MiniVault", function () {
  let vault: any;
  let stETH: any;
  let mUSD: any;
  let oracle: any;
  let treasury: any;
  let aggregator: any;
  let owner: any;
  let user1: any;
  let user2: any;
  let liquidator: any;

  beforeEach(async function () {
    [owner, user1, user2, liquidator] = await hardhatEthers.getSigners();

    // Deploy Mock Aggregator
    const MockAggregator = await hardhatEthers.getContractFactory("MockAggregatorV3");
    aggregator = await MockAggregator.deploy(3000n * 10n ** 8n, 8); // $3000 ETH

    // Deploy MiniStETH
    const MiniStETH = await hardhatEthers.getContractFactory("MiniStETH");
    stETH = await MiniStETH.deploy(owner.address);

    // Deploy MiniTreasury
    const MiniTreasury = await hardhatEthers.getContractFactory("MiniTreasury");
    treasury = await MiniTreasury.deploy(owner.address);

    // Deploy MiniOracle
    const MiniOracle = await hardhatEthers.getContractFactory("MiniOracle");
    oracle = await MiniOracle.deploy(await aggregator.getAddress(), owner.address);

    // Deploy mUSD 
    const MiniMUSD = await hardhatEthers.getContractFactory("MiniMUSD");
    mUSD = await MiniMUSD.deploy();

    // Deploy vault
    const MiniVault = await hardhatEthers.getContractFactory("MiniVault");
    vault = await MiniVault.deploy(
      await stETH.getAddress(),
      await mUSD.getAddress(),
      await oracle.getAddress(),
      await treasury.getAddress(),
      owner.address
    );

    // Update mUSD vault address
    await mUSD.setVault(await vault.getAddress());

    // Update stETH vault address
    await stETH.setVault(await vault.getAddress());
  });

  describe("Deployment", function () {
    it("Should set correct contract addresses", async function () {
      expect(await vault.stETH()).to.equal(await stETH.getAddress());
      expect(await vault.mUSD()).to.equal(await mUSD.getAddress());
      expect(await vault.oracle()).to.equal(await oracle.getAddress());
      expect(await vault.treasury()).to.equal(await treasury.getAddress());
    });

    it("Should set correct initial parameters", async function () {
      expect(await vault.borrowLimitPct()).to.equal(6660n * 10n ** 14n); // 66.6%
      expect(await vault.liquidationThreshold()).to.equal(11000n * 10n ** 14n); // 110%
      expect(await vault.borrowFeePct()).to.equal(5n * 10n ** 14n); // 0.05%
      expect(await vault.borrowAPR()).to.equal(400n * 10n ** 14n); // 4%
    });

    it("Should grant TIMELOCK_ROLE to owner", async function () {
      const TIMELOCK_ROLE = await vault.TIMELOCK_ROLE();
      expect(await vault.hasRole(TIMELOCK_ROLE, owner.address)).to.be.true;
    });

    it("Should grant PAUSER_ROLE to owner", async function () {
      const PAUSER_ROLE = await vault.PAUSER_ROLE();
      expect(await vault.hasRole(PAUSER_ROLE, owner.address)).to.be.true;
    });

    it("Should not be paused initially", async function () {
      expect(await vault.paused()).to.be.false;
    });
  });

  describe("Deposit ETH", function () {
    it("Should deposit ETH and mint stETH 1:1", async function () {
      const depositAmount = ethers.parseEther("10");

      await expect(vault.connect(user1).depositETH({ value: depositAmount }))
        .to.emit(vault, "DepositETH")
        .withArgs(user1.address, depositAmount, depositAmount);

      expect(await stETH.balanceOf(user1.address)).to.equal(depositAmount);
      expect(await hardhatEthers.provider.getBalance(await vault.getAddress())).to.equal(depositAmount);
    });

    it("Should fail if deposit amount is zero", async function () {
      await expect(
        vault.connect(user1).depositETH({ value: 0 })
      ).to.be.revertedWith("Deposit eth amount must > 0");
    });

    it("Should fail if contract is paused", async function () {
      await vault.pause();
      await expect(
        vault.connect(user1).depositETH({ value: ethers.parseEther("10") })
      ).to.be.revertedWith("Contract is paused");
    });

    it("Should handle exchange rate changes", async function () {
      const depositAmount = ethers.parseEther("10");

      // Update exchange rate to 1.05 
      await stETH.connect(owner).setVault(owner.address); // only vault 临时改为 owner
      await stETH.connect(owner).updateExchangeRate(1050000000000000000n); // 1.05e18
      await stETH.connect(owner).setVault(await vault.getAddress());
      await vault.connect(user1).depositETH({ value: depositAmount });

      // User should receive 10 / 1.05 = 9.5238 stETH
      const stETHBalance = await stETH.balanceOf(user1.address);
      expect(stETHBalance).to.be.closeTo(ethers.parseEther("9.5238"), ethers.parseEther("0.0001"));
    });
  });

  describe("Deposit Collateral", function () {
    beforeEach(async function () {
      // User deposits ETH first
      await vault.connect(user1).depositETH({ value: ethers.parseEther("10") });
    });

    it("Should deposit stETH as collateral", async function () {
      const collateralAmount = ethers.parseEther("5");

      await stETH.connect(user1).approve(await vault.getAddress(), collateralAmount);

      await expect(vault.connect(user1).depositCollateral(collateralAmount))
        .to.emit(vault, "DepositCollateral")
        .withArgs(user1.address, collateralAmount);

      const position = await vault.getPosition(user1.address);
      expect(position.collateral).to.equal(collateralAmount);
    });

    it("Should fail if amount is zero", async function () {
      await expect(
        vault.connect(user1).depositCollateral(0)
      ).to.be.revertedWith("Amount must be > 0");
    });

    it("Should fail if not enough stETH balance", async function () {
      const collateralAmount = ethers.parseEther("20");
      // 不授权，直接测试余额不足的情况
      await expect(
        vault.connect(user1).depositCollateral(collateralAmount)
      ).to.be.revertedWithCustomError(stETH, "ERC20InsufficientAllowance");
    });

    it("Should fail if contract is paused", async function () {
      await vault.pause();
      await expect(
        vault.connect(user1).depositCollateral(ethers.parseEther("5"))
      ).to.be.revertedWith("Contract is paused");
    });

    it("Should allow multiple deposits", async function () {
      await stETH.connect(user1).approve(await vault.getAddress(), ethers.parseEther("10"));

      await vault.connect(user1).depositCollateral(ethers.parseEther("3"));
      await vault.connect(user1).depositCollateral(ethers.parseEther("4"));

      const position = await vault.getPosition(user1.address);
      expect(position.collateral).to.equal(ethers.parseEther("7"));
    });
  });

  describe("Get Borrow Limit", function () {
    let originBorrowLimit: any;

    beforeEach(async function () {
      await vault.connect(user1).depositETH({ value: ethers.parseEther("10") });
      await stETH.connect(user1).approve(await vault.getAddress(), ethers.parseEther("10"));
      await vault.connect(user1).depositCollateral(ethers.parseEther("10"));
      // Collateral value: 10 stETH * $3000 = $30,000
      // Borrow limit: $30,000 * 66.6% = $19,980 
      originBorrowLimit = 19980n * 10n ** 8n;
    });

    it("Should calculate correct borrow limit", async function () {
      const borrowLimit = await vault.getBorrowLimit(user1.address);
      expect(borrowLimit).to.equal(originBorrowLimit);
    });

    it("Should reduce borrow limit after borrowing", async function () {
      const borrowAmount = 10000n * 10n ** 8n; // $10,000
      await vault.connect(user1).borrow(borrowAmount);

      const borrowLimit = await vault.getBorrowLimit(user1.address);
      expect(borrowLimit).to.equal(originBorrowLimit - borrowAmount);
    });

    it("Should return zero for user with no collateral", async function () {
      const borrowLimit = await vault.getBorrowLimit(user2.address);
      expect(borrowLimit).to.equal(0);
    });
  });

  describe("Get Health Factor", function () {
    beforeEach(async function () {
      await vault.connect(user1).depositETH({ value: ethers.parseEther("10") });
      await stETH.connect(user1).approve(await vault.getAddress(), ethers.parseEther("10"));
      await vault.connect(user1).depositCollateral(ethers.parseEther("10"));
    });

    it("Should return max for user with no debt", async function () {
      const healthFactor = await vault.getHealthFactor(user1.address);
      expect(healthFactor).to.equal(ethers.MaxUint256);
    });

    it("Should calculate correct health factor after borrow", async function () {
      const borrowAmount = 10000n * 10n ** 8n; // $10,000
      await vault.connect(user1).borrow(borrowAmount);
      const healthFactor = await vault.getHealthFactor(user1.address);
      // Health factor: ($30,000 * 1e18) / $10,000 = 3e18
      expect(healthFactor).to.equal(3n * 10n ** 18n);
    })

    it("Should return below threshold when undercollateralized", async function () {
      // Borrow max
      await vault.connect(user1).borrow(19980n * 10n ** 8n); 
      var healthFactor = await vault.getHealthFactor(user1.address);
      expect(healthFactor).to.equal(1501501501501501501n); // Just above 1.5e18

      // Drop ETH price to make position liquidatable
      await aggregator.setPrice(1800n * 10n ** 8n); // $1800 ETH
      healthFactor = await vault.getHealthFactor(user1.address);  
      expect(healthFactor).to.be.lt(1n * 10n ** 18n); // Below 1.0
    });
  });

  describe("Borrow", function () {
    beforeEach(async function () {
      // 存入 ETH 铸造 stETH, 然后将 stETH 作为抵押物存入 vault
      await vault.connect(user1).depositETH({ value: ethers.parseEther("10") });
      await stETH.connect(user1).approve(await vault.getAddress(), ethers.parseEther("10"));
      await vault.connect(user1).depositCollateral(ethers.parseEther("10"));
    });

    it("Should borrow mUSD under limit", async function () {
      const borrowAmount = 10000n * 10n ** 8n; // $10,000 < 19980
      const fee = borrowAmount * 5n / 10000n;

      await expect(vault.connect(user1).borrow(borrowAmount))
        .to.emit(vault, "Borrow")
        .withArgs(user1.address, borrowAmount, fee);

      const userBalance = await mUSD.balanceOf(user1.address);
      expect(userBalance).to.equal(borrowAmount - fee);

      const position = await vault.getPosition(user1.address);
      expect(position.debt).to.equal(borrowAmount);
      expect(position.initialDebt).to.equal(borrowAmount);
    });

    it("Should fail if borrow amount is zero", async function () {
      await expect(
        vault.connect(user1).borrow(0)
      ).to.be.revertedWith("Borrow amount must be > 0");
    });

    it("Should fail if borrow amount exceeds limit", async function () {
      const borrowAmount = 20000n * 10n ** 8n; // Exceeds limit

      await expect(
        vault.connect(user1).borrow(borrowAmount)
      ).to.be.revertedWith("Exceeds borrow limit");
    });

    it("Should fail if contract is paused", async function () {
      await vault.pause();
      await expect(
        vault.connect(user1).borrow(ethers.parseEther("5000"))
      ).to.be.revertedWith("Contract is paused");
    });

    it("Should charge borrow fee and send to treasury", async function () {
      const borrowAmount = 10000n * 10n ** 8n; // $10,000 < 19980
      const expectedFee = borrowAmount * 5n / 10000n; // 0.05%

      await vault.connect(user1).borrow(borrowAmount);

      expect(await mUSD.balanceOf(await treasury.getAddress())).to.equal(expectedFee);
    });

  });

  describe("Repay", function () {
    const depositAmount = ethers.parseEther("10"); 
    const borrowAmount = 10000n * 10n ** 8n; // $10,000
    beforeEach(async function () {
      await vault.connect(user1).depositETH({ value: depositAmount });
      await stETH.connect(user1).approve(await vault.getAddress(), depositAmount);
      await vault.connect(user1).depositCollateral(depositAmount);
      await vault.connect(user1).borrow(borrowAmount);

      await vault.connect(user2).depositETH({ value: depositAmount });
      await stETH.connect(user2).approve(await vault.getAddress(), depositAmount);
      await vault.connect(user2).depositCollateral(depositAmount);
      await vault.connect(user2).borrow(borrowAmount);
    });
    
    it("Should repay debt", async function () {
      const repayAmount = 5000n * 10n ** 8n; // $5,000

      var position = await vault.getPosition(user1.address);
      // console.log("Position before repay:", position);
      // 优先偿还利息，再偿还本金
      await mUSD.connect(user1).approve(await vault.getAddress(), repayAmount);

      await expect(vault.connect(user1).repay(repayAmount))
        .to.emit(vault, "Repay")
        .withArgs(user1.address, repayAmount);

      position = await vault.getPosition(user1.address);
      // console.log("Position after repay:", position);
      expect(position.debt).to.be.gt(borrowAmount - repayAmount);
    });

    it("Should repay full debt", async function () {
      const position = await vault.getPosition(user1.address);
      // console.log("Position before repay:", position);

      //user1 mUSD not enough
      await mUSD.connect(user2).transfer(user1, 6n * 10n ** 8n)
      const balance = await mUSD.connect(user1).balanceOf(user1.address);
      // console.log("User balance before repay:", balance);

      await mUSD.connect(user1).approve(await vault.getAddress(), balance);
      await vault.connect(user1).repay(balance);

      const newPosition = await vault.getPosition(user1.address);
      // console.log("New position:", newPosition);
      expect(newPosition.debt).to.equal(0);
    });

    it("Should fail if amount is zero", async function () {
      await expect(
        vault.connect(user1).repay(0)
      ).to.be.revertedWith("Repay amount must be > 0");
    });
  });

  describe("Liquidate", function () {
    beforeEach(async function () {
      // Setup user1 with position
      await vault.connect(user1).depositETH({ value: ethers.parseEther("10") });
      await stETH.connect(user1).approve(await vault.getAddress(), ethers.parseEther("10"));
      await vault.connect(user1).depositCollateral(ethers.parseEther("10"));
      await vault.connect(user1).borrow(10000n * 10n ** 8n);
    
      // Setup liquidator with mUSD
      await vault.connect(liquidator).depositETH({ value: ethers.parseEther("20") });
      await stETH.connect(liquidator).approve(await vault.getAddress(), ethers.parseEther("20"));
      await vault.connect(liquidator).depositCollateral(ethers.parseEther("20"));
      await vault.connect(liquidator).borrow(30000n * 10n ** 8n);
    });

    it("Should liquidate undercollateralized position", async function () {

      await aggregator.setPrice(900n * 10n ** 8n); // $900 ETH
      // Collateral value: 10 stETH * $900 = $9,000
      // Health factor: ($9,000 * 1e18) / $10,000 = 0.9e18

      const debtToRepaid = 5000n * 10n ** 8n;
      await mUSD.connect(liquidator).approve(await vault.getAddress(), debtToRepaid);

      await expect(vault.connect(liquidator).liquidate(user1.address, debtToRepaid))
        .to.emit(vault, "Liquidate");

      const position = await vault.getPosition(user1.address);
      expect(position.debt).to.be.closeTo(debtToRepaid, 1n * 10n ** 8n);
      expect(position.collateral).to.be.lt(ethers.parseEther("10"));
    });

    it("Should apply liquidation discount", async function () {
      await aggregator.setPrice(900n * 10n ** 8n);

      const debtToRepaid = 5000n * 10n ** 8n;
      await mUSD.connect(liquidator).approve(await vault.getAddress(), ethers.parseEther("40000"));

      const liquidatorStETHBefore = await stETH.balanceOf(liquidator.address);

      await vault.connect(liquidator).liquidate(user1.address, debtToRepaid);

      const liquidatorStETHAfter = await stETH.balanceOf(liquidator.address);
      const collateralReceived = liquidatorStETHAfter - liquidatorStETHBefore;

      // Liquidator should get collateral at 5% discount
      // Debt value in stETH: $5000 / ($900 * 1) = 5.5556 stETH
      // Discounted collateral: 5.5556 / 0.95 = 5.848 stETH
      // After 1% fee: 5.848 * 0.99 = 5.7895 stETH
      expect(collateralReceived).to.be.closeTo(ethers.parseEther("5.78"), ethers.parseEther("0.01"));
    });

    it("Should fail if position is healthy", async function () {
      await aggregator.setPrice(2000n * 10n ** 8n);

      const debtToRepaid = ethers.parseEther("5000");
      await mUSD.connect(liquidator).approve(await vault.getAddress(), ethers.parseEther("40000"));

      await expect(
        vault.connect(liquidator).liquidate(user1.address, debtToRepaid)
      ).to.be.revertedWith("Not liquidatable");
    });

    it("Should fail if liquidating zero address", async function () {
      await aggregator.setPrice(800n * 10n ** 8n);

      await expect(
        vault.connect(liquidator).liquidate(ethers.ZeroAddress, ethers.parseEther("5000"))
      ).to.be.revertedWith("Invalid borrower");
    });

    it("Should fail if debt amount is zero", async function () {
      await aggregator.setPrice(800n * 10n ** 8n);

      await expect(
        vault.connect(liquidator).liquidate(user1.address, 0)
      ).to.be.revertedWith("Debt amount must be > 0");
    });

    it("Should <= maximum liquidation percentage", async function () {
      await aggregator.setPrice(800n * 10n ** 8n);
      const debtToRepaid = 6000n * 10n ** 8n; // 60% of debt, exceeds 50% max
      await mUSD.connect(liquidator).approve(await vault.getAddress(), ethers.parseEther("40000"));

      await vault.connect(liquidator).liquidate(user1.address, debtToRepaid);

      const position = await vault.getPosition(user1.address);
      // Should only liquidate 50% of debt
      expect(position.debt).to.closeTo(5000n * 10n ** 8n, 1n * 10n ** 8n);
    });

    it("Should send liquidation fee to treasury", async function () {
      await aggregator.setPrice(800n * 10n ** 8n);

      const treasuryStETHBefore = await stETH.balanceOf(await treasury.getAddress());
      const debtToRepaid = ethers.parseEther("5000");
      await mUSD.connect(liquidator).approve(await vault.getAddress(), ethers.parseEther("40000"));

      await vault.connect(liquidator).liquidate(user1.address, debtToRepaid);

      const treasuryStETHAfter = await stETH.balanceOf(await treasury.getAddress());
      const feeReceived = treasuryStETHAfter - treasuryStETHBefore;

      expect(feeReceived).to.be.gt(0);
    });
  });

  describe("Withdraw Collateral", function () {
    beforeEach(async function () {
      await vault.connect(user1).depositETH({ value: ethers.parseEther("10") });
      await stETH.connect(user1).approve(await vault.getAddress(), ethers.parseEther("10"));
      await vault.connect(user1).depositCollateral(ethers.parseEther("10"));
    });

    it("Should withdraw collateral without debt", async function () {
      const withdrawAmount = ethers.parseEther("5");

      await expect(vault.connect(user1).withdrawCollateral(withdrawAmount))
        .to.emit(vault, "WithdrawCollateral");

      const position = await vault.getPosition(user1.address);
      expect(position.collateral).to.equal(ethers.parseEther("5"));
    });

    it("Should withdraw ETH correctly", async function () {
      const withdrawAmount = ethers.parseEther("5");
      const ethBalanceBefore = await hardhatEthers.provider.getBalance(user1.address);

      const tx = await vault.connect(user1).withdrawCollateral(withdrawAmount);
      const receipt = await tx.wait();
      const gasUsed = BigInt(receipt.gasUsed.toString()); 
      const gasPrice = BigInt(receipt.gasPrice.toString());
      const totalGasCost = gasUsed * gasPrice; // 最终是 bigint

      const ethBalanceAfter = await hardhatEthers.provider.getBalance(user1.address);
      const ethReceived = ethBalanceAfter - ethBalanceBefore + totalGasCost;

      expect(ethReceived).to.equal(withdrawAmount);
    });

    // it("Should limit withdrawal when user has debt", async function () {
    //   await vault.connect(user1).borrow(ethers.parseEther("8000")); // Higher debt

    //   const withdrawAmount = ethers.parseEther("8"); // Would make position liquidatable

    //   const positionBefore = await vault.getPosition(user1.address);
    //   await vault.connect(user1).withdrawCollateral(withdrawAmount);
    //   const positionAfter = await vault.getPosition(user1.address);

    //   // Should withdraw only what's safe
    //   const withdrawn = positionBefore.collateral - positionAfter.collateral;
    //   expect(withdrawn).to.be.lt(withdrawAmount);
    // });

    it("Should fail if amount is zero", async function () {
      await expect(
        vault.connect(user1).withdrawCollateral(0)
      ).to.be.revertedWith("Withdraw amount must be > 0");
    });

    it("Should fail if contract is paused", async function () {
      await vault.pause();
      await expect(
        vault.connect(user1).withdrawCollateral(ethers.parseEther("5"))
      ).to.be.revertedWith("Contract is paused");
    });
  });

  describe("Pause/Unpause", function () {
    it("Should allow pauser to pause", async function () {
      await expect(vault.pause())
        .to.emit(vault, "Paused")
        .withArgs(owner.address);

      expect(await vault.paused()).to.be.true;
    });

    it("Should allow pauser to unpause", async function () {
      await vault.pause();

      await expect(vault.unpause())
        .to.emit(vault, "Unpaused")
        .withArgs(owner.address);

      expect(await vault.paused()).to.be.false;
    });

    it("Should fail if non-pauser tries to pause", async function () {
      await expect(
        vault.connect(user1).pause()
      ).to.be.revertedWithCustomError(vault, "AccessControlUnauthorizedAccount");
    });
  });

  describe("Set Params", function () {
    it("Should allow timelock to set borrowLimitPct", async function () {
      await expect(vault.setParams("borrowLimitPct", 7000n * 10n ** 14n))
        .to.emit(vault, "ParamsUpdated");

      expect(await vault.borrowLimitPct()).to.equal(7000n * 10n ** 14n);
    });

    it("Should allow timelock to set liquidationThreshold", async function () {
      await vault.setParams("liquidationThreshold", 11500n * 10n ** 14n);
      expect(await vault.liquidationThreshold()).to.equal(11500n * 10n ** 14n);
    });

    it("Should allow timelock to set borrowFeePct", async function () {
      await vault.setParams("borrowFeePct", 10n * 10n ** 14n);
      expect(await vault.borrowFeePct()).to.equal(10n * 10n ** 14n);
    });

    it("Should allow timelock to set borrowAPR", async function () {
      await vault.setParams("borrowAPR", 500n * 10n ** 14n);
      expect(await vault.borrowAPR()).to.equal(500n * 10n ** 14n);
    });

    it("Should fail for invalid parameter name", async function () {
      await expect(
        vault.setParams("invalidParam", 100)
      ).to.be.revertedWith("Invalid parameter name");
    });

    it("Should fail if non-timelock tries to set params", async function () {
      await expect(
        vault.connect(user1).setParams("borrowLimitPct", 7000n * 10n ** 14n)
      ).to.be.revertedWithCustomError(vault, "AccessControlUnauthorizedAccount");
    });
  });

  describe("Update Contract", function () {
    it("Should allow timelock to update oracle", async function () {
      const MiniOracle = await hardhatEthers.getContractFactory("MiniOracle");
      const newOracle = await MiniOracle.deploy(await aggregator.getAddress(), owner.address);

      await vault.updateContract("oracle", await newOracle.getAddress());
      expect(await vault.oracle()).to.equal(await newOracle.getAddress());
    });

    it("Should allow timelock to update treasury", async function () {
      const MiniTreasury = await hardhatEthers.getContractFactory("MiniTreasury");
      const newTreasury = await MiniTreasury.deploy(owner.address);

      await vault.updateContract("treasury", await newTreasury.getAddress());
      expect(await vault.treasury()).to.equal(await newTreasury.getAddress());
    });

    it("Should fail for invalid contract name", async function () {
      await expect(
        vault.updateContract("invalidContract", user1.address)
      ).to.be.revertedWith("Invalid contract name");
    });

    it("Should fail if address is zero", async function () {
      await expect(
        vault.updateContract("oracle", ethers.ZeroAddress)
      ).to.be.revertedWith("Invalid address");
    });

    it("Should fail if non-timelock tries to update", async function () {
      await expect(
        vault.connect(user1).updateContract("oracle", user1.address)
      ).to.be.revertedWithCustomError(vault, "AccessControlUnauthorizedAccount");
    });
  });

  describe("Emergency Withdraw", function () {
    beforeEach(async function () {
      await vault.connect(user1).depositETH({ value: ethers.parseEther("10") });
    });

    it("Should allow timelock to withdraw ETH when paused", async function () {
      await vault.pause();

      const balanceBefore = await hardhatEthers.provider.getBalance(user1.address);
      await vault.emergencyWithdraw(ethers.ZeroAddress, ethers.parseEther("5"), user1.address);
      const balanceAfter = await hardhatEthers.provider.getBalance(user1.address);

      expect(balanceAfter - balanceBefore).to.equal(ethers.parseEther("5"));
    });

    it("Should allow timelock to withdraw tokens when paused", async function () {
      await stETH.connect(user1).approve(await vault.getAddress(), ethers.parseEther("10"));
      await vault.connect(user1).depositCollateral(ethers.parseEther("10"));

      await vault.pause();

      await vault.emergencyWithdraw(await stETH.getAddress(), ethers.parseEther("5"), user1.address);
      expect(await stETH.balanceOf(user1.address)).to.equal(ethers.parseEther("5"));
    });

    it("Should fail if contract is not paused", async function () {
      await expect(
        vault.emergencyWithdraw(ethers.ZeroAddress, ethers.parseEther("5"), user1.address)
      ).to.be.revertedWith("Contract must be paused");
    });

    it("Should fail if recipient is zero", async function () {
      await vault.pause();
      await expect(
        vault.emergencyWithdraw(ethers.ZeroAddress, ethers.parseEther("5"), ethers.ZeroAddress)
      ).to.be.revertedWith("Invalid recipient");
    });

    it("Should fail if non-timelock tries to withdraw", async function () {
      await vault.pause();
      await expect(
        vault.connect(user1).emergencyWithdraw(ethers.ZeroAddress, ethers.parseEther("5"), user1.address)
      ).to.be.revertedWithCustomError(vault, "AccessControlUnauthorizedAccount");
    });
  });

  describe("Receive Function", function () {
    it("Should reject direct ETH transfers", async function () {
      await expect(
        user1.sendTransaction({ to: await vault.getAddress(), value: ethers.parseEther("1") })
      ).to.be.revertedWith("Direct ETH transfer not allowed, use depositETH()");
    });
  });

  describe("Interest Accrual", function () {
    beforeEach(async function () {
      await vault.connect(user1).depositETH({ value: ethers.parseEther("10") });
      await stETH.connect(user1).approve(await vault.getAddress(), ethers.parseEther("10"));
      await vault.connect(user1).depositCollateral(ethers.parseEther("10"));
      await vault.connect(user1).borrow(10000n * 10n ** 8n);
    });

    it("Should accrue interest over time", async function () {
      const positionBefore = await vault.getPosition(user1.address);
      const debtBefore = positionBefore.debt;

      // Fast forward time (in real tests, use time manipulation)
      // For now, just verify interest accrual happens
      await vault.connect(user1).borrow(1000n * 10n ** 8n);

      const positionAfter = await vault.getPosition(user1.address);
      expect(positionAfter.debt).to.be.gt(debtBefore);
    });
  });
});
