package server

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"sao-datastore-storage/model"
	"sao-datastore-storage/util"
	"sao-datastore-storage/util/api"
	"strconv"
)

func (s *Server) AddCollectionComment(ctx *gin.Context) {
	ethAddress := ctx.GetHeader("address")
	if ethAddress == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	var comment model.CollectionComment
	decoder := json.NewDecoder(ctx.Request.Body)
	err := decoder.Decode(&comment)
	if err != nil {
		api.BadRequest(ctx, "invalid.param", err.Error())
		return
	}
	if comment.Comment == "" {
		api.BadRequest(ctx, "invalid.param", "comment must not be empty")
		return
	}
	comment.EthAddr = ethAddress

	result, err := s.Model.AddCollectionComment(&comment)
	if err != nil {
		api.ServerError(ctx, "createCollection.error", err.Error())
		return
	}
	api.Success(ctx, result)
}

func (s *Server) DeleteCollectionComment(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}
	commentIdParam := ctx.Param("commentId")
	commentId, err := strconv.ParseUint(commentIdParam, 10, 0)
	if err != nil {
		api.BadRequest(ctx, "invalid.param.commentId", err.Error())
		return
	}
	err = s.Model.DeleteCollectionComment(uint(commentId))
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "deleteCollectionComment.error", err.Error())
		return
	}
	api.Success(ctx, true)
}

func (s *Server) GetCollectionComments(ctx *gin.Context) {
	ethAddress := ctx.GetHeader("address")
	util.VerifySignature(ctx)
	user, _ := ctx.Get("User")
	if ethAddress != "" && user.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	collectionIdParam, got := ctx.GetQuery("collectionId")
	if !got {
		collectionIdParam = "0"
	}
	collectionId, err := strconv.ParseUint(collectionIdParam, 10, 0)
	if err != nil {
		collectionId = 0
	}

	comments, err := s.Model.GetCollectionComment(uint(collectionId), ethAddress)
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "getCollectionComments.error", err.Error())
		return
	}
	api.Success(ctx, comments)
}

func (s *Server) LikeCollectionComment(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	commentIdParam, got := ctx.GetQuery("commentId")
	if !got {
		commentIdParam = "0"
	}
	commentId, err := strconv.ParseUint(commentIdParam, 10, 0)
	if err != nil {
		commentId = 0
	}

	err = s.Model.LikeCollectionComment(ethAddress.(string), uint(commentId))
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "likeCollectionComment.error", err.Error())
		return
	}
	api.Success(ctx, true)
}

func (s *Server) UnLikeCollectionComment(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	commentIdParam, got := ctx.GetQuery("commentId")
	if !got {
		commentIdParam = "0"
	}
	commentId, err := strconv.ParseUint(commentIdParam, 10, 0)
	if err != nil {
		commentId = 0
	}

	err = s.Model.UnlikeCollectionComment(ethAddress.(string), uint(commentId))
	if err != nil {
		log.Error(err)
	}
	api.Success(ctx, true)
}