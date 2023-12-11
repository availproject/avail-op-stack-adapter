#!/usr/bin/env bash

# This script is used to generate the four wallets that are used in the Getting
# Started quickstart guide on the docs site. Simplifies things for users
# slightly while also avoiding the need for users to manually copy/paste a
# bunch of stuff over to the environment file.

# Generate wallets
wallet1=$(cast wallet new)
wallet2=$(cast wallet new)
wallet3=$(cast wallet new)
wallet4=$(cast wallet new)

# Grab wallet addresses
AVL_OP_ADMIN_ADDRESS=$(echo "$wallet1" | awk '/Address/ { print $2 }')
AVL_OP_BATCHER_ADDRESS=$(echo "$wallet2" | awk '/Address/ { print $2 }')
AVL_OP_PROPOSER_ADDRESS=$(echo "$wallet3" | awk '/Address/ { print $2 }')
AVL_OP_SEQUENCER_ADDRESS=$(echo "$wallet4" | awk '/Address/ { print $2 }')

# Grab wallet private keys
AVL_OP_ADMIN_PRIVATE_KEY=$(echo "$wallet1" | awk '/Private key/ { print $3 }')
AVL_OP_BATCHER_PRIVATE_KEY=$(echo "$wallet2" | awk '/Private key/ { print $3 }')
AVL_OP_PROPOSER_PRIVATE_KEY=$(echo "$wallet3" | awk '/Private key/ { print $3 }')
AVL_OP_SEQUENCER_PRIVATE_KEY=$(echo "$wallet4" | awk '/Private key/ { print $3 }')

# Generate the config file
wallets=$(cat << EOL
{

  "Admin" : {
    "Address":     "$AVL_OP_ADMIN_ADDRESS",
    "Private key": "$AVL_OP_ADMIN_PRIVATE_KEY"
  },
  "Batcher" : {
    "Address":     "$AVL_OP_BATCHER_ADDRESS",
    "Private key": "$AVL_OP_BATCHER_PRIVATE_KEY"
  },
  "Proposer" : {
    "Address":     "$AVL_OP_PROPOSER_ADDRESS",
    "Private key": "$AVL_OP_PROPOSER_PRIVATE_KEY"
  },
  "Sequencer" : {
    "Address":     "$AVL_OP_SEQUENCER_ADDRESS",
    "Private key": "$AVL_OP_SEQUENCER_PRIVATE_KEY"
  }
}
EOL
)

# Check if directory already exists,
# if it doesnt, create one.
if [ ! -d "./.testnet" ]; then
        mkdir ".testnet"
else
        rm -rf ".testnet"
        mkdir ".testnet"
fi

# Write the config file
echo "$wallets" > ./.testnet/avail-op-testnet-wallets.json









#!/usr/bin/env bash

# This script is used to generate the getting-started.json configuration file
# used in the Getting Started quickstart guide on the docs site. Avoids the
# need to have the getting-started.json committed to the repo since it's an
# invalid JSON file when not filled in, which is annoying.


# Get the finalized block timestamp and hash
L1_RPC_URL="https://eth-goerli.g.alchemy.com/v2/zmvISxoaiFivtoDZHrkB_GEvzBwH5NLT"
block=$(cast block finalized --rpc-url $L1_RPC_URL)
timestamp=$(echo "$block" | awk '/timestamp/ { print $2 }')
blockhash=$(echo "$block" | awk '/hash/ { print $2 }')

# Generate the config file
config=$(cat << EOL
{
  "numDeployConfirmations": 1,
  "finalSystemOwner": "$AVL_OP_ADMIN_ADDRESS",
  "portalGuardian": "$AVL_OP_ADMIN_ADDRESS",
  "controller": "$AVL_OP_ADMIN_ADDRESS",
  "l1StartingBlockTag": "$blockhash",
  "l1ChainID": 5,
  "l2ChainID": 42069,
  "l2BlockTime": 5,
  "maxSequencerDrift": 600,
  "sequencerWindowSize": 3600,
  "channelTimeout": 300,
  "p2pSequencerAddress": "$AVL_OP_SEQUENCER_ADDRESS",
  "batchInboxAddress": "0xff00000000000000000000000000000000042069",
  "batchSenderAddress": "$AVL_OP_BATCHER_ADDRESS",
  "l2OutputOracleSubmissionInterval": 120,
  "l2OutputOracleStartingBlockNumber": 0,
  "l2OutputOracleStartingTimestamp": $timestamp,
  "l2OutputOracleProposer": "$AVL_OP_PROPOSER_ADDRESS",
  "l2OutputOracleChallenger": "$AVL_OP_ADMIN_ADDRESS",
  "finalizationPeriodSeconds": 12,
  "proxyAdminOwner": "$AVL_OP_ADMIN_ADDRESS",
  "baseFeeVaultRecipient": "$AVL_OP_ADMIN_ADDRESS",
  "l1FeeVaultRecipient": "$AVL_OP_ADMIN_ADDRESS",
  "sequencerFeeVaultRecipient": "$AVL_OP_ADMIN_ADDRESS",
  "baseFeeVaultMinimumWithdrawalAmount": "0x8ac7230489e80000",
  "l1FeeVaultMinimumWithdrawalAmount": "0x8ac7230489e80000",
  "sequencerFeeVaultMinimumWithdrawalAmount": "0x8ac7230489e80000",
  "baseFeeVaultWithdrawalNetwork": 0,
  "l1FeeVaultWithdrawalNetwork": 0,
  "sequencerFeeVaultWithdrawalNetwork": 0,
  "gasPriceOracleOverhead": 2100,
  "gasPriceOracleScalar": 1000000,
  "enableGovernance": true,
  "governanceTokenSymbol": "OP",
  "governanceTokenName": "Optimism",
  "governanceTokenOwner": "$AVL_OP_ADMIN_ADDRESS",
  "l2GenesisBlockGasLimit": "0x1c9c380",
  "l2GenesisBlockBaseFeePerGas": "0x3b9aca00",
  "l2GenesisRegolithTimeOffset": "0x0",
  "eip1559Denominator": 50,
  "eip1559Elasticity": 10,
  "enableDA": true
}
EOL
)

# Write the config file
echo "$config" > ./packages/contracts-bedrock/deploy-config/avail-op-testnet.json
