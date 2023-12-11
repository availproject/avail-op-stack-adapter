const fs = require('fs')

const { ethers } = require('ethers')

const delay = (ms) => {
  return new Promise(resolve => setTimeout(resolve, ms))
}

// Read the JSON configuration
const walletsRaw = fs.readFileSync(
  './.testnet/avail-op-testnet-wallets.json',
  'utf-8'
)
const wallets = JSON.parse(walletsRaw)

const main = async () => {

  const privateKey = process.env.PRIVATE_KEY
  const L1_RPC_URL = process.env.L1_RPC_URL
  const amount = process.env.AMOUNT

  if (!privateKey || !L1_RPC_URL || !amount) {
    throw new Error('Required environment variable not found')
  }

  const l1Provider = new ethers.providers.JsonRpcProvider(L1_RPC_URL)

  const l1Signer = new ethers.Wallet(privateKey).connect(l1Provider)

  ////////////////////////////////////////////////
  ///       Funding batcher and proposer      ///
  //////////////////////////////////////////////
  console.log('Funding batcher accounts on L1 chain with 0.1 ETH')
  const tx1 = await l1Signer.sendTransaction({
    to: wallets.Batcher.Address,
    value: ethers.utils.parseEther('0.1'),
  })
  console.log(`Transaction hash on L1 chain: ${tx1.hash}`)
  const receipt1 = await tx1.wait()
  console.log(
    `Transaction was mined in block ${receipt1.blockNumber} on L1 chain`
  )

  console.log('Funding proposer accounts on L1 chain with 0.1 ETH')
  const tx2 = await l1Signer.sendTransaction({
    to: wallets.Proposer.Address,
    value: ethers.utils.parseEther('0.1'),
  })
  console.log(`Transaction hash on L1 chain: ${tx2.hash}`)
  const receipt2 = await tx2.wait()
  console.log(
    `Transaction was mined in block ${receipt2.blockNumber} on L1 chain`
  )

  console.log('Funding admin accounts on L1 chain with 0.1 ETH')
  const tx3 = await l1Signer.sendTransaction({
    to: wallets.Admin.Address,
    value: ethers.utils.parseEther('0.1'),
  })
  console.log(`Transaction hash on L1 chain: ${tx3.hash}`)
  const receipt3 = await tx3.wait()
  console.log(
    `Transaction was mined in block ${receipt3.blockNumber} on L1 chain`
  )

  console.log('Funding sequencer accounts on L1 chain with 0.1 ETH')
  const tx4 = await l1Signer.sendTransaction({
    to: wallets.Sequencer.Address,
    value: ethers.utils.parseEther('0.1'),
  })
  console.log(`Transaction hash on L1 chain: ${tx4.hash}`)
  const receipt4 = await tx4.wait()
  console.log(
    `Transaction was mined in block ${receipt4.blockNumber} on L1 chain`
  )

}

main()
  .then(() => process.exit(0))
  .catch(error => {
    console.error(error)
    process.exit(1)
  })
