package server

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"sao-datastore-storage/model"
	"sao-datastore-storage/util"
	"sao-datastore-storage/util/api"
	"strconv"
)

func (s *Server) AddFileComment(ctx *gin.Context) {
	ethAddress := ctx.GetHeader("address")
	if ethAddress == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	var comment model.FileComment
	decoder := json.NewDecoder(ctx.Request.Body)
	err := decoder.Decode(&comment)
	if err != nil {
		api.BadRequest(ctx, "invalid.param", err.Error())
		return
	}
	comment.EthAddr = ethAddress

	result, err := s.Model.AddFileComment(&comment)
	if err != nil {
		api.ServerError(ctx, "createCollection.error", err.Error())
		return
	}
	api.Success(ctx, result)
}

func (s *Server) DeleteFileComment(ctx *gin.Context) {
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
	s.Model.DeleteFileComment(uint(commentId))
	api.Success(ctx, true)
}

func (s *Server) GetFileComments(ctx *gin.Context) {
	ethAddress := ctx.GetHeader("address")
	util.VerifySignature(ctx)
	user, _ := ctx.Get("User")
	if ethAddress != "" && user.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	fileIdParam, got := ctx.GetQuery("fileId")
	if !got {
		fileIdParam = "0"
	}
	fileId, err := strconv.ParseUint(fileIdParam, 10, 0)
	if err != nil {
		fileId = 0
	}

	comments, err := s.Model.GetFileComment(uint(fileId), ethAddress)
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "getFileComments.error", err.Error())
		return
	}
	api.Success(ctx, comments)
}

func (s *Server) LikeFileComment(ctx *gin.Context) {
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

	err = s.Model.LikeFileComment(ethAddress.(string), uint(commentId))
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "likeFileComment.error", err.Error())
		return
	}
	api.Success(ctx, true)
}

func (s *Server) UnLikeFileComment(ctx *gin.Context) {
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

	err = s.Model.UnlikeFileComment(ethAddress.(string), uint(commentId))
	if err != nil {
		log.Error(err)
	}
	api.Success(ctx, true)
}