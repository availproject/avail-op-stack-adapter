// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

// TODO(maurelian): remove this when the contract is complete
import { console } from "forge-std/console.sol";

contract DelayedVetoable {
    event Forwarded(bytes data);

    address internal _target;

    constructor(address target) {
        _target = target;
    }

    receive() external payable {
        _handleCall();
    }

    fallback() external payable {
        _handleCall();
    }

    function _handleCall() internal {
        require(_target != address(0), "DelayedVetoable: target not initialized");

        address target = _target;
        // bytes32 topic = keccak256("Forwarded(bytes)");
        emit Forwarded(msg.data);
        assembly {
            // Copy calldata into memory at 0x0....calldatasize.
            calldatacopy(0x0, 0x0, calldatasize())

            // TODO(maurelian): can we emit this in the assembly block to deduplicate
            //   getting the calldata? I think I was doing it correctly, but the forge
            //   test doesn't recognize this
            // log1(0x0, calldatasize(), topic)

            // Perform the call, make sure to pass all available gas.
            let success := call(gas(), target, callvalue(), 0x0, calldatasize(), 0x0, 0x0)

            // Copy returndata into memory at 0x0....returndatasize. Note that this *will*
            // overwrite the calldata that we just copied into memory but that doesn't really
            // matter because we'll be returning in a second anyway.
            returndatacopy(0x0, 0x0, returndatasize())

            // Success == 0 means a revert. We'll revert too and pass the data up.
            if iszero(success) { revert(0x0, returndatasize()) }

            // Otherwise we'll just return and pass the data up.
            return(0x0, returndatasize())
        }
    }
}
