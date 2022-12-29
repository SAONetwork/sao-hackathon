package server

// TODO: if there are much more apis, split into different files.

import (
	"fmt"
	"path/filepath"
	"sao-datastore-storage/cmd"
	"sao-datastore-storage/common"
	"sao-datastore-storage/model"
	"sao-datastore-storage/store"
	"sao-datastore-storage/util"

	logging "github.com/ipfs/go-log/v2"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var log = logging.Logger("server")

type Server struct {
	StoreService store.StoreService
	Model        *model.Model
	Config       common.ApiServerInfo
	Repodir      string
}

func (s *Server) ServeAPI(listen string, contextPath string, swagHandler gin.HandlerFunc) {
	r := gin.Default()
	r.Use(cors.New(s.CorsConfig()))

	// hackathon
	hackathon := r.Group(contextPath+"/api/v1", util.VerifySignature)
	{
		hackathon.POST("/file/upload", s.UploadFile)
		hackathon.POST("/file/addFileWithPreview", s.AddFileWithPreview)
		hackathon.DELETE("/file/upload/:previewId", s.DeleteUploaded)
		hackathon.GET("/file/order/download/:fileId", s.Download)
		hackathon.POST("/fileStar", s.StarFile)
		hackathon.DELETE("/fileStar", s.DeleteStarFile)

		hackathon.POST("/user", s.UpdateUserProfile)
		hackathon.GET("/user", s.GetUserProfile)
		hackathon.GET("/user/purchases", s.GetUserPurchases)
		hackathon.GET("/user/dashboard", s.GetUserDashboard)
		hackathon.GET("/user/summary", s.GetUserSummary)

		hackathon.GET("/collection", s.GetCollection)
		hackathon.GET("/collection/recommendedTags", s.GetRecommendedTags)
		hackathon.POST("/collection", s.UpsertCollection)
		hackathon.DELETE("/collection/:collectionId", s.DeleteCollection)
		hackathon.POST("/collectionFile", s.AddFileToCollection)
		hackathon.DELETE("/collectionFile", s.RemoveFileFromCollection)
		hackathon.POST("/collectionLike", s.LikeCollection)
		hackathon.DELETE("/collectionLike", s.UnLikeCollection)
		hackathon.POST("/collectionStar", s.StarCollection)
		hackathon.DELETE("/collectionStar", s.DeleteStarCollection)
	}

	noSignature := r.Group(contextPath + "/api/v1")
	{
		noSignature.GET("/fileInfos", s.FileInfos)
		noSignature.GET("/file/:fileId", s.FileInfo)
		noSignature.GET("/search", s.GeneralSearch)
	}

	fmt.Println(s.Config.PreviewsPath)
	r.Static(contextPath + "/previews", s.Config.PreviewsPath)
	procDir := filepath.Join(s.Repodir, cmd.FsStaging, "proc")
	r.Static(contextPath + "/api/v1/proc/file", procDir)

	// swagger
	r.GET(contextPath+ "/swagger/*any", swagHandler)

	r.Run(listen)
}

func (s *Server) CorsConfig() cors.Config {
	corsConf := cors.DefaultConfig()
	corsConf.AllowAllOrigins = true
	corsConf.AllowMethods = []string{"GET", "POST", "DELETE", "OPTIONS", "PUT"}
	corsConf.AllowHeaders = []string{
		"Authorization", "Content-Type", "Upgrade", "Origin",
		"Connection", "Accept-Encoding", "Accept-Language", "Host", "Access-Control-Request-Method", "Access-Control-Request-Headers",
		"signature", "signaturemessage", "address", "contenttype",
	}
	return corsConf
}
