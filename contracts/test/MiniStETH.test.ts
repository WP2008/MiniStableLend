import { expect } from "chai";
import { network } from "hardhat";

const { ethers } = await network.connect();

describe("MiniStETH", function () {
  let stETH: any;
  let vault: any;
  let owner: any;
  let user: any;

  beforeEach(async function () {
    // 获取测试账户  假设 vault 是一个合约
    [owner, user, vault] = await ethers.getSigners();

    const MiniStETH = await ethers.getContractFactory("MiniStETH");
    stETH = await MiniStETH.deploy(owner.address);

    await stETH.setVault(vault.address);
  });

  describe("Deployment", function () {
    it("Should init with correct params", async function () {
      expect(await stETH.name()).to.equal("Mini Staked ETH");
      expect(await stETH.symbol()).to.equal("stETH");
      expect(await stETH.vault()).to.equal(vault.address);
      expect(await stETH.exchangeRate()).to.equal(1n * 10n ** 18n);
      expect(await stETH.totalSupply()).to.equal(0);
      expect(await stETH.hasRole(await stETH.TIMELOCK_ROLE(), owner.address)).to.be.true;
    });

    it("Should revert when timelock address is zero", async function () {
      const MiniStETH = await ethers.getContractFactory("MiniStETH");
      await expect(
        MiniStETH.deploy(ethers.ZeroAddress)
      ).to.be.revertedWith("Timelock address cannot be zero");
    });
  });

  describe("Mint", function () {
    it("Should allow vault to mint tokens", async function () {
      const amount = ethers.parseEther("100");
      await stETH.connect(vault).mint(user.address, amount);

      expect(await stETH.balanceOf(user.address)).to.equal(amount);
      expect(await stETH.totalSupply()).to.equal(amount);
    });

    it("Should revert if not vault to mint", async function () {
      const amount = ethers.parseEther("100");
      await expect(
        stETH.connect(user).mint(user.address, amount)
      ).to.be.revertedWith("Only Vault");
    });

    it("Should revert when mint to zero address", async function () {
      const amount = ethers.parseEther("100");
      await expect(
        stETH.connect(vault).mint(ethers.ZeroAddress, amount)
      ).to.be.revertedWith("Recipient cannot be zero");
    });

    it("Should emit Transfer event on mint", async function () {
      const amount = ethers.parseEther("100");
      await expect(stETH.connect(vault).mint(user.address, amount))
        .to.emit(stETH, "Transfer")
        .withArgs(ethers.ZeroAddress, user.address, amount);
    });
  });

  describe("Burn", function () {
    beforeEach(async function () {
      const amount = ethers.parseEther("100");
      await stETH.connect(vault).mint(user.address, amount);
    });

    it("Should allow vault to burn tokens", async function () {
      const amount = ethers.parseEther("50");
      await stETH.connect(vault).burn(user.address, amount);

      expect(await stETH.balanceOf(user.address)).to.equal(ethers.parseEther("50"));
      expect(await stETH.totalSupply()).to.equal(ethers.parseEther("50"));
    });

    it("Should revert when not vault to burn", async function () {
      const amount = ethers.parseEther("50");
      await expect(
        stETH.connect(user).burn(user.address, amount)
      ).to.be.revertedWith("Only Vault");
    });

    it("Should revert if burn from zero address", async function () {
      const amount = ethers.parseEther("50");
      await expect(
        stETH.connect(vault).burn(ethers.ZeroAddress, amount)
      ).to.be.revertedWith("From cannot be zero");
    });

    it("Should emit Transfer event on burn", async function () {
      const amount = ethers.parseEther("50");
      await expect(stETH.connect(vault).burn(user.address, amount))
        .to.emit(stETH, "Transfer")
        .withArgs(user.address, ethers.ZeroAddress, amount);
    });
  });

  describe("Exchange Rate", function () {
    it("Should allow vault to update exchange rate with ExchangeRateUpdated event", async function () {
      const oldRate = await stETH.exchangeRate();
      const newRate = 105n * 10n ** 16n;

      await expect(stETH.connect(vault).updateExchangeRate(newRate))
        .to.emit(stETH, "ExchangeRateUpdated")
        .withArgs(oldRate, newRate);

      expect(await stETH.exchangeRate()).to.equal(newRate);
    });

    it("Should revert when not vault to update exchange rate", async function () {
      const newRate = 105n * 10n ** 16n;
      await expect(
        stETH.connect(user).updateExchangeRate(newRate)
      ).to.be.revertedWith("Only Vault");
    });

    it("Should revert when exchange rate decreases", async function () {
      const newRate = 95n * 10n ** 16n; // newRate < 1
      await expect(
        stETH.connect(vault).updateExchangeRate(newRate)
      ).to.be.revertedWith("Rate cannot decrease");
    });
  });

  describe("Add StETH Interest", function () {
    it("Should allow vault to add stETH interest", async function () {
      const interest = ethers.parseEther("10");

      await stETH.connect(vault).addStETHInterest(interest);

      expect(await stETH.totalStETHInterest()).to.equal(interest);
    });

    it("Should revert when not vault to add interest", async function () {
      const interest = ethers.parseEther("10");
      await expect(
        stETH.connect(user).addStETHInterest(interest)
      ).to.be.revertedWith("Only Vault");
    });
  });

  describe("Set Vault", function () {
    it("Should allow timelock to set new vault", async function () {
      const newVault = user.address;
      await stETH.setVault(newVault);

      expect(await stETH.vault()).to.equal(newVault);
    });

    it("Should revert when not timelock to set vault", async function () {
      await expect(
        stETH.connect(user).setVault(user.address)
      ).to.be.revertedWithCustomError(stETH, "AccessControlUnauthorizedAccount");
    });

    it("Should revert if vault address is zero", async function () {
      await expect(
        stETH.setVault(ethers.ZeroAddress)
      ).to.be.revertedWith("Vault address cannot be zero");
    });
  });
});
