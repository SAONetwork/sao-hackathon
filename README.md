# Storverse Hackathon

[![Made by SaoNetwork](https://img.shields.io/badge/made%20by-SaoNetwork-green.svg)](https://sao.network/)
[![Chat on discord](https://img.shields.io/badge/join%20-discord-brightgreen.svg)](https://discord.com/invite/q58XsnQqQF)

# Introduction

SAO Network is to provide a decentralized storage network that is composed by worldwide sao nodes which offers moduralized and extensible services on top of IPFS/Filecoin. Services includes:
* data cache service
* private data storage
* copyright management
* verifiable claim
* etc

Storverse Hackathon implements a simple sao nodes with private data storage support and a file market demo application to demonstrate how users can upload and purchase file use rights. 

In this hackathon project, we choose to use Filswan MCS storage solution to store files on IPFS/Filecoin, because Filswan MCS provides a complete and convenient solution for end users and application to pay and store on IPFS/Filecoin. Thanks to Filswan's great project.
Since there is no go MCS SDK, we are developing [go-mcs-sdk](https://github.com/SaoNetwork/sao-hackathon/tree/main/go-mcs-sdk) by ourselves, we will open source and contribute it to community once it's complete.

# Getting Started
In Storverse project, we have three parts of service to accomplish the work
- server
  - the main server for the demonstration website, which provide api service for end users and interact with procnode
- monitor
  - the project to listen the contract event in ethereum
- procnode
  - the node which provide data processing ability, like data encryption and decryption, as well as all possible data processing way like version tracking, provenance... to be expanded

### Prerequisites
The required prerequisites that need to be set up before the workshop.

- Install [Go](https://golang.org/doc/install)
    - Minimum version: 1.17
- IPFS node, or a wallet account with enough MATIC and USDC balance in Mumbai Testnet to use [MCS](https://docs.filswan.com/multi-chain-storage/overview)
- Mysql
- Ethereum client provider

### Build
Init submodule
```shell
git submodule update --init --recursive
cd extern/filecoin-ffi
git checkout 943e335
```
Build the project
```shell
make all
```

### Config
#### server
Create server repo, the default repo path is ~/.sao-ds, you can change it by setting environment var SAO_DS_PATH or parameter --repo
```shell
mkdir ~/.sao-ds
touch ~/.sao-ds/config.toml
```
config.toml format:
```shell
[ipfs]
ip = "localhost"
port = 5001

[mcs]
enabled = true
mcsEndpoint = "https://mcs-api.filswan.com/api/v1"
storageEndpoint = "https://api.filswan.com"
enableFilecoin = false
providerRpc = "https://rpc-mumbai.maticvigil.com"
privateKey = ""

[mysql]
user = ""
password = ""
ip = "127.0.0.1"
port = 3306
dbname = "saoserver"

[apiServer]
ip = "127.0.0.1"
port = 8097
contextPath = "/saods"
exposedPath = "http://127.0.0.1:8097"
previewsPath = "my/previews/path"
host = "https://rinkeby.sao.network/saods"

[libp2p]
directPeers = ["/ip4/127.0.0.1/tcp/[port_number]/p2p/[peer_id]"]
```

###### ipfs
ipfs section defines the ipfs node used to upload and download files

###### mcs
mcs section defines the basic information to use FilSwan multi-chain storage
- **enabled:** set true to use MCS to store files, false to use ipfs node defined in ipfs section
- **mcsEndpoint:** the mcs end point
- **storageEndpoint:**  the mcs storage end point
- **enableFilecoin:**  set true to store files in filecoin, it charges MATIC and USDC so your must prepare enough fund to pay
- **providerRpc:**  Mumbai testnet RPC URL

###### mysql
mysql section defines mysql info

###### apiServer
apiServer section is used to provide api service.ip, 
- **port, contextPath:** defined to construct api server 
- **exposedPath:** to interact with procnode, for example procnode use the exposedPath to transfer the original file section and encrypted file section
- **previewsPath:** specify the folder to store the preview of uploaded files 
- **host:** the internet address of our service

###### libp2p
directPeers is defined in this section, the peer id and address can be found in logs when you start your procnode service
```text
2022-08-03T16:35:18.382+0800    INFO    proc    procnode/main.go:141    node peer id: 12D3KooWBhUiC13vCsh4ByWAkVpGnvBrfhZJfu98yM87UF9cpSyb, multiaddrs: [/ip4/127.0.0.1/tcp/36951]
```

#### monitor
The default repo path is ~/.sao-ds and can be custom by environment var SAO_DS_PATH or parameter --repo

config.toml format:
```shell
[monitor]
provider = "wss://rinkeby.infura.io/ws/v3/[project_id]"
contract = "[contract_address]"
blockNumber = [contract_creation_block_number]
mnemonic = ""
```
###### monitor
monitor section is used to listen ethereum event. In this case we deploy contract https://github.com/SaoNetwork/hackathon-contracts/blob/main/contracts/NFT.sol at 0xFA5D30eAC8c9831eCe8b082F2A353Ba86Ee59cb8, from block number 11027543, mnemonic should be filled in config for download event

#### procnode
Create sao-procnode repo, the default repo path is ~/.sao-procnode, you can change it by setting environment var SAO_PROCNODE_PATH or parameter --repo
```shell
mkdir ~/.sao-procnode
touch ~/.sao-procnode/config.toml
```
config.toml format:
```shell
[mysql]
user = ""
password = ""
ip = "127.0.0.1"
port = 3306
dbname = "saonode"

[transport]
maxTransferDuration = 60

[apiServer]
ip = "127.0.0.1"
port = 8098
exposedPath = "http://127.0.0.1:8098"

[libp2p]
listenAddresses = ["/ip4/127.0.0.1/tcp/[port_number]"]
```
###### mysql
mysql section defines mysql info

###### transport
transport section defines the attributes of file transport, for example maxTransferDuration defines the time limit of file transport

###### apiServer
we use http to transfer file sections between server and procnode, so the api server info should also be included in config

###### libp2p
listenAddresses: the p2p address of ds server

### Run

Initialize database schema
```shell
./sao-ds init
./sao-procnode init
```

server
```shell
./sao-ds [--repo=my/server/path] [--vv] run
```

proc
```shell
./sao-procnode [--repo=my/proc/path] [--vv] run
```

monitor
```shell
./sao-monitor [--repo=my/proc/path] run
```

# Tech Design

### Encryption/Decryption Mechanism
Unlike encryption and decryption in client side, we are designing a mechanism to do it in server side. This may be more practical in some cases:
* user may not directly own the file.
* user may not have suitable machine to process big files.

Requirement
* Any single SAO node never has complete file or have any way to recover complete file from other nodes.
* File split detail is unpredictable.
* Original file can be reassembled back.

We are designing draft protocol - /sao/file/encrypt/0.0.1  
Client Node and Proc Node are connected by p2p.  
Encryption Flow:
1. Client Node splits file into chunks, each trunk size is random and around 16M.
2. Client Node sends file encryption request to a random Proc Node by Beacon.
3. Proc Node Retrieves file chunk from request's Transfer.
4. Proc Node generates random file key and encrypt.
5. Proc Node encrypts encrypted chunk info and file key and store in decentralized storage.
6. Proc Node shares file chunk info and file key with file owner.
7. Proc Node respond Client Node with encrypted file info. 
8. Client Node receives all encrypted file chunk, reassemble and store on IPFS/Filecoin.

Decryption Flow:
1. Client Node retrieves encrypted complete file from IPFS/Filecoin.
1. Client Node broadcasts file id
2. Proc Node who has chunk info of this file replies with chunk info.
3. Client Node split out file range according to encrypted chunk info.
4. Client Node sends chunk decryption request to Proc Node.
5. Proc Node receives the chunk from request's Transfer and decrypts encrypted chunk using file key.
6. Proc Node responds Client Node with decrypted chunk info.
6. Client Node receives decrypted chunk and reassemble the original file.

This video also explains the flow - [Encrypt/Decrypt Flow](https://www.youtube.com)

```golang
// /sao/file/encrypt/0.0.1
type FileEncryptReq struct {
  FileId string
  ClientId string
  Offset uint64
  Size uint64
  Transfer types.Transfer
}
type FileEncryptResp struct {
  FileKey  string
  Transfer types.Transfer
  Accepted bool
}
type FileDecryptReq struct {
  FileId string
  ClientId string
  Offset uint64
  Size uint64
  Transfer types.Transfer
}
type FileDecryptResp struct {
  FileId   string
  Offset   uint64
  Size     uint64
  Transfer types.Transfer
  Accepted bool
}
```

The protocol will soon iterate into newer version with more security and efficiency consideration.
