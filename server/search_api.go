package server

import (
	"github.com/gin-gonic/gin"
	"sao-datastore-storage/util"
	"sao-datastore-storage/util/api"
	"strconv"
)

func (s *Server) GeneralSearch(ctx *gin.Context) {
	ethAddress := ctx.GetHeader("address")
	util.VerifySignature(ctx)
	owner, _ := ctx.Get("User")
	if ethAddress != "" && owner.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	offset, got := ctx.GetQuery("offset")
	if !got {
		offset = "0"
	}
	o, err := strconv.Atoi(offset)
	if err != nil {
		log.Info(err)
		o = 0
	}
	limit, got := ctx.GetQuery("limit")
	if !got {
		limit = "10"
	}
	l, err := strconv.Atoi(limit)
	if err != nil {
		log.Info(err)
		l = 10
	}

	key,_ := ctx.GetQuery("key")

	searchScope,_ := ctx.GetQuery("scope")

	switch searchScope {
	case "collection":
		collections,err := s.Model.GetSearchCollectionResult(key)
		if err != nil {
			log.Error(err)
			api.ServerError(ctx, "getSearchCollectionResult.error", err.Error())
		}
		api.Success(ctx, collections)
		return
	case "user":
		users,err := s.Model.GetSearchUserResult(key)
		if err != nil {
			log.Error(err)
			api.ServerError(ctx, "getSearchUserResult.error", err.Error())
		}
		api.Success(ctx, users)
		return
	default:
		fi := s.Model.GetSearchFileResult(key, ethAddress, o, l)
		api.Success(ctx, fi)
		return
	}
}