package main

import "github.com/gin-gonic/gin"

type MockCollection struct {
	EthAddr     string
	Preview     string `gorm:"varchar(255);"`
	Labels      string
	Title       string
	Description string
	Type        int
}

type MockCollectionRequest struct {
	CollectionIds []uint
	FileId       uint
	Status       int
}

// @Tags Collection
// @Title GetCollection
// @Description get collection by address
// @Param	collectionId		query 	string	false		"The collection id for query"
// @Param	fileId		query 	string	false		"The file id for query"
// @Param	owner		query 	string	false		"The owner for query"
// @router /collection [get]
func GetCollection(ctx *gin.Context) {
}

// @Tags Collection
// @Title CreateCollection
// @Description create collection
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	body		body 	MockCollection	true		"body for request"
// @router /collection [post]
func CreateCollection(ctx *gin.Context) {
}

// @Tags Collection
// @Title DeleteCollection
// @Description delete collection
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	collectionId		path 	string	true		"The collection id for deletion"
// @router /collection [delete]
func DeleteCollection(ctx *gin.Context) {
}

// @Tags Collection
// @Title AddFileToCollection
// @Description add file to collection
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	body		body 	MockCollectionRequest	true		"body for request"
// @router /collectionFile [post]
func AddFileToCollection(ctx *gin.Context) {
}

// @Tags Collection
// @Title RemoveFileFromCollection
// @Description remove file from collection
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	collectionId		query 	string	false		"The collection id for deletion"
// @Param	fileId		query 	string	false		"The file id for deletion"
// @router /collectionFile [delete]
func RemoveFileFromCollection(ctx *gin.Context) {
}

// @Tags Collection
// @Title LikeCollection
// @Description like collection
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	collectionId		query 	string	false		"The collection id for like operation"
// @router /collectionLike [post]
func LikeCollection(ctx *gin.Context) {
}

// @Tags Collection
// @Title UnLikeCollection
// @Description unlike collection
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	collectionId		query 	string	false		"The collection id for unlike operation"
// @router /collectionLike [delete]
func UnLikeCollection(ctx *gin.Context) {
}