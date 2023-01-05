package main

import "github.com/gin-gonic/gin"

type MockCollection struct {
	Id uint
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

type MockFileComment struct {
	Comment  string
	FileId   uint
	ParentId uint
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
// @Title GetRecommendedTags
// @Description get recommended tags for collection
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	desc		formData 	string	true		"get recommended tags by description"
// @router /collection/recommendedTags [post]
func GetRecommendedTags(ctx *gin.Context) {
}

// @Tags Collection
// @Title GetFileInfosByCollectionId
// @Description get file infos by collection id
// @Param address header string false "user's ethereum address"
// @Param signaturemessage header string false "user's ethereum signaturemessage"
// @Param signature header string false "user's ethereum signature"
// @Param	collectionId		query 	string	true		"The collection id for query"
// @Param	offset		query 	string	false		"offset default 0"
// @Param	limit		query 	string	false		"limit default 10"
// @router /collection/fileInfos [get]
func FileInfosByCollectionId(ctx *gin.Context) {
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
// @router /collection/{collectionId} [delete]
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

// @Tags File
// @Title StarFile
// @Description mark star to a file
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	fileId		query 	string	false		"The file id for star operation"
// @router /fileStar [post]
func StarFile(ctx *gin.Context) {
}

// @Tags File
// @Title DeleteStarFile
// @Description cancel star operation from file
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	fileId		query 	string	false		"The file id for delete star operation"
// @router /fileStar [delete]
func DeleteStarFile(ctx *gin.Context) {
}


// @Tags Search
// @Title GeneralSearch
// @Description search files, collections and users etc.
// @Param address header string false "user's ethereum address"
// @Param signaturemessage header string false "user's ethereum signaturemessage"
// @Param signature header string false "user's ethereum signature"
// @Param	key		query 	string	true		"The key you want to search"
// @Param	scope		query 	string	false		"Set search scope, file/collection/user"
// @router /search [get]
func GeneralSearch(ctx *gin.Context) {
}


// @Tags Comment
// @Title AddFileComment
// @Description add file comment
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	body		body 	MockFileComment	true		"body for request"
// @router /comment/file [post]
func AddFileComment(ctx *gin.Context) {
}

// @Tags Comment
// @Title DeleteFileComment
// @Description delete file comment
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	commentId		path 	string	true		"The comment id for deletion"
// @router /comment/file/{commentId} [delete]
func DeleteFileComment(ctx *gin.Context) {
}

// @Tags Comment
// @Title GetFileComments
// @Description get file comments
// @Param address header string false "user's ethereum address"
// @Param signaturemessage header string false "user's ethereum signaturemessage"
// @Param signature header string false "user's ethereum signature"
// @Param	fileId		query 	string	false		"The file id for query"
// @router /comment/file [get]
func GetFileComments(ctx *gin.Context) {
}

// @Tags Comment
// @Title LikeFileComment
// @Description like file comment
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	commentId		query 	string	true		"The comment id for like operation"
// @router /comment/like [post]
func LikeFileComment(ctx *gin.Context) {
}

// @Tags Comment
// @Title UnLikeFileComment
// @Description unlike file comment
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	commentId		query 	string	true		"The comment id for unlike operation"
// @router /comment/like [delete]
func UnLikeFileComment(ctx *gin.Context) {
}

// @Tags User
// @Title GetUserProfile
// @Description get user followers
// @Param address header string false "user's ethereum address"
// @Param signaturemessage header string false "user's ethereum signaturemessage"
// @Param signature header string false "user's ethereum signature"
// @Param	address		query 	string	false		"user's ethereum address"
// @router /user [get]
func GetUserProfile(ctx *gin.Context) {
}

// @Tags User
// @Title GetUserFollowings
// @Description get user followings
// @Param address header string false "user's ethereum address"
// @Param signaturemessage header string false "user's ethereum signaturemessage"
// @Param signature header string false "user's ethereum signature"
// @Param	address		query 	string	false		"user's ethereum address"
// @router /user/followings [get]
func GetUserFollowings(ctx *gin.Context) {
}

// @Tags User
// @Title GetUserFollowers
// @Description get user followers
// @Param address header string false "user's ethereum address"
// @Param signaturemessage header string false "user's ethereum signaturemessage"
// @Param signature header string false "user's ethereum signature"
// @Param	address		query 	string	false		"user's ethereum address"
// @router /user/followers [get]
func GetUserFollowers(ctx *gin.Context) {
}

// @Tags User
// @Title FollowUser
// @Description follow a user
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	address		path 	string	false		"the ethereum address of the user who you want to follow"
// @router /user/follow/{address} [post]
func FollowUser(ctx *gin.Context) {
}

// @Tags User
// @Title UnFollowUser
// @Description cancel following of a user
// @Param address header string true "user's ethereum address"
// @Param signaturemessage header string true "user's ethereum signaturemessage"
// @Param signature header string true "user's ethereum signature"
// @Param	address		path 	string	false		"the ethereum address of the user who you want to unfollow"
// @router /user/follow/{address} [delete]
func UnFollowUser(ctx *gin.Context) {
}