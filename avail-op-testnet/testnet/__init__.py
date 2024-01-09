import argparse
import logging
import os
import subprocess
import json
import socket
import time
import shutil
import http.client


pjoin = os.path.join

parser = argparse.ArgumentParser(description='Avail OP Testnet Launcher')
parser.add_argument('--monorepo-dir', help='Directory of the monorepo', default=os.getcwd())
parser.add_argument('--deploy', help='Whether the contracts should be predeployed or deployed', type=bool, action=argparse.BooleanOptionalAction)

log = logging.getLogger()

class Bunch:
    def __init__(self, **kwds):
        self.__dict__.update(kwds)

def main():
    args = parser.parse_args()

    monorepo_dir = os.path.abspath(args.monorepo_dir)
    testnet_dir = pjoin(monorepo_dir, '.testnet')
    contracts_bedrock_dir = pjoin(monorepo_dir, 'packages', 'contracts-bedrock')
    deployment_dir = pjoin(contracts_bedrock_dir, 'deployments', 'avail-op-testnet')
    op_node_dir = pjoin(args.monorepo_dir, 'op-node')
    ops_optimium_dir=pjoin(monorepo_dir, 'ops-optimium')

    paths = Bunch(
      mono_repo_dir=monorepo_dir,
      testnet_dir=testnet_dir,
      contracts_bedrock_dir=contracts_bedrock_dir,
      deployment_dir=deployment_dir,
      deploy_config_dir=pjoin(contracts_bedrock_dir, 'deploy-config'),
      op_node_dir=op_node_dir,
      ops_optimium_dir=ops_optimium_dir,
      genesis_l1_path=pjoin(testnet_dir, 'genesis-l1.json'),
      genesis_l2_path=pjoin(testnet_dir, 'genesis-l2.json'),
      addresses_json_path=pjoin(testnet_dir, 'addresses.json'),
      sdk_addresses_json_path=pjoin(testnet_dir, 'sdk-addresses.json'),
      rollup_config_path=pjoin(testnet_dir, 'rollup.json')
    )

    os.makedirs(testnet_dir, exist_ok=True)

    # run_command(['docker-compose', 'build', '--progress', 'plain'], cwd=paths.ops_optimium_dir, env={
    #     'PWD': paths.ops_optimium_dir
    # })

    if args.deploy:
      log.info('Testnet with upcoming smart contract deployments')
      #generate_accounts_and_network_config(paths)
      #funding_avail_op_accounts(paths)
      testnet_deploy(paths)



def generate_accounts_and_network_config(paths):
    log.info("Generating accounts and network config")
    run_command(['bash', './avail-op-testnet/deploy.sh'], cwd=paths.mono_repo_dir, env={'PWD': paths.mono_repo_dir})

def funding_avail_op_accounts(paths):
    PRIVATE_KEY = os.environ['FUND_ACCOUNT_PRIVATE_KEY']
    L1_NODE_URL = os.environ['L1_NODE_URL']
    # PRIVATE_KEY = ''
    # L1_NODE_URL = ""
    AMOUNT = '0.1'
    run_command(['node', './avail-op-testnet/fund-accounts.js'], cwd=paths.mono_repo_dir, env={'PWD': paths.mono_repo_dir, 'PRIVATE_KEY': PRIVATE_KEY, 'L1_RPC_URL': L1_NODE_URL, 'AMOUNT': AMOUNT})

