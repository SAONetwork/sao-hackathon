package common

import (
	"time"

	"github.com/BurntSushi/toml"
)

type MysqlInfo struct {
	User     string
	Password string
	Ip       string
	Port     int
	Dbname   string
}

type Transport struct {
	MaxTransferDuration time.Duration
}

type ApiServerInfo struct {
	Ip           string
	Port         int
	Host         string
	ContextPath  string
	ExposedPath  string
	PreviewsPath string
}

type Libp2p struct {
	ListenAddresses []string
	DirectPeers     []string
}

type MonitorInfo struct {
	Provider    string
	Contract    string
	BlockNumber int64
	Mnemonic    string
}

type Config struct {
	ApiServer    ApiServerInfo
	Ipfs         IpfsInfo
	Mysql        MysqlInfo
	Monitor      MonitorInfo
	Libp2p       Libp2p
	PreviewsPath string
	Transport    Transport
}

type IpfsInfo struct {
	Ip            string
	Port          int
	ProjectId     string
	ProjectSecret string
}

func GetConfig(cfgPath string) (*Config, error) {
	var cfg Config
	_, err := toml.DecodeFile(cfgPath, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
