import { expect } from "chai";
import { network } from "hardhat";


const { ethers } = await network.connect();

describe("MiniOracle", function () {
  let oracle: any;
  let aggregator: any;
  let owner: any;
  let user: any;
  let mockPrice: any;

  beforeEach(async function () {
    [owner, user] = await ethers.getSigners();
    mockPrice = 3000e8;

    // Deploy Mock Aggregator
    const MockAggregator = await ethers.getContractFactory("MockAggregatorV3");
    aggregator = await MockAggregator.deploy(mockPrice, 8); // mock price 3000 usd/eth, 8 decimals

    // Deploy MiniOracle
    const MiniOracle = await ethers.getContractFactory("MiniOracle");
    oracle = await MiniOracle.deploy(await aggregator.getAddress(), owner.address);
  });

  describe("Deployment", function () {
    it("Should init with correct params", async function () {
      expect(await oracle.ethUsdFeed()).to.equal(await aggregator.getAddress());
      expect(await oracle.hasRole(await oracle.TIMELOCK_ROLE(), owner.address)).to.be.true;
      expect(await oracle.DECIMALS()).to.equal(8);
      expect(await oracle.PRICE_VALID_WINDOW()).to.equal(3600);
    });

    it("Should revert when feed address is zero", async function () {
      const MiniOracle = await ethers.getContractFactory("MiniOracle");
      await expect(
        MiniOracle.deploy(ethers.ZeroAddress, owner.address)
      ).to.be.revertedWith("Feed address cannot be zero");
    });

    it("Should revert when timelock address is zero", async function () {
      const MiniOracle = await ethers.getContractFactory("MiniOracle");
      await expect(
        MiniOracle.deploy(await aggregator.getAddress(), ethers.ZeroAddress)
      ).to.be.revertedWith("Timelock address cannot be zero");
    });
  });

  describe("getETHPrice", function () {
    it("Should return the current ETH price", async function () {
      const price = await oracle.getETHPrice();
      expect(price).to.equal(mockPrice);
    });

    it("Should fail if price invalid", async function () {
      await aggregator.setPrice(0);
      await expect(oracle.getETHPrice()).to.be.revertedWith("Invalid price");

      await aggregator.setPrice(-100);
      await expect(oracle.getETHPrice()).to.be.revertedWith("Invalid price");
    });

    it("Should fail if price is expired", async function () {
      // Set old timestamp
      await aggregator.setPrice(3000e8);

      // Increase time by more than valid window
      await ethers.provider.send("evm_increaseTime", [3600 + 1]);
      await ethers.provider.send("evm_mine");
      
      await expect(oracle.getETHPrice()).to.be.revertedWith("Price expired");
    });
  });

  describe("ethToUsd", function () {
    it("Should convert ETH to USD correctly", async function () {
      const ethAmount = ethers.parseEther("10");
      const usdAmount = await oracle.ethToUsd(ethAmount);
      expect(usdAmount).to.equal(ethers.parseEther("30000"));
    });


    it("Should handle different ETH prices", async function () {
      await aggregator.setPrice(4000e8); // $4000

      const ethAmount = ethers.parseEther("1");
      const usdAmount = await oracle.ethToUsd(ethAmount);
      expect(usdAmount).to.equal(ethers.parseEther("4000"));
    });
  });

  describe("setFeed", function () {
    it("Should allow timelock to set new feed", async function () {
      const newAggregator = await (
        await ethers.getContractFactory("MockAggregatorV3")
      ).deploy(3500e8, 8);

      await oracle.setFeed(await newAggregator.getAddress());
      expect(await oracle.ethUsdFeed()).to.equal(await newAggregator.getAddress());

      const price = await oracle.getETHPrice();
      expect(price).to.equal(3500e8);
    });

    it("Should fail if non-timelock tries to set feed", async function () {
      const newAggregator = await (
        await ethers.getContractFactory("MockAggregatorV3")
      ).deploy(3500e8, 8);

      await expect(
        oracle.connect(user).setFeed(await newAggregator.getAddress())
      ).to.be.revertedWithCustomError(oracle, "AccessControlUnauthorizedAccount");
    });

    it("Should fail if new feed address is zero", async function () {
      await expect(oracle.setFeed(ethers.ZeroAddress)).to.be.revertedWith(
        "New feed address cannot be zero"
      );
    });
  });

  describe("Access Control", function () {
    it("Should have TIMELOCK_ROLE correctly set", async function () {
      const TIMELOCK_ROLE = await oracle.TIMELOCK_ROLE();
      expect(await oracle.hasRole(TIMELOCK_ROLE, owner.address)).to.be.true;
    });

    it("Should not grant TIMELOCK_ROLE to user", async function () {
      const TIMELOCK_ROLE = await oracle.TIMELOCK_ROLE();
      expect(await oracle.hasRole(TIMELOCK_ROLE, user.address)).to.be.false;
    });
  });
});
