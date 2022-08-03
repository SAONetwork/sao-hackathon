# Storverse


# Getting Started

### Prerequisites
The required prerequisites that need to be set up before the workshop.

- Install [Go](https://golang.org/doc/install)
    - Minimum version: 1.16
- IPFS node
- Mysql
- Ethereum client provider

### Build
introduce Filecoin FFI first
```shell
cd ./extern/filecoin-ffi
git clone git@github.com:filecoin-project/filecoin-ffi.git ./
git checkout 943e335
```
at the first time, uncomment these lines in Makefile to make sure Filecoin FFI has been built
```shell
#FFI_PATH:=extern/filecoin-ffi/
#FFI_DEPS:=.install-filcrypto
#FFI_DEPS:=$(addprefix $(FFI_PATH),$(FFI_DEPS))

#build/.filecoin-install: $(FFI_PATH)
	#$(MAKE) -C $(FFI_PATH) $(FFI_DEPS:$(FFI_PATH)%=%)
#BUILD_DEPS+=build/.filecoin-install
```
build the project
```shell
make all
```

### Config
#### server
create server repo, the default repo path is ~/.sao-ds, you can change it by setting environment var SAO_DS_PATH or parameter --repo
```shell
mkdir ~/.sao-ds
touch ~/.sao-ds/config.toml
```
config.toml format:
```shell
[ipfs]
ip = "localhost"
port = 5001

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
previewsPath = "/home/ubuntu/go-sao-datastore/previews"
host = "https://rinkeby.sao.network/saods"

[libp2p]
directPeers = ["/ip4/127.0.0.1/tcp/[port_number]/p2p/xxx...xxx"]
```

###### ipfs
ipfs section defines the ipfs node used to upload and download files.

###### mysql
mysql section defines mysql info

###### apiServer
apiServer section is used to provide api service.ip, port, contextPath used to construct api server, exposedPath is used to interact with procnode, for example, procnode use the exposedPath to transfer the original file section and encrypted file section.
previewsPath specify the folder to store the preview of uploaded files. host is the internet address of our service.

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
provider = "wss://rinkeby.infura.io/ws/v3/xxx...xxx"
contract = "0xXXX...XXX"
blockNumber = 11027543
mnemonic = ""
```
###### monitor
monitor section is used to listen ethereum event. In this case we deploy contract https://github.com/SaoNetwork/hackathon-contracts/blob/main/contracts/NFT.sol at 0xFA5D30eAC8c9831eCe8b082F2A353Ba86Ee59cb8, from block number 11027543, mnemonic should be filled in config for download event.

#### procnode
create sao-procnode repo, the default repo path is ~/.sao-procnode, you can change it by setting environment var SAO_PROCNODE_PATH or parameter --repo
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
we use http to transfer file sections between server and procnode, so the api server info should also be included in config.

###### libp2p
listenAddresses: the p2p address of ds server

### Run

initialize database schema
```shell
./sao-ds init
```

server
```shell
./sao-ds --repo=my/server/path --vv run
```

proc
```shell
./sao-procnode --repo=my/proc/path --vv run
```

monitor
```shell
./sao-monitor run
```
