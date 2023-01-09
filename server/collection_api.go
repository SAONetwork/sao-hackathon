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
	"io/ioutil"
	"net/http"
	"net/url"
	"sao-datastore-storage/model"
	"sao-datastore-storage/util"
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

func (s *Server) GetLikedCollection(ctx *gin.Context) {
	ethAddress := ctx.GetHeader("address")
	util.VerifySignature(ctx)
	user, _ := ctx.Get("User")
	if ethAddress != "" && user.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	userAddress, got := ctx.GetQuery("address")
	if !got {
		userAddress = ethAddress
	}
	if userAddress == "" {
		api.BadRequest(ctx, "invalid.param", "")
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

	collections, err := s.Model.GetLikedCollection(userAddress, o, l, ethAddress)
	if err != nil {
		log.Error(err)
	}
	api.Success(ctx, collections)
}

func (s *Server) GetCollection(ctx *gin.Context) {
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

	fileIdParam, got := ctx.GetQuery("fileId")
	if !got {
		fileIdParam = "0"
	}
	fileId, err := strconv.ParseUint(fileIdParam, 10, 0)
	if err != nil {
		fileId = 0
	}

	owner, _ := ctx.GetQuery("owner")

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

	collections, err := s.Model.GetCollection(uint(collectionId), owner, uint(fileId), ethAddress, o, l)
	if err != nil {
		log.Error(err)
	}
	api.Success(ctx, collections)
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

	err = s.Model.AddFileToCollections(collectionFile.FileId, collectionFile.CollectionIds, ethAddress.(string))
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "addFileToCollection.error", err.Error())
	}
	var result bool
	if len(collectionFile.CollectionIds) > 0{
		result = true
	} else {
		result = false
	}
	api.Success(ctx, result)
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
		api.ServerError(ctx, "likeCollection.error", err.Error())
		return
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

func (s *Server) GetRecommendedTags(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	desc, got := ctx.GetPostForm("desc")
	if !got {
		return
	}

	textRazorUrl := "https://api.textrazor.com/"
	method := "POST"

	payload := strings.NewReader("extractors=topics&text="+ url.QueryEscape(desc))

	client := &http.Client {
	}
	req, err := http.NewRequest(method, textRazorUrl, payload)

	if err != nil {
		api.ServerError(ctx, "getRecommendedTags.error", err.Error())
		return
	}
	req.Header.Add("x-textrazor-key", "ab968d7fd7770398cd498757947a58d9334377387d7a707a161ce108")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "getRecommendedTags.error", err.Error())
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "getRecommendedTags.error", err.Error())
		return
	}

	var textRazor TextRazorResponse
	err = json.Unmarshal(body, &textRazor)
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "getRecommendedTags.error", err.Error())
		return
	}

	var labels []string
	labelIndex := 0
	for _, coarseTopic := range textRazor.Response.CoarseTopics {
		labels = append(labels, coarseTopic.Label)
		labelIndex++
		if labelIndex >= 2 {
			break
		}
	}

	for _, topic := range textRazor.Response.Topics {
		labels = append(labels, topic.Label)
		labelIndex++
		if labelIndex >= 6 {
			break
		}
	}

	api.Success(ctx, strings.Join(labels, ","))
}