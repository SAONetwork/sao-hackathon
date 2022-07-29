package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"path/filepath"
	"sao-datastore-storage/common"
)

const (
	FsConfig              = "config.toml"
	FsDefaultDsRepo       = "~/.sao-ds"
	FsDefaultProcNodeRepo = "~/.sao-procnode"
	FsStaging             = "staging"
)

func GetDBConnString(mysqlInfo common.MysqlInfo) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlInfo.User,
		mysqlInfo.Password,
		mysqlInfo.Ip,
		mysqlInfo.Port,
		mysqlInfo.Dbname,
	)
}

func GetConfig(cfgdir string) (*common.Config, error) {
	cfgPath := filepath.Join(cfgdir, FsConfig)
	config, err := common.GetConfig(cfgPath)
	if err != nil {
		return nil, err
	}
	return config, nil
}

var FlagDsRepo = &cli.StringFlag{
	Name:    "repo",
	Usage:   "repo directory for saods",
	EnvVars: []string{"SAO_DS_PATH"},
	Value:   FsDefaultDsRepo,
}
