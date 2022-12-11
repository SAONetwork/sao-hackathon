package server

import (
	"github.com/gin-gonic/gin"
	"sao-datastore-storage/util/api"
)

func (s *Server) CreateCollection(ctx *gin.Context) {
}


func (s *Server) EditCollection(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}
	api.Success(ctx, true)
}

