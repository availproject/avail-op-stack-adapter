const { Keyring } = require('@polkadot/api')
const { mnemonicGenerate } = require('@polkadot/util-crypto')




const pair = keyring.addFromUri(mnemonic)

console.log("Address: ",pair.address, "Seed: ", mnemonic)

// async function createKey(api: ApiPromise, sender: KeyringPair, id: string) {
//   try {
//       await api.tx.dataAvailability.createApplicationKey(id)
//           .signAndSend(
//               sender,
//               (result: ISubmittableResult) => {
//                   console.log(`Tx status: ${result.status}`);
//                   if (result.status.isFinalized) {
//                       let block_hash = result.status.asFinalized;
//                       let extrinsic_hash = result.txHash;
//                       console.log(`\nBlock finalized, extrinsic hash: ${extrinsic_hash}\nin block ${block_hash}`);
//                       process.exit(0);
//                   }
//               });
//   } catch (e) {
//       console.log(e);
//       process.exit(1);
//   }
// }

// async function main() {
//   const argv = await cli_arguments();
//   const api = await createApi();

//   const mnemonic = mnemonicGenerate()
//   const keyring = new Keyring({type: 'sr25519'});

//   const wallet = keyring.addFromUri('//Bob');

//   // creating api key
//   console.log(`Creating api key ${argv.i}`);
//   await createKey(api, bob, argv.i);
// }

// main().catch((err) => {
//   console.error(err);
//   process.exit(1);
// });
