package main

import (
	"errors"
	"fmt"
	"os"
	"sao-datastore-storage/build"
	"sao-datastore-storage/cmd"
	"sao-datastore-storage/model"
	"sao-datastore-storage/monitor"

	"github.com/ethereum/go-ethereum/log"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
)

var runCmd = &cli.Command{
	Name: "run",
	Action: func(cctx *cli.Context) error {
		// repo dir
		cfgdir, err := homedir.Expand(cctx.String(cmd.FlagDsRepo.Name))
		if err != nil {
			return err
		}

		_, err = os.Stat(cfgdir)
		if err != nil && errors.Is(err, os.ErrNotExist) {
			return errors.New("repo dir doesn't exist.")
		}

		// config
		config, err := cmd.GetConfig(cfgdir)
		if err != nil {
			return err
		}

		// mysql
		log.Info("connect db...")
		connString := cmd.GetDBConnString(config.Mysql)

		model, err := model.NewModel(connString, true, config)
		if err != nil {
			fmt.Println(err)
		}

		m, err := monitor.NewMonitor(config.Monitor, model)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		m.Run()

		return nil
	},
}

func main() {
	app := &cli.App{
		Name:                 "sao-monitor",
		Usage:                "SAO Monitor",
		EnableBashCompletion: true,
		Version:              build.UserVersion(),
		Flags: []cli.Flag{
			cmd.FlagDsRepo,
		},
		Commands: []*cli.Command{
			runCmd,
		},
	}
	app.Setup()
	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString("Error: " + err.Error() + "\n")
	}
}