# Bring up the devnet where the contracts are deployed to L1
def testnet_deploy(paths):
    log.info('Starting L1.')
    L1_NODE_URL = os.environ["L1_RPC"]

    log.info('Fetching avail-op-chain accounts')
    testnet_wallets_orig = pjoin(paths.testnet_dir, 'avail-op-testnet-wallets.json')
    wallets = read_json(testnet_wallets_orig)

    fqn = 'scripts/Deploy.s.sol:Deploy'
    private_key = wallets['Admin']['Private key']

    if os.path.exists(paths.addresses_json_path):
        log.info('Contracts already deployed.')
        addresses = read_json(paths.addresses_json_path)
    else:
        log.info('Deploying contracts.')
        run_command([
            'forge', 'script', fqn, '--private-key', private_key,
            '--rpc-url', L1_NODE_URL, '--broadcast'
        ], env={"DEPLOYMENT_CONTEXT": "avail-op-testnet"}, cwd=paths.contracts_bedrock_dir)

        run_command([
            'forge', 'script', fqn, '--private-key', private_key,
            '--sig', 'sync()', '--rpc-url', L1_NODE_URL, '--broadcast'
        ], env={"DEPLOYMENT_CONTEXT": "avail-op-testnet"}, cwd=paths.contracts_bedrock_dir)

        contracts = os.listdir(paths.deployment_dir)
        addresses = {}
        for c in contracts:
            if not c.endswith('.json'):
                continue
            data = read_json(pjoin(paths.deployment_dir, c))
            addresses[c.replace('.json', '')] = data['address']
        sdk_addresses = {}
        sdk_addresses.update({
            'AddressManager': '0x0000000000000000000000000000000000000000',
            'StateCommitmentChain': '0x0000000000000000000000000000000000000000',
            'CanonicalTransactionChain': '0x0000000000000000000000000000000000000000',
            'BondManager': '0x0000000000000000000000000000000000000000',
        })

        sdk_addresses['L1CrossDomainMessenger'] = addresses['L1CrossDomainMessengerProxy']
        sdk_addresses['L1StandardBridge'] = addresses['L1StandardBridgeProxy']
        sdk_addresses['OptimismPortal'] = addresses['OptimismPortalProxy']
        sdk_addresses['L2OutputOracle'] = addresses['L2OutputOracleProxy']
        write_json(paths.addresses_json_path, addresses)
        write_json(paths.sdk_addresses_json_path, sdk_addresses)
        log.info(f'Wrote sdk addresses to {paths.sdk_addresses_json_path}')


    testnet_cfg_orig = pjoin(paths.contracts_bedrock_dir, 'deploy-config', 'avail-op-testnet.json')
    testnet_cfg_backup = pjoin(paths.testnet_dir, 'avail-op-testnet.json.bak')


    if os.path.exists(paths.genesis_l2_path):
        log.info('L2 genesis and rollup configs already generated.')
    else:
        log.info('Generating L2 genesis and rollup configs.')
        run_command([
            'go', 'run', 'cmd/main.go', 'genesis', 'l2',
            '--l1-rpc', L1_NODE_URL,
            '--deploy-config', testnet_cfg_orig,
            '--deployment-dir', paths.deployment_dir,
            '--outfile.l2', pjoin(paths.testnet_dir, 'genesis-l2.json'),
            '--outfile.rollup', pjoin(paths.testnet_dir, 'rollup.json')
        ], cwd=paths.op_node_dir)

    rollup_config = read_json(paths.rollup_config_path)

    if os.path.exists(testnet_cfg_backup):
        shutil.move(testnet_cfg_backup, testnet_cfg_orig)

    log.info('Bringing up op-geth(L2 execution enviornment)')
    run_command(['docker-compose', 'up', '-d', 'op-geth'], cwd=paths.ops_optimium_dir, env={
        'PWD': paths.ops_optimium_dir
    })
    wait_up(9545, wait_secs=120)
    wait_for_rpc_server('127.0.0.1:9545')

    log.info('Bringing up op-node(L2 consensus client)')
    run_command(['docker-compose', 'up', '-d', 'op-node'], cwd=paths.ops_optimium_dir, env={
        'PWD': paths.ops_optimium_dir,
        'L1_RPC_URL': L1_NODE_URL,
        'SEQ_PRIVATE_KEY': wallets['Sequencer']['Private key'].split('x')[1],
    })
    log.info("Its working")
    wait_up(7545, wait_secs=120)
    wait_for_node_server('127.0.0.1:7545')


    log.info('Bringing up everything else.')
    run_command(['docker-compose', 'up', '-d', 'op-proposer', 'op-batcher'], cwd=paths.ops_optimium_dir, env={
        'PWD': paths.ops_optimium_dir,
        'L1_RPC_URL': L1_NODE_URL,
        'BATCH_PRIVATE_KEY': wallets['Batcher']['Private key'].split('x')[1],
        'PROP_PRIVATE_KEY': wallets['Proposer']['Private key'].split('x')[1],
        'L2OO_ADDRESS': addresses['L2OutputOracleProxy'],
    })

    log.info('Testnet ready.')


def wait_for_rpc_server(url):
    log.info(f'Waiting for RPC server at {url}')

    conn = http.client.HTTPConnection(url)
    headers = {'Content-type': 'application/json'}
    body = '{"id":1, "jsonrpc":"2.0", "method": "eth_chainId", "params":[]}'

    while True:
        try:
            conn.request('POST', '/', body, headers)
            response = conn.getresponse()
            conn.close()
            if response.status < 300:
                log.info(f'RPC server at {url} ready')
                return
        except Exception as e:
            log.info(f'Waiting for RPC server at {url}')
            time.sleep(1)


def wait_for_node_server(url):
    log.info(f'Waiting for Node server at {url}')

    conn = http.client.HTTPConnection(url)
    headers = {'Content-type': 'application/json'}
    body = '{"jsonrpc":"2.0","method":"optimism_rollupConfig","params":[],"id":1}'

    while True:
        try:
            conn.request('POST', '/', body, headers)
            response = conn.getresponse()
            conn.close()
            if response.status < 300:
                log.info(f'RPC server at {url} ready')
                return
        except Exception as e:
            log.info(f'Waiting for RPC server at {url}')
            time.sleep(1)

def run_command(args, check=True, shell=False, cwd=None, env=None):
    env = env if env else {}
    return subprocess.run(
        args,
        check=check,
        shell=shell,
        env={
            **os.environ,
            **env
        },
        cwd=cwd
    )


def wait_up(port, retries=10, wait_secs=1):
    for i in range(0, retries):
        log.info(f'Trying 127.0.0.1:{port}')
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        try:
            s.connect(('127.0.0.1', int(port)))
            s.shutdown(2)
            log.info(f'Connected 127.0.0.1:{port}')
            return True
        except Exception:
            time.sleep(wait_secs)

    raise Exception(f'Timed out waiting for port {port}.')


def write_json(path, data):
    with open(path, 'w+') as f:
        json.dump(data, f, indent='  ')


def read_json(path):
    with open(path, 'r') as f:
        return json.load(f)
