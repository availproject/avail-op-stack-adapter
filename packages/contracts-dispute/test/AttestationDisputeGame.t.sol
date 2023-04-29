// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import "forge-std/Test.sol";

import "src/types/Errors.sol";
import "src/types/Types.sol";

import { LibClock } from "src/lib/LibClock.sol";
import { LibHashing } from "src/lib/LibHashing.sol";
import { LibPosition } from "src/lib/LibPosition.sol";

import { ResourceMetering } from "contracts-bedrock/L1/ResourceMetering.sol";
import { SystemConfig } from "contracts-bedrock/L1/SystemConfig.sol";
import { L2OutputOracle } from "contracts-bedrock/L1/L2OutputOracle.sol";

import { AttestationDisputeGame } from "src/AttestationDisputeGame.sol";
import { IDisputeGameFactory } from "src/interfaces/IDisputeGameFactory.sol";
import { IDisputeGame } from "src/interfaces/IDisputeGame.sol";
import { IBondManager } from "src/interfaces/IBondManager.sol";
import { BondManager } from "src/BondManager.sol";
import { DisputeGameFactory } from "src/DisputeGameFactory.sol";

/// @title AttestationDisputeGame_Test
contract AttestationDisputeGame_Test is Test {
    DisputeGameFactory factory;
    BondManager bm;
    AttestationDisputeGame disputeGameImplementation;
    SystemConfig systemConfig;
    L2OutputOracle l2oo;
    AttestationDisputeGame disputeGameProxy;

    // L2OutputOracle Constructor arguments
    address internal proposer = 0x000000000000000000000000000000000000AbBa;
    address internal owner = 0x000000000000000000000000000000000000ACDC;
    uint256 internal submissionInterval = 1800;
    uint256 internal l2BlockTime = 2;
    uint256 internal startingBlockNumber = 200;
    uint256 internal startingTimestamp = 1000;

    /// @notice Emitted when a new dispute game is created by the [DisputeGameFactory]
    event DisputeGameCreated(address indexed disputeProxy, GameType indexed gameType, Claim indexed rootClaim);

    function setUp() public {
        factory = new DisputeGameFactory(address(this));
        vm.label(address(factory), "DisputeGameFactory");
        bm = new BondManager(factory);
        vm.label(address(bm), "BondManager");

        ResourceMetering.ResourceConfig memory _config = ResourceMetering.ResourceConfig({
            maxResourceLimit: 1000000000,
            elasticityMultiplier: 2,
            baseFeeMaxChangeDenominator: 2,
            minimumBaseFee: 10,
            systemTxMaxGas: 100000000,
            maximumBaseFee: 1000
        });

        systemConfig = new SystemConfig(
            address(this), // _owner,
            100, // _overhead,
            100, // _scalar,
            keccak256("BATCHER.HASH"), // _batcherHash,
            uint64(100000000), // _gasLimit,
            address(0), // _unsafeBlockSigner,
            _config
        );
        vm.label(address(systemConfig), "SystemConfig");

        l2oo = new L2OutputOracle({
            _submissionInterval: submissionInterval,
            _l2BlockTime: l2BlockTime,
            _startingBlockNumber: startingBlockNumber,
            _startingTimestamp: block.timestamp + 1,
            _proposer: proposer,
            _challenger: owner,
            _finalizationPeriodSeconds: 7 days
        });
        vm.label(address(l2oo), "L2OutputOracle");

        // Create the dispute game implementation
        disputeGameImplementation = new AttestationDisputeGame(IBondManager(address(bm)), systemConfig, l2oo);
        vm.label(address(disputeGameImplementation), "AttestationDisputeGame_Implementation");

        // Set the implementation in the factory
        GameType gt = GameType.ATTESTATION;
        factory.setImplementation(gt, IDisputeGame(address(disputeGameImplementation)));

        // Create the attestation dispute game in the factory
        bytes memory extraData = bytes("");
        Claim rootClaim = Claim.wrap(bytes32(""));
        vm.expectEmit(false, true, true, false);
        emit DisputeGameCreated(address(0), gt, rootClaim);
        disputeGameProxy = AttestationDisputeGame(address(factory.create(gt, rootClaim, extraData)));
        assertEq(address(factory.games(gt, rootClaim, extraData)), address(disputeGameProxy));
        vm.label(address(disputeGameProxy), "AttestationDisputeGame_Proxy");
    }

    function test_defaultInitialization_succeeds() public {
        uint256 _signatureThreshold = disputeGameProxy.signatureThreshold();
        assertEq(_signatureThreshold, 1);
    }

}
