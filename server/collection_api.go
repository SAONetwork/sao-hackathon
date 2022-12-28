package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"image/jpeg"
	"image/png"
	"sao-datastore-storage/model"
	"sao-datastore-storage/util/api"
	"strconv"
	"strings"
)

func (s *Server) UpsertCollection(ctx *gin.Context) {
	ethAddress := ctx.GetHeader("address")
	if ethAddress == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	var collection model.Collection
	decoder := json.NewDecoder(ctx.Request.Body)
	err := decoder.Decode(&collection)
	if err != nil {
		api.BadRequest(ctx, "invalid.param", err.Error())
		return
	}

	if collection.Preview != "" {
		img, err := png.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(collection.Preview)))
		if err != nil {
			log.Info(err)
			img, err = jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(collection.Preview)))
			if err != nil {
				api.BadRequest(ctx, "invalid.preview", fmt.Sprintf("decode preview failed: %v", "png and jpeg decode failed"))
				return
			}
		}
		id := uuid.New().String()
		dc := gg.NewContextForImage(img)
		preview := fmt.Sprintf("%s/%s.png", s.Config.PreviewsPath, id)
		dc.SavePNG(preview)
		collection.Preview = fmt.Sprintf("%s.png", id)
	}

	err = s.Model.UpsertCollection(&collection)
	if err != nil {
		api.ServerError(ctx, "createCollection.error", err.Error())
		return
	}
	api.Success(ctx, collection)
}

func (s *Server) DeleteCollection(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}
	collectionIdParam := ctx.Param("collectionId")
	collectionId, err := strconv.ParseUint(collectionIdParam, 10, 0)
	if err != nil {
		api.BadRequest(ctx, "invalid.param.collectionId", err.Error())
		return
	}
	s.Model.DeleteCollection(uint(collectionId))
	api.Success(ctx, true)
}

func (s *Server) GetCollection(ctx *gin.Context) {
	collectionIdParam, got := ctx.GetQuery("collectionId")
	if !got {
		collectionIdParam = "0"
	}
	collectionId, err := strconv.ParseUint(collectionIdParam, 10, 0)
	if err != nil {
		collectionId = 0
	}

	fileIdParam, got := ctx.GetQuery("fileId")
	if !got {
		fileIdParam = "0"
	}
	fileId, err := strconv.ParseUint(fileIdParam, 10, 0)
	if err != nil {
		fileId = 0
	}

	owner, _ := ctx.GetQuery("owner")

	collections, err := s.Model.GetCollection(uint(collectionId), owner, uint(fileId))
	if err != nil {
		log.Error(err)
	}

	var result []model.CollectionVO
	for _, c := range *collections {
		result = append(result, model.CollectionVO{
			Id:          c.Id,
			CreatedAt:   c.CreatedAt.UnixMilli(),
			UpdatedAt:   c.UpdatedAt.UnixMilli(),
			Preview:     fmt.Sprintf("%s/%s/%s", s.Config.Host, "previews", c.Preview),
			Title:       c.Title,
			Labels:      c.Labels,
			Description: c.Description,
			Type:        c.Type,
		})
	}
	api.Success(ctx, result)
}

func (s *Server) AddFileToCollection(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	var collectionFile model.CollectionRequest
	decoder := json.NewDecoder(ctx.Request.Body)
	err := decoder.Decode(&collectionFile)
	if err != nil {
		api.BadRequest(ctx, "invalid.param", err.Error())
		return
	}

	for _, collectionId := range collectionFile.CollectionIds {
		err = s.Model.AddFileToCollection(collectionFile.FileId, collectionId, ethAddress.(string))
		if err != nil {
			log.Error(err)
			api.ServerError(ctx, "addFileToCollection.error", err.Error())
		}
	}
	api.Success(ctx, true)
}

func (s *Server) RemoveFileFromCollection(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
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

	fileIdParam, got := ctx.GetQuery("fileId")
	if !got {
		fileIdParam = "0"
	}
	fileId, err := strconv.ParseUint(fileIdParam, 10, 0)
	if err != nil {
		fileId = 0
	}

	err = s.Model.RemoveFileFromCollection(ethAddress.(string), uint(fileId), uint(collectionId))
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "removeFileFromCollection.error", err.Error())
	}
	api.Success(ctx, true)
}

func (s *Server) LikeCollection(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
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

	err = s.Model.LikeCollection(ethAddress.(string), uint(collectionId))
	if err != nil {
		log.Error(err)
	}
	api.Success(ctx, true)
}

func (s *Server) UnLikeCollection(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
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

	err = s.Model.UnlikeCollection(ethAddress.(string), uint(collectionId))
	if err != nil {
		log.Error(err)
	}
	api.Success(ctx, true)
}

func (s *Server) StarCollection(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
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

	err = s.Model.StarCollection(ethAddress.(string), uint(collectionId))
	if err != nil {
		log.Error(err)
	}
	api.Success(ctx, true)
}

func (s *Server) DeleteStarCollection(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
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

	err = s.Model.DeleteStarCollection(ethAddress.(string), uint(collectionId))
	if err != nil {
		log.Error(err)
	}
	api.Success(ctx, true)
}
