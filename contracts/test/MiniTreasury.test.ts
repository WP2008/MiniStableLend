import { expect } from "chai";
import { network } from "hardhat";

const { ethers } = await network.connect();

describe("MiniTreasury", function () {
  let treasury: any;
  let mockToken: any;
  let owner: any;
  let user: any;
  let recipient: any;

  beforeEach(async function () {
    [owner, user, recipient] = await ethers.getSigners();

    const MiniTreasury = await ethers.getContractFactory("MiniTreasury");
    treasury = await MiniTreasury.deploy(owner.address)

    const MockToken = await ethers.getContractFactory("MockToken");
    mockToken = await MockToken.deploy("Test Token", "TEST");
  });

  describe("Deployment", function () {
    it("Should set TIMELOCK_ROLE to owner", async function () {
      const TIMELOCK_ROLE = await treasury.TIMELOCK_ROLE();
      expect(await treasury.hasRole(TIMELOCK_ROLE, owner.address)).to.be.true;
    });

    it("Should revert if timelock address is zero", async function () {
      const MiniTreasury = await ethers.getContractFactory("MiniTreasury");
      await expect(
        MiniTreasury.deploy(ethers.ZeroAddress)
      ).to.be.revertedWith("Timelock address cannot be zero");
    });

    it("Should start with zero balance", async function () {
      expect(await treasury.getBalance(ethers.ZeroAddress)).to.equal(0);
    });
  });

  describe("Receiving ETH", function () {
    it("Should receive ETH via fallback", async function () {
      const amount = ethers.parseEther("10");

      await expect(
        owner.sendTransaction({ to: await treasury.getAddress(), value: amount })
      ).to.emit(treasury, "FeeReceived")
        .withArgs(ethers.ZeroAddress, amount);

      expect(await treasury.getBalance(ethers.ZeroAddress)).to.equal(amount);
    });

    it("Should track total ETH received", async function () {
      await owner.sendTransaction({ to: await treasury.getAddress(), value: ethers.parseEther("5") });
      await owner.sendTransaction({ to: await treasury.getAddress(), value: ethers.parseEther("10") });

      expect(await treasury.getBalance(ethers.ZeroAddress)).to.equal(ethers.parseEther("15"));
    });
  });

  describe("Receiving ERC20", function () {
    beforeEach(async function () {
      // Mint tokens to user
      await mockToken.mint(user.address, ethers.parseEther("1000"));
    });

    it("Should receive ERC20 tokens", async function () {
      const amount = ethers.parseEther("100");
      const tokenAddress = await mockToken.getAddress();
      const treasuryAddress = await treasury.getAddress();
      await mockToken.connect(user).approve(treasuryAddress, amount);

      await expect(treasury.connect(user).receiveFee(tokenAddress, amount))
        .to.emit(treasury, "FeeReceived")
        .withArgs(tokenAddress, amount);

      expect(await mockToken.balanceOf(treasuryAddress)).to.equal(amount);
    });

    it("Should fail if amount is zero", async function () {
      await expect(
        treasury.connect(user).receiveFee(await mockToken.getAddress(), 0)
      ).to.be.revertedWith("Amount must be > 0");
    });

    it("Should fail if token address is zero", async function () {
      const amount = ethers.parseEther("100");
      await expect(
        treasury.connect(user).receiveFee(ethers.ZeroAddress, amount)
      ).to.be.revertedWith("Token address cannot be zero");
    });

    it("Should fail if don't approve", async function () {
      const amount = ethers.parseEther("100");
      await expect(
        treasury.connect(user).receiveFee(await mockToken.getAddress(), amount)
      ).to.be.revert(ethers);
    });

  });

  describe("Withdrawing ETH", function () {
    beforeEach(async function () {
      // transfer ETH to treasury
      await owner.sendTransaction({ to: await treasury.getAddress(), value: ethers.parseEther("10") });
    });

    it("Should allow timelock to withdraw ETH", async function () {
      const amount = ethers.parseEther("5");
      const balanceBefore = await ethers.provider.getBalance(recipient.address);

      await expect(treasury.withdraw(recipient.address, ethers.ZeroAddress, amount))
        .to.emit(treasury, "FeeWithdrawn")
        .withArgs(recipient.address, ethers.ZeroAddress, amount);

      const balanceAfter = await ethers.provider.getBalance(recipient.address);
      expect(balanceAfter - balanceBefore).to.equal(amount);
      expect(await treasury.getBalance(ethers.ZeroAddress)).to.equal(ethers.parseEther("5"));
    });

    it("Should fail if non-timelock tries to withdraw", async function () {
      const amount = ethers.parseEther("5");
      await expect(
        treasury.connect(user).withdraw(recipient.address, ethers.ZeroAddress, amount)
      ).to.be.revertedWithCustomError(treasury, "AccessControlUnauthorizedAccount");
    });

    it("Should fail if recipient is zero address", async function () {
      const amount = ethers.parseEther("5");
      await expect(
        treasury.withdraw(ethers.ZeroAddress, ethers.ZeroAddress, amount)
      ).to.be.revertedWith("Recipient cannot be zero");
    });

    it("Should fail if amount is zero", async function () {
      await expect(
        treasury.withdraw(recipient.address, ethers.ZeroAddress, 0)
      ).to.be.revertedWith("Amount must be > 0");
    });

    it("Should fail if insufficient ETH balance", async function () {
      const amount = ethers.parseEther("20");
      await expect(
        treasury.withdraw(recipient.address, ethers.ZeroAddress, amount)
      ).to.be.revertedWith("Insufficient ETH balance");
    });
  });

  describe("Withdrawing ERC20", function () {
    beforeEach(async function () {
      // Mint and deposit tokens to treasury
      await mockToken.mint(user.address, ethers.parseEther("1000"));
      await mockToken.connect(user).approve(await treasury.getAddress(), ethers.parseEther("100"));
      await treasury.connect(user).receiveFee(await mockToken.getAddress(), ethers.parseEther("100"));
    });

    it("Should allow timelock to withdraw ERC20", async function () {
      const amount = ethers.parseEther("50");

      const mockTokenAddress = await mockToken.getAddress();

      await expect(treasury.withdraw(recipient.address, mockTokenAddress, amount))
        .to.emit(treasury, "FeeWithdrawn")
        .withArgs(recipient.address, mockTokenAddress, amount);

      expect(await mockToken.balanceOf(recipient.address)).to.equal(amount);
      expect(await mockToken.balanceOf(await treasury.getAddress())).to.equal(ethers.parseEther("50"));
      expect(await treasury.getBalance(mockTokenAddress)).to.equal(ethers.parseEther("50"));
    });

    it("Should fail if non-timelock tries to withdraw", async function () {
      const amount = ethers.parseEther("50");
      await expect(
        treasury.connect(user).withdraw(recipient.address, await mockToken.getAddress(), amount)
      ).to.be.revertedWithCustomError(treasury, "AccessControlUnauthorizedAccount");
    });

    it("Should fail if insufficient token balance", async function () {
      const amount = ethers.parseEther("200");
      await expect(
        treasury.withdraw(recipient.address, await mockToken.getAddress(), amount)
      ).to.be.revertedWith("Insufficient token balance");
    });
  });


  describe("Access Control", function () {
    it("Should have correct TIMELOCK_ROLE", async function () {
      const TIMELOCK_ROLE = await treasury.TIMELOCK_ROLE();
      expect(await treasury.hasRole(TIMELOCK_ROLE, owner.address)).to.be.true;
      expect(await treasury.hasRole(TIMELOCK_ROLE, user.address)).to.be.false;
    });

    it("Should verify TIMELOCK_ROLE is granted to owner", async function () {
      const TIMELOCK_ROLE = await treasury.TIMELOCK_ROLE();
      expect(await treasury.hasRole(TIMELOCK_ROLE, owner.address)).to.be.true;
    });

    it("Should verify user doesn't have TIMELOCK_ROLE", async function () {
      const TIMELOCK_ROLE = await treasury.TIMELOCK_ROLE();
      expect(await treasury.hasRole(TIMELOCK_ROLE, user.address)).to.be.false;
    });
  });
});
