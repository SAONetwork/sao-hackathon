package proc

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"path/filepath"
	"sao-datastore-storage/cmd"
)

func (p *ProcNode) serverAPI() {
	r := gin.Default()
	// TODO authorization
	procDir := filepath.Join(p.repodir, cmd.FsStaging, "proc")
	r.Static(p.config.ApiServer.ContextPath + "/api/v1/proc/encrypt", procDir)
	r.Static(p.config.ApiServer.ContextPath + "/api/v1/proc/decrypt", procDir)

	listen := fmt.Sprintf("%s:%d", p.config.ApiServer.Ip, p.config.ApiServer.Port)
	go func() {
		log.Info("listening ", listen)
		r.Run(listen)
	}()
}
