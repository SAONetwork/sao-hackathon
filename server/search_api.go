package server

import (
	"github.com/gin-gonic/gin"
	"sao-datastore-storage/util"
	"sao-datastore-storage/util/api"
)

func (s *Server) GeneralSearch(ctx *gin.Context) {
	ethAddress := ctx.GetHeader("address")
	util.VerifySignature(ctx)
	owner, _ := ctx.Get("User")
	if ethAddress != "" && owner.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}
	key := ctx.Param("key")

	searchScope,_ := ctx.GetQuery("scope")

	switch searchScope {
	case "collection":
		fi,err := s.Model.GetSearchCollectionResult(key)
		if err != nil {
			log.Error(err)
			api.ServerError(ctx, "getSearchCollectionResult.error", err.Error())
		}
		api.Success(ctx, fi)
		return
	case "user":
		// todo
		return
	default:
		fi := s.Model.GetSearchFileResult(key, ethAddress)
		api.Success(ctx, fi)
		return
	}
}