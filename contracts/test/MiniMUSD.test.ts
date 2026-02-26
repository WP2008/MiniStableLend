import { expect } from "chai";
import { network } from "hardhat";

// 连接网络并获取 ethers 对象
const provider = await network.connect();
const { ethers } = provider;

describe("MiniMUSD", function () {
  let mUSD: any;
  let owner: any;
  let user: any;
  let vault: any;

  beforeEach(async function () {
    // 获取测试账户
    [owner, user, vault] = await ethers.getSigners();
    // 部署 MiniMUSD 合约
    const MiniMUSD = await ethers.getContractFactory("MiniMUSD");
    mUSD = await MiniMUSD.deploy();
    // 设置 vault 地址
    await mUSD.setVault(vault.address);
  });

  describe("Deployment", function () {
    it("Should set the correct name and symbol", async function () {
      expect(await mUSD.name()).to.equal("Mini MUSD");
      expect(await mUSD.symbol()).to.equal("mUSD");
    });

    it("Should set the correct vault address", async function () {
      expect(await mUSD.vault()).to.equal(vault.address);
    });

    it("Should revert when vault address is zero", async function () {
      const MiniMUSD = await ethers.getContractFactory("MiniMUSD");
      const mUSD = await MiniMUSD.deploy();
      await expect(
        mUSD.setVault(ethers.ZeroAddress)
      ).to.be.revertedWith("Vault address cannot be zero");
    });

    it("Should start with zero supply", async function () {
      expect(await mUSD.totalSupply()).to.equal(0n);
    });
  });

  describe("Minting", function () {
    it("Should allow vault to mint tokens", async function () {
      const amount = ethers.parseEther("1000");
      await mUSD.connect(vault).mint(user.address, amount);

      expect(await mUSD.balanceOf(user.address)).to.equal(amount);
      expect(await mUSD.totalSupply()).to.equal(amount);
    });

    it("Should revert when non-vault tries to mint", async function () {
      const amount = ethers.parseEther("1000");
      await expect(
        mUSD.connect(user).mint(user.address, amount)
      ).to.be.revertedWith("Only Vault");
    });

    it("Should revert when minting to zero address", async function () {
      const amount = ethers.parseEther("1000");
      await expect(
        mUSD.connect(vault).mint(ethers.ZeroAddress, amount)
      ).to.be.revertedWith("Recipient cannot be zero");
    });

    it("Should emit Transfer event on mint", async function () {
      const amount = ethers.parseEther("1000");
      await expect(mUSD.connect(vault).mint(user.address, amount))
        .to.emit(mUSD, "Transfer")
        .withArgs(ethers.ZeroAddress, user.address, amount);
    });
  });

  describe("Burning", function () {
    beforeEach(async function () {
      // mint 1000 token 给 user
      const amount = ethers.parseEther("1000");
      await mUSD.connect(vault).mint(user.address, amount);
    });

    it("Should allow vault to burn tokens", async function () {
      const amount = ethers.parseEther("500");
      await mUSD.connect(vault).burn(user.address, amount);

      expect(await mUSD.balanceOf(user.address)).to.equal(ethers.parseEther("500"));
      expect(await mUSD.totalSupply()).to.equal(ethers.parseEther("500"));
    });

    it("Should revert when non-vault tries to burn", async function () {
      const amount = ethers.parseEther("500");
      await expect(
        mUSD.connect(user).burn(user.address, amount)
      ).to.be.revertedWith("Only Vault");
    });

    it("Should revert when burning from zero address", async function () {
      const amount = ethers.parseEther("500");
      await expect(
        mUSD.connect(vault).burn(ethers.ZeroAddress, amount)
      ).to.be.revertedWith("From cannot be zero");
    });

    it("Should revert when burning more than balance", async function () {
      const amount = ethers.parseEther("1500");
      await expect(
        mUSD.connect(vault).burn(user.address, amount)
      ).to.be.revertedWithCustomError(mUSD, "ERC20InsufficientBalance");
    });

    it("Should emit Transfer event on burn", async function () {
      const amount = ethers.parseEther("500");
      await expect(mUSD.connect(vault).burn(user.address, amount))
        .to.emit(mUSD, "Transfer")
        .withArgs(user.address, ethers.ZeroAddress, amount);
    });
  });

  describe("Transfers", function () {
    beforeEach(async function () {
      // mint 1000 token 给 user
      const amount = ethers.parseEther("1000");
      await mUSD.connect(vault).mint(user.address, amount);
    });

    it("Should allow ERC20 transfers", async function () {
      const signers = await ethers.getSigners();
      const recipient = signers[3];
      const amount = ethers.parseEther("100");

      await mUSD.connect(user).transfer(recipient.address, amount);

      expect(await mUSD.balanceOf(user.address)).to.equal(ethers.parseEther("900"));
      expect(await mUSD.balanceOf(recipient.address)).to.equal(amount);
    });

    it("Should allow transferFrom with approval", async function () {
      const signers = await ethers.getSigners();
      const spender = signers[3];
      const amount = ethers.parseEther("100");

      await mUSD.connect(user).approve(spender.address, amount);
      await mUSD.connect(spender).transferFrom(user.address, spender.address, amount);

      expect(await mUSD.balanceOf(user.address)).to.equal(ethers.parseEther("900"));
      expect(await mUSD.balanceOf(spender.address)).to.equal(amount);
      expect(await mUSD.allowance(user.address, spender.address)).to.equal(0);
    });

    it("Should revert when insufficient allowance", async function () {
      const signers = await ethers.getSigners();
      const spender = signers[3];
      const amount = ethers.parseEther("100");

      await expect(
        mUSD.connect(spender).transferFrom(user.address, spender.address, amount)
      ).to.be.revertedWithCustomError(mUSD, "ERC20InsufficientAllowance");
    });
  });
});