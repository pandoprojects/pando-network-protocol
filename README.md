
# Pando Blockchain Ledger Protocol

  

The Pando Blockchain Ledger is a Proof-of-Stake decentralized ledger designed for the video streaming industry. It powers the Pando token economy which incentives end users to share their redundant bandwidth and storage resources, and encourage them to engage more actively with video platforms and content creators. The ledger employs a novel **multi-level BFT consensus engine**, which supports high transaction throughput, fast block confirmation, and allows mass participation in the consensus process. Off-chain payment support is built directly into the ledger through the resource-oriented micropayment pool, which is designed specifically to achieve the “pay-per-byte” granularity for streaming use cases. Moreover, the ledger storage system leverages the microservice architecture and reference counting based history pruning techniques, and is thus able to adapt to different computing environments, ranging from high-end data center server clusters to commodity PCs and laptops. The ledger also supports Turing-Complete smart contracts, which enables rich user experiences for DApps built on top of the Pando Ledger. For more details, please refer to our [whitepaper](https://pandoproject.org/wp-content/themes/pando-project/img/whitepaper.pdf) and [2019 IEEE ICBC paper](https://arxiv.org/pdf/1911.04698.pdf) "Scalable BFT Consensus Mechanism Through Aggregated

Signature Gossip".

  

To learn more about the Pando Network in general, please visit the **Pando Documentation site**: https://docs.pandoprojects.org/docs/what-is-pando-network.

  

## Table of Contents

- Rametrons Description

- [Setup](#setup)

- [Smart Contract and DApp Development on Pando](#smart-contract-and-dapp-development-on-pando)

  

## Rametrons Description


Rametron nodes are instances of the cluster between the user and cluster's machines. Users can run their jobs on the rametron node instead of doing it directly on the master nodes, which are critical for the overall functioning. This way you can prevent capacity losses on these nodes.
  

### Rametron Enterprise 

Rametron Enterprise can have a stake of more than 35000 ptx only for a locking period of a year. Till the period a reward of 15% would be maintained. After the locking period the whole amount could be withdrawn at once by simply withdrawing stake .

### Rametron Pro

Rametron Pro can have a stake of more than 10000 ptx only for a locking period of a year. Till the period a reward of 12% would be maintained. After the locking period the whole amount could be withdrawn at once by simply withdrawing the stake.

### Rametron Lite

Rametron Lite can have a stake of more than 1000 ptx only for a locking period of a year. Till the period a reward of 10% would be maintained. After the locking period the whole amount could be withdrawn at once by simply withdrawing the stake.

### Rametron Mobile 

Rametron Mobile  can have a stake of more than 250 ptx only for a locking period of a year. Till the period a reward of 10% would be maintained. After the locking period the whole amount could be withdrawn at once by simply withdrawing the stake.




  

## Setup

  

### Intall Go

  

Install Go and set environment variables `GOPATH` , `GOBIN`, and `PATH`. The current code base should compile with **Go 1.14.2**. On macOS, install Go with the following command

  

```

brew install go@1.14.1

brew link go@1.14.1 --force

```

  

### Build and Install

  

Next, clone this repo into your `$GOPATH`. The path should look like this: `$GOPATH/src/github.com/pandoprojects/pando`

  

```

git clone https://github.com/pandoprojects/pando-protocol-ledger.git $GOPATH/src/github.com/pandoprojects/pando

export Pando_HOME=$GOPATH/src/github.com/pandoprojects/pando

cd $Pando_HOME

```

  

Now, execute the following commands to build the Pando binaries under `$GOPATH/bin`. Two binaries `pando` and `pandocli` are generated. `pando` can be regarded as the launcher of the Pando Ledger node, and `pandocli` is a wallet with command line tools to interact with the ledger.

  

```

export GO111MODULE=on

make install

```

  

#### Notes for Linux binary compilation

The build and install process on **Linux** is similar, but note that Ubuntu 18.04.4 LTS / Centos 8 or higher version is required for the compilation.

  

#### Notes for Windows binary compilation

The Windows binary can be cross-compiled from macOS. To cross-compile a **Windows** binary, first make sure `mingw64` is installed (`brew install mingw-w64`) on your macOS. Then you can cross-compile the Windows binary with the following command:

  

```

make exe

```

  

You'll also need to place three `.dll` files `libgcc_s_seh-1.dll`, `libstdc++-6.dll`, `libwinpthread-1.dll` under the same folder as `pando.exe` and `pandocli.exe`.

  
  

### Run Unit Tests

Run unit tests with the command below

```

make test_unit

```

  

## Smart Contract and DApp Development on Pando

  

Pando provides full support for Turing-Complete smart contract, and is EVM compatible. To start developing on the Pando Blockchain, please check out the following links:

  

### Smart Contracts

* Smart contract and DApp development Overview: [link here](https://docs.pandoproject.org/pandoproject/smart-contracts).

* Tutorials on how to interact with the Pando blockchain through [Metamask](https://docs.pandoproject.org/pandoproject/connect-to-metamask).

* PNC20 Token (i.e. ERC20 on Pando) integration guide: [link here](https://docs.pandoproject.org/pandoproject/smart-contracts/pnc-20).

  

### Local Test Environment Setup

* Launching a local privatenet: [link here](https://docs.pandoproject.org/pandoproject/blockchain-integration/launch-a-local-private-net).

* Command line tools: [link here](https://docs.pandoproject.org/pandoproject/blockchain-integration/command-line-tool).

* Connect to the [Testnet](https://docs.pandoproject.org/pandoproject/blockchain-integration/connect-to-the-testnet).

* Node configuration: [link here](https://docs.pandoproject.org/pandoproject/blockchain-integration/configuration).

  

### API References

* Native RPC API references: [link here](https://chainapi.pandoproject.org/#e3785136-4a50-4472-9226-7ac827a0fbf4).

* Ethereum RPC API support: [link here](https://docs.pandoproject.org/pandoproject/smart-contracts/ethereum-rpc-api-support).
