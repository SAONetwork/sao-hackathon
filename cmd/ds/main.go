package main

import (
	"fmt"
	lcli "github.com/filecoin-project/lotus/cli"
	cliutil "github.com/filecoin-project/lotus/cli/util"
	logging "github.com/ipfs/go-log/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/urfave/cli/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"sao-datastore-storage/build"
	"sao-datastore-storage/cmd"
	"sao-datastore-storage/docs"
	"sao-datastore-storage/model"
	"sao-datastore-storage/node"
	saoserver "sao-datastore-storage/server"
	"sao-datastore-storage/store"
)

var log = logging.Logger("ds")

func before(cctx *cli.Context) error {
	_ = logging.SetLogLevel("ds", "info")
	_ = logging.SetLogLevel("store", "info")
	_ = logging.SetLogLevel("proc", "info")
	_ = logging.SetLogLevel("node", "info")
	_ = logging.SetLogLevel("file", "info")

	if cliutil.IsVeryVerbose {
		_ = logging.SetLogLevel("ds", "debug")
		_ = logging.SetLogLevel("store", "debug")
		_ = logging.SetLogLevel("proc", "debug")
		_ = logging.SetLogLevel("node", "debug")
		_ = logging.SetLogLevel("file", "debug")
	}
	return nil
}

func main() {
	app := &cli.App{
		Name:                 "sao-ds",
		Usage:                "SAO data store service",
		EnableBashCompletion: true,
		Version:              build.UserVersion(),
		Before:               before,
		Flags: []cli.Flag{
			cmd.FlagDsRepo,
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
	}

}

var initCmd = &cli.Command{
	Name: "init",
	Action: func(cctx *cli.Context) error {
		log.Info("initializing saods...")

		log.Info("load configuration...")
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

		if err = db.AutoMigrate(&model.FileInfo{}); err != nil {
			return err
		}
		if err = db.AutoMigrate(&model.FilePreview{}); err != nil {
			return err
		}
		// ds.Model.Migrate(model.XXX)
		if err = db.AutoMigrate(&model.PurchaseOrder{}); err != nil {
			return err
		}
		if err = db.AutoMigrate(&model.UserProfile{}); err != nil {
			return err
		}
		if err = db.AutoMigrate(&model.UserFollowing{}); err != nil {
			return err
		}

		if err = db.AutoMigrate(&model.FileChunkMetadata{}); err != nil {
			return err
		}
		if err = db.AutoMigrate(&model.McsInfo{}); err != nil {
			return err
		}
		if err = db.AutoMigrate(&model.Collection{}); err != nil {
			return err
		}
		if err = db.AutoMigrate(&model.CollectionFile{}); err != nil {
			return err
		}
		if err = db.AutoMigrate(&model.CollectionLike{}); err != nil {
			return err
		}
		if err = db.AutoMigrate(&model.CollectionStar{}); err != nil {
			return err
		}
		if err = db.AutoMigrate(&model.FileStar{}); err != nil {
			return err
		}
		if err = db.AutoMigrate(&model.FileComment{}); err != nil {
			return err
		}
		if err = db.AutoMigrate(&model.FileCommentLike{}); err != nil {
			return err
		}

		log.Info("initialize saods succeed.")

		return nil
	},
}

var runCmd = &cli.Command{
	Name: "run",
	Action: func(cctx *cli.Context) error {
		ctx := lcli.ReqContext(cctx)

		shutdownChan := make(chan struct{})

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

		log.Info("setup repo ", cmd.FlagDsRepo.Value)
		n, err := node.Setup(ctx, cfgdir, config.Libp2p)
		if err != nil {
			return err
		}
		log.Infof("node peer id: %v, multiaddrs: %v", n.Host.ID(), n.Host.Addrs())

		// mysql
		log.Info("connect db...")
		connString := cmd.GetDBConnString(config.Mysql)

		m, err := model.NewModel(connString, cliutil.IsVeryVerbose, config)
		if err != nil {
			return err
		}

		storeService, err := store.NewStoreService(config, m, n.Host, cfgdir)
		if err != nil {
			return err
		}

		server := saoserver.Server{
			StoreService: storeService,
			Model:        m,
			Config:       config.ApiServer,
			Repodir:      cfgdir,
		}
		listen := fmt.Sprintf("%s:%d", config.ApiServer.Ip, config.ApiServer.Port)
		log.Info("listening ", listen)
		docs.SwaggerInfo.BasePath = config.ApiServer.ContextPath + "/api/v1"
		docs.SwaggerInfo.Version = "1.0"
		docs.SwaggerInfo.Title = "Storverse API Documentation"

		go func() {
			server.ServeAPI(listen, config.ApiServer.ContextPath, ginSwagger.WrapHandler(swaggerFiles.Handler))
		}()

		finishCh := node.MonitorShutdown(
			shutdownChan,
		)
		<-finishCh

		return nil
	},
}
