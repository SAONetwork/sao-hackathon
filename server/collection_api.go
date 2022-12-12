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

func (s *Server) CreateCollection(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
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

	err = s.Model.CreateCollection(&collection)
	if err != nil {
		api.ServerError(ctx, "createCollection.error", err.Error())
		return
	}
	api.Success(ctx, collection)
}

func (s *Server) EditCollection(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
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

	s.Model.UpsertCollection(&collection)
	api.Success(ctx, true)
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

	collection, err := s.Model.GetCollection(uint(collectionId), owner, uint(fileId))
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "error.get.collection", "database error")
		return
	}
	api.Success(ctx, collection)
}
