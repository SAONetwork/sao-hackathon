package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"sao-datastore-storage/model"
	"sao-datastore-storage/util"
	"sao-datastore-storage/util/api"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) UploadFile(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		api.BadRequest(ctx, "invalid.param.file", err.Error())
		return
	}

	f, err := file.Open()
	if err != nil {
		api.BadRequest(ctx, "invalid.param.file", err.Error())
		return
	}
	defer f.Close()

	contentType := ctx.ContentType()
	contentType, err = util.DetectReaderType(f)
	if err != nil {
		log.Errorf("detect file type error: %s", err)
	}

	f.Seek(0, 0)

	filename := ctx.PostForm("Filename")
	if filename == "" {
		filename = file.Filename
	}

	additionalInfo := ctx.PostForm("AdditionalInfo")

	log.Infof("%s: staging file", filename)

	fi, err := s.uploadFile(f, filename, contentType, ethAddress.(string), additionalInfo)
	if err != nil {
		api.ServerError(ctx, "uploadfile.error", err.Error())
		return
	}
	api.Success(ctx, fi)
}

func (s *Server) AddFileWithPreview(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	var filePreview model.FilePreview
	decoder := json.NewDecoder(ctx.Request.Body)
	err := decoder.Decode(&filePreview)
	if err != nil {
		api.BadRequest(ctx, "invalid.param", err.Error())
		return
	}

	if filePreview.Id <= 0 {
		api.BadRequest(ctx, "invalid.param", "id must be specified")
		return
	}

	img, err := png.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(filePreview.Preview)))
	if err != nil {
		log.Info(err)
		img, err = jpeg.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(filePreview.Preview)))
		if err != nil {
			api.BadRequest(ctx, "invalid.preview", fmt.Sprintf("decode preview failed: %v", "png and jpeg decode failed"))
			return
		}
	}
	id := uuid.New().String()
	dc := gg.NewContextForImage(img)
	preview := fmt.Sprintf("%s/%s.png", s.Config.PreviewsPath, id)
	dc.SavePNG(preview)
	filePreview.Preview = fmt.Sprintf("%s.png", id)

	fi, err := s.StoreFileWithPreview(ctx.Request.Context(), filePreview, ethAddress.(string))
	if err != nil {
		api.ServerError(ctx, "addFileWithPreview.error", err.Error())
		return
	}
	fi.Preview = fmt.Sprintf("%s/%s/%s", s.Config.Host, "previews", fi.Preview)
	api.Success(ctx, fi)
}

func (s *Server) DeleteUploaded(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	previewId, err := strconv.ParseInt(ctx.Param("previewId"), 10, 64)
	if err != nil {
		api.BadRequest(ctx, "invalid.param", "")
		return
	}
	err = s.deleteUploaded(uint(previewId), ethAddress.(string))
	if err != nil {
		api.ServerError(ctx, "deleteUploaded.error", err.Error())
		return
	}
	api.Success(ctx, nil)
}

func (s *Server) DeleteFile(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	fileId, err := strconv.ParseInt(ctx.Param("fileId"), 10, 64)
	if err != nil {
		api.BadRequest(ctx, "invalid.param", "")
		return
	}
	err = s.deleteFile(ctx, uint(fileId), ethAddress.(string))
	if err != nil {
		api.ServerError(ctx, "deleteFile.error", err.Error())
		return
	}
	api.Success(ctx, nil)
}

func (s *Server) FileInfo(ctx *gin.Context) {
	ethAddress := ctx.GetHeader("address")
	util.VerifySignature(ctx)
	owner, _ := ctx.Get("User")
	if ethAddress != "" && owner.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	fileIdParam := ctx.Param("fileId")
	fileId, err := strconv.ParseUint(fileIdParam, 10, 0)
	if err != nil {
		api.BadRequest(ctx, "invalid.param.fileId", err.Error())
		return
	}

	fi, err := s.getFileInfo(uint(fileId), ethAddress)
	if err != nil {
		api.ServerError(ctx, "getfile.error", err.Error())
		return
	}
	api.Success(ctx, fi)
}

func (s *Server) FileInfosByCollectionId(ctx *gin.Context) {
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

	collectionId,_ := ctx.GetQuery("collectionId")

	fi := s.Model.GetFileInfosByCollectionId(collectionId, ethAddress, o, l)
	api.Success(ctx, fi)
}

func (s *Server) FileInfos(ctx *gin.Context) {
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

	category,_ := ctx.GetQuery("type")

	format,_ := ctx.GetQuery("format")

	pricing, got := ctx.GetQuery("pricing")
	p := -1
	if got {
		if strings.EqualFold(pricing, "true") {
			p = 1
		} else if strings.EqualFold(pricing, "false") {
			p = 0
		}
	}

	fi := s.getFileInfos(ethAddress, o, l, category, format, p)
	api.Success(ctx, fi)
}

// TODO download optimazation
func (s *Server) Download(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	fileIdParam := ctx.Param("fileId")
	fileId, err := strconv.ParseUint(fileIdParam, 10, 0)
	if err != nil {
		api.BadRequest(ctx, "invalid.param.fileId", err.Error())
		return
	}

	err = s.checkFileStatus(uint(fileId), ethAddress.(string))
	if err != nil {
		api.ServerError(ctx, "download.error", err.Error())
		return
	}

	fileInfo, reader, err := s.StoreService.GetFile(ctx, uint(fileId), ethAddress.(string))
	if err != nil {
		api.ServerError(ctx, "getfile.error", err.Error())
		return
	}
	defer reader.Close()

	contentType := "application/octet-stream"
	if fileInfo.ContentType != "" {
		contentType = fileInfo.ContentType
	}
	ctx.Writer.Header().Add("Content-type", contentType)
	ctx.Writer.Header().Add("access-control-expose-headers", "Content-Disposition")
	ctx.Writer.Header().Add("Content-Disposition", "attachment;filename="+fileInfo.Filename)
	_, err = io.Copy(ctx.Writer, reader)
	if err != nil {
		api.ServerError(ctx, "getfile.error", err.Error())
		return
	}
}


func (s *Server) StarFile(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
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

	err = s.Model.StarFile(ethAddress.(string), uint(fileId))
	if err != nil {
		log.Error(err)
	}
	api.Success(ctx, true)
}

func (s *Server) DeleteStarFile(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
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

	err = s.Model.DeleteStarFile(ethAddress.(string), uint(fileId))
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "deleteStarFile.error", err.Error())
		return
	}
	api.Success(ctx, true)
}