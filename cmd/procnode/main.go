package main

import (
	lcli "github.com/filecoin-project/lotus/cli"
	cliutil "github.com/filecoin-project/lotus/cli/util"
	logging "github.com/ipfs/go-log/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"sao-datastore-storage/build"
	"sao-datastore-storage/cmd"
	"sao-datastore-storage/model"
	"sao-datastore-storage/node"
	"sao-datastore-storage/proc"
)

var log = logging.Logger("proc")

var FlagRepo = &cli.StringFlag{
	Name:    "repo",
	Usage:   "repo directory for proc node",
	EnvVars: []string{"SAO_PROCNODE_PATH"},
	Value:   cmd.FsDefaultProcNodeRepo,
}

func before(cctx *cli.Context) error {
	_ = logging.SetLogLevel("proc", "INFO")
	_ = logging.SetLogLevel("node", "INFO")

	if cliutil.IsVeryVerbose {
		_ = logging.SetLogLevel("proc", "DEBUG")
		_ = logging.SetLogLevel("node", "DEBUG")
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:                 "sao-procnode",
		Usage:                "File Processing Node for SAO",
		EnableBashCompletion: true,
		Version:              build.UserVersion(),
		Before:               before,
		Flags: []cli.Flag{
			FlagRepo,
			cliutil.FlagVeryVerbose,
		},
		Commands: []*cli.Command{
			initCmd,
			runCmd,
		},
	}
	app.Setup()

	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString("Error: " + err.Error() + "\n")
		os.Exit(1)
	}
}

var initCmd = &cli.Command{
	Name: "init",
	Action: func(cctx *cli.Context) error {
		log.Info("initializing proc node...")

		log.Info("load configuration...")
		cfgdir, err := homedir.Expand(cctx.String(FlagRepo.Name))
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

		log.Infof("initializing database...")
		connString := cmd.GetDBConnString(config.Mysql)
		if err != nil {
			return err
		}
		logLevel := logger.Silent
		if cliutil.IsVeryVerbose {
			logLevel = logger.Info
		}
		db, err := gorm.Open(mysql.Open(connString), &gorm.Config{Logger: logger.Default.LogMode(logLevel)})

		if err = db.AutoMigrate(&model.KeyStore{}); err != nil {
			return err
		}

		log.Info("initialize proc node succeed.")
		return nil
	},
}

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "Start a procnode process",
	Flags: []cli.Flag{},
	Action: func(cctx *cli.Context) error {
		ctx := lcli.ReqContext(cctx)

		shutdownChan := make(chan struct{})

		log.Info("starting process node server...")

		log.Info("checking repo dir...")
		cfgdir, err := homedir.Expand(cctx.String(FlagRepo.Name))
		if err != nil {
			return err
		}

		_, err = os.Stat(cfgdir)
		if err != nil && errors.Is(err, os.ErrNotExist) {
			return errors.New("repo dir doesn't exist.")
		}

		log.Info("load configuration...")
		config, err := cmd.GetConfig(cfgdir)
		if err != nil {
			return err
		}

		log.Info("setup repo ", FlagRepo.Value)
		n, err := node.Setup(ctx, cfgdir, config.Libp2p)
		if err != nil {
			return err
		}
		log.Infof("node peer id: %v, multiaddrs: %v", n.Host.ID(), n.Host.Addrs())

		log.Info("Connecting db...")
		connString := cmd.GetDBConnString(config.Mysql)

		m, err := model.NewModel(connString, cliutil.IsVeryVerbose, config)
		if err != nil {
			return err
		}

		procNode := proc.NewProcNode(ctx, n.Host, n.Wallet, m, config, cfgdir)
		procNode.Start()

		log.Info("process node server is started.")

		finishCh := node.MonitorShutdown(
			shutdownChan,
			node.ShutdownHandler{Component: "procnode", StopFunc: procNode.Stop},
		)
		<-finishCh

		return nil
	},
}
