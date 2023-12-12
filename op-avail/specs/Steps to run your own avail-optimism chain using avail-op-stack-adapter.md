# Steps to run your own avail-optimism chain using avail-op-stack-adapter

# Overview

Hello! This guide is meant to help you kick off your Avail-OP Stack journey by taking you through the process of spinning up your very own avail-optimism chain on the Ethereum Goerli testnet as settlement layer and Avail kate testnet as Data Avalibility layer.

# Know before you go

Before we kick off, note that this is a relatively long tutorial! You should prepare to set aside an hour or two to get everything running. Here’s an itemized list of what we’re about to do:

1. Install dependencies
2. Build the source code
3. Generate and fund accounts and private keys
4. Configure your network
5. Deploy the L1 contracts
6. Initialize op-geth
7. Run op-geth
8. Run op-node
9. Get some Goerli ETH on your L2
10. Send some test transactions
11. Celebrate!

# Prerequisites

You’ll need the following software installed to follow this tutorial:

- **[Git](https://git-scm.com/)**
- **[Go](https://go.dev/)**
- **[Node](https://nodejs.org/en/)**
- **[Pnpm](https://classic.yarnpkg.com/lang/en/docs/install/)**
- **[Foundry](https://github.com/foundry-rs/foundry#installation)**
- **[Make](https://linux.die.net/man/1/make)**
- **[jq](https://github.com/jqlang/jq)**
- **[direnv](https://direnv.net/)**

This tutorial was checked on:

| Software | Version | Installation command(s) |
| --- | --- | --- |
| Ubuntu | 20.04 LTS |  |
| git, curl, jq, and make | OS default | sudo apt install -y git curl make jq |
| Go | 1.20 | sudo apt updatewget <https://go.dev/dl/go1.20.linux-amd64.tar.gztar> xvzf go1.20.linux-amd64.tar.gzsudo cp go/bin/go /usr/bin/gosudo mv go /usr/libecho export GOROOT=/usr/lib/go >> ~/.bashrc |
| Node | 16.19.0 | curl -fsSL <https://deb.nodesource.com/setup_16.x> | sudo -E bash -sudo apt-get install -y nodejs npm |
| pnpm | 8.5.6 | sudo npm install -g pnpm |
| Foundry | 0.2.0 | yarn install:foundry |

# Build the Source Code

We’re going to be spinning up an EVM Rollup from the Avail-OP-Stack-Adapter source code. The Avail-Optimism is split between two repositories, the **[Avail-OP-Stack Adapter](https://github.com/availproject/avail-op-stack-adapter) monorepo** and the **[op-geth](https://github.com/ethereum-optimism/op-geth)** repository.

# Build the Optimism Monorepo

1. Clone the [Avail-OP-Stack Adapter](https://github.com/availproject/avail-optimism).

    ```bash
    cd ~
    git clone https://github.com/availproject/avail-op-stack-adapter.git
    ```

2. Enter the Avail-OP-Stack Monorepo.

    ```bash
    cd avail-op-stack-adapter
    ```

3. Install required modules. This is a slow process, while it is running you can already start building `op-geth`, as shown below.

    ```bash
    pnpm install
    ```

4. Build the various packages inside of the Optimism Monorepo.

    ```bash
    make op-node op-batcher op-proposer
    pnpm build
    ```

# Build op-geth

1. Clone **[op-geth](https://github.com/ethereum-optimism/op-geth)** :

    ```bash
    cd ~
    git clone https://github.com/ethereum-optimism/op-geth.git
    ```

2. Enter `op-geth`:

    ```bash
    cd op-geth
    ```

3. Build `op-geth`:

    ```bash
    make geth
    ```

# Get access to a Goerli node

Since we’re deploying our Avail-Optimism chain to Goerli, you’ll need to have access to a Goerli L1 node. You can either use a node provider like **[Alchemy](https://www.alchemy.com/)** (easier) or **[run your own Goerli node](https://notes.ethereum.org/@launchpad/goerli)** (harder).

# Generate some keys

You’ll need four accounts and their private keys when setting up the chain:

- The `Admin` account which has the ability to upgrade contracts.
- The `Batcher` account which publishes Sequencer transaction data to L1.
- The `Proposer` account which publishes L2 transaction results to L1.
- The `Sequencer` account which signs blocks on the p2p network.

You can generate all of these keys with the `rekey` tool in the `contracts-bedrock` package.

1. Enter the Avail-Optimism Monorepo:

    ```bash
    cd avail-op-stack-adapter
    ```

2. Move into the `contracts-bedrock` package:

    ```bash
    cd packages/contracts-bedrock
    ```

3. Use `cast wallet` to generate new accounts

    ```bash
    echo "Admin:"
    cast wallet new
    echo "Proposer:"
    cast wallet new
    echo "Batcher:"
    cast wallet new
    echo "Sequencer:"
    cast wallet new
    ```

You should get an output like the following:

```
Admin:
Successfully created new keypair.
Address:     0x9f92bdF0db69264462FC305913960Edfcc7a7c7F
Private key: 0x30e66956e1a12b81f0f2cfb982286b2f566eb73649833831d9f80b12f8fa183c
Proposer:
Successfully created new keypair.
Address:     0x31dE9B6473fc47af36ec23878bA34824B9F4AB30
Private key: 0x8bd1c8dfffef880f8f9ab8162f97ccd119c1aac28fe00dacf919459f88e0f37d
Batcher:
Successfully created new keypair.
Address:     0x6A3DC843843139f17Fcf04C057bb536A421DC9c6
Private key: 0x3ce44144b7fde797a28f4e47b210a4d42c3a3b642e538b54458cba2740db5ac2
Sequencer:
Successfully created new keypair.
Address:     0x98C6cadB1fe77aBB7bD968fC3E9b206111e72848
Private key: 0x3f4241229bb6f155140d98e0f5dd2aad7ae983f5af5d61555d05eb8e5d9514db
```

Save these accounts and their respective private keys somewhere, you’ll need them later. Fund the `Admin` address with a small amount of Goerli ETH as we’ll use that account to deploy our smart contracts. You’ll also need to fund the `Proposer` and `Batcher` address — note that the `Batcher` burns through the most ETH because it publishes transaction data to L1.

Recommended funding amounts are as follows:

- `Admin` — 2 ETH
- `Proposer` — 5 ETH
- `Batcher` — 10 ETH

**Not for production deployments**

The `cast wallet new` tool is *not* designed for production deployments. If you are deploying an Avail-Optimism Stack based chain into production, you should likely be using a combination of hardware security modules and hardware wallets.

# Configure your network

Once you’ve built both repositories, you’ll need head back to the Avail-Op-Stack-Adapter Monorepo to set up the configuration for your chain. Currently, chain configuration lives inside of the **[contracts-bedrock](https://github.com/availproject/avail-optimism/tree/129032f15b76b0d2a940443a39433de931a97a44/packages/contracts-bedrock)** package.

1. Enter the Avail-OP-Stack Monorepo:

    ```bash
    cd ~/avail-op-stack-adapter
    ```

2. Move into the `contracts-bedrock` package:

    ```bash
    cd packages/contracts-bedrock
    ```

3. Inside of `contracts-bedrock`, copy the environment file

    ```
    cp .envrc.example .envrc
    ```

4. Fill out the environment variables inside of that file:
    - `ETH_RPC_URL` — URL for your L1 node.
    - `PRIVATE_KEY` — Private key of the `Admin` account.
    - `DEPLOYMENT_CONTEXT` - Name of the network, should be "avail-optimism"
5. Pull the environment variables into context using `direnv`

    ```bash
    direnv allow .
    ```

    If you need to install `direnv`, **[make sure you also modify the shell configuration](https://direnv.net/docs/hook.html)**.

6. Before we can create our configuration file, we’ll need to pick an L1 block to serve as the starting point for our Rollup. It’s best to use a finalized L1 block as our starting block. You can use the `cast` command provided by Foundry to grab all of the necessary information:

    ```bash
    cast block finalized --rpc-url $ETH_RPC_URL | grep -E "(timestamp|hash|number)"
    ```

    You’ll get back something that looks like the following:

    ```
    hash                 0x784d8e7f0e90969e375c7d12dac7a3df6879450d41b4cb04d4f8f209ff0c4cd9
    number               8482289
    timestamp            1676253324
    ```

7. Create a copy of file `deploy-config/getting-started.json` at `deploy-config` named `avail-optimism.json`
8. Fill out the remainder of the pre-populated config file **`deploy-config/avail-optimism.json`**. Use the default values in the config file and make following modifications:
    - Replace `"ADMIN"` with the address of the Admin account you generated earlier.
    - Replace `"PROPOSER"` with the address of the Proposer account you generated earlier.
    - Replace `"BATCHER"` with the address of the Batcher account you generated earlier.
    - Replace `"SEQUENCER"` with the address of the Sequencer account you generated earlier.
    - Replace `"BLOCKHASH"` with the blockhash you got from the `cast` command.
    - Replace `TIMESTAMP` with the timestamp you got from the `cast` command. Note that although all the other fields are strings, this field is a number! Don’t include the quotation marks.
9. Check the value of `enableDA` config variable should be `true` to use Avail chain as DA or you can set it to `false` for using ethereum as DA
10. Move into `op-avail` module

    ```bash
    cd ~/avail-op-stack-adapter/op-avail
    ```

11. Add a `config.json` file to this directory with these variables and change the `seed` and `app_id` according to your avail account

    ```
    {
      "seed": "test test test test test test test test test test test avail",
      "api_url": "wss://kate.avail.tools:443/ws",
      "app_id": 1
    }
    ```

# Deploy the L1 contracts

Once you’ve configured your network, it’s time to deploy the L1 smart contracts necessary for the functionality of the chain.

1. Create a `avail-optimism` deployment directory.

    ```bash
    mkdir deployments/avail-optimism
    ```

2. Once you’re ready, deploy the L1 smart contracts.

    ```bash
    forge script scripts/Deploy.s.sol:Deploy --private-key $PRIVATE_KEY --broadcast --rpc-url $ETH_RPC_URL
    forge script scripts/Deploy.s.sol:Deploy --sig 'sync()' --private-key $PRIVATE_KEY --broadcast --rpc-url $ETH_RPC_URL

    ```

Contract deployment can take up to 15 minutes. Please wait for all smart contracts to be fully deployed before continuing to the next step.

# Generate the L2 config files

We’ve set up the L1 side of things, but now we need to set up the L2 side of things. We do this by generating three important files, a `genesis.json` file, a `rollup.json` configuration file, and a `jwt.txt` **[JSON Web Token](https://jwt.io/introduction)** that allows the `op-node` and `op-geth` to communicate securely.

1. Head over to the `op-node` package.

    ```bash
    cd ~/avail-op-stack-adapter/op-node
    ```

2. Run the following command, and make sure to replace `<RPC>` with your L1 RPC URL:

    ```bash
    go run cmd/main.go genesis l2 \
        --deploy-config ../packages/contracts-bedrock/deploy-config/avail-optimism.json \
        --deployment-dir ../packages/contracts-bedrock/deployments/avail-optimism/ \
        --outfile.l2 genesis.json \
        --outfile.rollup rollup.json \
        --l1-rpc=$L1_RPC
    ```

    You should then see the `genesis.json` and `rollup.json` files inside the `op-node` package.

3. Next, generate the `jwt.txt` file with the following command:

    ```bash
    openssl rand -hex 32 > jwt.txt
    ```

4. Finally, we’ll need to copy the `genesis.json` file and `jwt.txt` file into `op-geth` so we can use it to initialize and run `op-geth` in just a minute:

    ```bash
    cp genesis.json ~/op-geth
    cp jwt.txt ~/op-geth
    ```

# Initialize op-geth

We’re almost ready to run our chain! Now we just need to run a few commands to initialize `op-geth`. We’re going to be running a Sequencer node, so we’ll need to import the `Sequencer` private key that we generated earlier. This private key is what our Sequencer will use to sign new blocks.

1. Head over to the `op-geth` repository:

    ```bash
    cd ~/op-geth
    ```

2. Create a data directory folder:

    ```bash
    mkdir datadir
    ```

3. Next we need to initialize `op-geth` with the genesis file we generated and copied earlier:

    ```bash
    build/bin/geth init --datadir=datadir genesis.json
    ```

Everything is now initialized and ready to go!

# Run the node software

There are four components that need to run for a rollup. The first two, `op-geth` and `op-node`, have to run on every node. The other two, `op-batcher` and `op-proposer`, run only in one place, the sequencer that accepts transactions.

Set these environment variables for the configuration

| Variable | Value |
| --- | --- |
| SEQ_KEY | Private key of the Sequencer account |
| BATCHER_KEY | Private key of the Batcher accounts, which should have at least 1 ETH |
| PROPOSER_KEY | Private key of the Proposer account |
| L1_RPC | URL for the L1 (such as Goerli) you're using |
| RPC_KIND | The type of L1 server to which you connect, which can optimize requests. Available options are alchemy, quicknode, parity, nethermind, debug_geth, erigon, basic, and any |
| L2OO_ADDR | The address of the L2OutputOracleProxy, available at ~/optimism/packages/contracts-bedrock/deployments/getting-started/L2OutputOracleProxy.json |

# `op-geth`

Run `op-geth` with the following commands.

```bash
cd ~/op-geth

./build/bin/geth \
        --datadir ./datadir \
        --http \
        --http.corsdomain="*" \
        --http.vhosts="*" \
        --http.addr=0.0.0.0 \
        --http.port=9545 \
        --http.api=web3,debug,eth,txpool,net,engine \
        --ws \
        --ws.addr=0.0.0.0 \
        --ws.port=9546 \
        --ws.origins="*" \
        --ws.api=debug,eth,txpool,net,engine \
        --syncmode=full \
        --gcmode=archive \
        --nodiscover \
        --maxpeers=0 \
        --networkid=42069 \
        --authrpc.vhosts="*" \
        --authrpc.addr=0.0.0.0 \
        --authrpc.port=9551 \
        --authrpc.jwtsecret=./jwt.txt \
        --rollup.disabletxpoolgossip=true
```

And `op-geth` should be running! You should see some output, but you won’t see any blocks being created yet because `op-geth` is driven by the `op-node`. We’ll need to get that running next.

**Why archive mode?**

Archive mode takes more disk storage than full mode. However, using it is important for two reasons:

- The `op-proposer` requires access to the full state. If at some point `op-proposer` needs to look beyond 256 blocks in the past (8.5 minutes in the default configuration), for example because it was down for that long, we need archive mode.
- The **explorer** requires archive mode.

### Reinitializing op-geth

There are several situations are indicate database corruption and require you to reset the `op-geth` component:

- When `op-node` errors out when first started and exits.
- When `op-node` emits this error:

    ```
    stage 0 failed resetting: temp: failed to find the L2 Heads to start from: failed to fetch L2 block by hash 0x0000000000000000000000000000000000000000000000000000000000000000
    ```

This is the reinitialization procedure:

1. Stop the `op-geth` process.
2. Delete the geth data.

    ```bash
    cd ~/op-geth
    rm -rf datadir/geth
    ```

3. Rerun init.

    ```bash
    build/bin/geth init --datadir=datadir genesis.json
    ```

4. Start `op-geth`
5. Start `op-node`

# `op-node`

Once we’ve got `op-geth` running we’ll need to run `op-node`. Like Ethereum, the OP Stack has a consensus client (the `op-node`) and an execution client (`op-geth`). The consensus client drives the execution client over the Engine API.

```bash
cd ~/optimism/op-node

./bin/op-node \
 --l2=http://localhost:9551 \
 --l2.jwt-secret=./jwt.txt \
 --sequencer.enabled \
 --sequencer.l1-confs=3 \
 --verifier.l1-confs=3 \
 --rollup.config=./rollup.json \
 --rpc.addr=0.0.0.0 \
 --rpc.port=9547 \
 --p2p.disable \
 --rpc.enable-admin \
 --p2p.sequencer.key=$SEQ_KEY \
 --l1=$L1_RPC \
 --l1.rpckind=$RPC_KIND
```

Once you run this command, you should start seeing the `op-node` begin to process all of the L1 information after the starting block number that you picked earlier. Once the `op-node` has enough information, it’ll begin sending Engine API payloads to `op-geth`. At that point, you’ll start to see blocks being created inside of `op-geth`. We’re live!

**Peer to peer synchronization**

If you use a chain ID that is also used by others, for example the default (42069), your `op-node` will try to use peer to peer to speed up synchronization. These attempts will fail, because they will be signed with the wrong key, but they will waste time and network resources.

To avoid this , we start with peer to peer synchronization disabled (`--p2p.disable`). Once you have multiple nodes, it makes sense to use these command line parameters to synchronize between them without getting confused by other blockchains.

```
 --p2p.static=<nodes> \
 --p2p.listen.ip=0.0.0.0 \
 --p2p.listen.tcp=9003 \
 --p2p.listen.udp=9003 \
```

# `op-batcher`

The `op-batcher` takes transactions from the Sequencer and publishes those transactions to L1. Once transactions are on L1, they’re officially part of the Rollup. Without the `op-batcher`, transactions sent to the Sequencer would never make it to L1 and wouldn’t become part of the canonical chain. The `op-batcher` is critical!

It is best to give the `Batcher` at least 1 Goerli ETH to ensure that it can continue operating without running out of ETH for gas.

```bash
cd ~/optimism/op-batcher

./bin/op-batcher \
    --l2-eth-rpc=http://localhost:9545 \
    --rollup-rpc=http://localhost:9547 \
    --poll-interval=10s \
    --sub-safety-margin=6 \
    --num-confirmations=1 \
    --safe-abort-nonce-too-low-count=3 \
    --resubmission-timeout=30s \
    --rpc.addr=0.0.0.0 \
    --rpc.port=9548 \
    --rpc.enable-admin \
    --max-channel-duration=1 \
    --l1-eth-rpc=$L1_RPC \
    --private-key=$BATCHER_KEY
```

**Controlling batcher costs**

The `--max-channel-duration=n` setting tells the batcher to write all the data to L1 every `n` L1 blocks. When it is low, transactions are written to L1 frequently, withdrawals are quick, and other nodes can synchronize from L1 fast. When it is high, transactions are written to L1 less frequently, and the batcher spends less ETH.

# `op-proposer`

Now start `op-proposer`, which proposes new state roots.

```bash
cd ~/optimism/op-proposer

./bin/op-proposer \
    --poll-interval=12s \
    --rpc.port=9560 \
    --rollup-rpc=http://localhost:9547 \
    --l2oo-address=$L2OO_ADDR \
    --private-key=$PROPOSER_KEY \
    --l1-eth-rpc=$L1_RPC
```

# Get some ETH on your Rollup

Once you’ve connected your wallet, you’ll probably notice that you don’t have any ETH on your Rollup. You’ll need some ETH to pay for gas on your Rollup. The easiest way to deposit Goerli ETH into your chain is to send funds directly to the `L1StandardBridge` contract. You can find the address of the `L1StandardBridge` contract for your chain by looking inside the `deployments` folder in the `contracts-bedrock` package.

1. First, head over to the `contracts-bedrock` package:

    ```bash
    cd ~/optimism/packages/contracts-bedrock
    ```

2. Grab the address of the proxy to the L1 standard bridge contract:

    ```bash
    cat deployments/avail-optimism/L1StandardBridgeProxy.json | jq -r .address
    ```

3. Grab the L1 bridge proxy contract address and, using the wallet that you want to have ETH on your Rollup, send that address a small amount of ETH on Goerli (0.1 or less is fine). It may take up to 5 minutes for that ETH to appear in your wallet on L2.

# Hurray

**Congratulations, you made it! You now have a complete Avail-OP-Stack-Adapter based EVM Rollup.**

You can use this rollup the same way you’d use any other test blockchain.
