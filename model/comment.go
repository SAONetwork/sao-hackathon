package model

import (
	"errors"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type FileComment struct {
	SaoModel
	EthAddr  string
	Comment  string
	FileId   uint
	ParentId uint
	Children string
	Status int `gorm:"type:int(11);default:0"`
}

type FileCommentLike struct {
	SaoModel
	EthAddr   string
	CommentId uint
}

type CommentVO struct {
	Id          uint
	ObjectId    string
	DateTime    int64
	EthAddr     string
	UserName    string
	Editable    bool
	Avatar      string
	Comment     string
	Liked bool
	TotalLikes    int64
	ParentComment *ParentCommentVO
}

type ParentCommentVO struct {
	Id         uint
	ObjectId   string
	DateTime   int64
	EthAddr    string
	UserName   string
	Avatar     string
	Comment    string
	Status string
}

func (model *Model) AddFileComment(comment *FileComment) (*CommentVO, error) {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&FileComment{}).Create(&comment).Error; err != nil {
			return err
		}

		if comment.ParentId > 0 {
			var parentComment FileComment
			tx.Model(&FileComment{}).Where("id = ?", comment.ParentId).First(&parentComment)
			if parentComment.Id <= 0 {
				return xerrors.Errorf("parent not found: %d", comment.ParentId)
			}
			childrenIds := strings.Split(parentComment.Children, ",")
			childrenIds = append([]string{strconv.FormatUint(uint64(comment.Id), 10)}, childrenIds...)
			if err := tx.Model(&FileComment{}).Where("id = ?", comment.ParentId).Update("children", strings.Join(childrenIds, ",")).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var user UserProfile
	model.DB.Model(&UserProfile{}).Where("eth_addr = ?", comment.EthAddr).First(&user)
	commentVO := CommentVO{Id: comment.Id, ObjectId: strconv.FormatUint(uint64(comment.FileId), 10), DateTime: comment.CreatedAt.UnixMilli(), EthAddr: comment.EthAddr, Comment: comment.Comment, UserName: user.Username, Avatar: user.Avatar,
		Editable: true}
	return &commentVO, nil
}

func (model *Model) DeleteFileComment(commentId uint) error {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var toDelete FileComment
		tx.Model(&FileComment{}).Where("id = ?", commentId).First(&toDelete)
		if toDelete.Id <= 0 {
			return xerrors.Errorf("The comment is not existing: %d", commentId)
		}
		if err := tx.Model(&FileComment{}).Where("id = ?", commentId).Update("status", 2).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

func (model *Model) GetFileComment(fileId uint, address string) (*[]CommentVO, error) {
	var comments []FileComment
	model.DB.Order("id desc").Where("status != 2 and file_id = ?", fileId).Find(&comments)

	var result []CommentVO
	for _, comment := range comments {
		var user UserProfile
		model.DB.Model(&UserProfile{}).Where("eth_addr = ?", comment.EthAddr).First(&user)

		var liked int64
		model.DB.Model(&FileCommentLike{}).Where("eth_addr = ? and comment_id = ? ", address, comment.Id).Count(&liked)

		var totalLikes int64
		model.DB.Model(&FileCommentLike{}).Where("comment_id = ? ", comment.Id).Count(&totalLikes)

		commentVO := CommentVO{Id: comment.Id, ObjectId: strconv.FormatUint(uint64(fileId), 10), DateTime: comment.CreatedAt.UnixMilli(), EthAddr: comment.EthAddr, Comment: comment.Comment, UserName: user.Username, Avatar: user.Avatar,
			Editable: comment.EthAddr == address, Liked: liked>0, TotalLikes: totalLikes}

		if comment.ParentId > 0 {
			var parentComment FileComment
			model.DB.Where("id = ?", comment.ParentId).Find(&parentComment)

			var parentCommentVO ParentCommentVO
			if parentComment.Status == 2{
				parentCommentVO = ParentCommentVO{Id: parentComment.Id, ObjectId: strconv.FormatUint(uint64(fileId), 10), DateTime: parentComment.CreatedAt.UnixMilli(), EthAddr: parentComment.EthAddr, Status: "deleted"}
			} else {
				var subCommentUser UserProfile
				model.DB.Model(&UserProfile{}).Where("eth_addr = ?", comment.EthAddr).First(&subCommentUser)
				parentCommentVO = ParentCommentVO{Id: parentComment.Id, ObjectId: strconv.FormatUint(uint64(fileId), 10), DateTime: parentComment.CreatedAt.UnixMilli(), EthAddr: parentComment.EthAddr, Comment: parentComment.Comment, UserName: subCommentUser.Username, Avatar: subCommentUser.Avatar}
			}
			commentVO.ParentComment = &parentCommentVO
		}

		result = append(result, commentVO)
	}
	return &result, nil
}

func (model *Model) LikeFileComment(ethAddress string, commentId uint) error {
	commentLike := FileCommentLike{
		CommentId: commentId,
		EthAddr:   ethAddress,
	}
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		tx.Model(&FileComment{}).Where("id = ? ", commentLike.CommentId).Count(&count)
		if count <= 0 {
			return xerrors.Errorf("the comment not exist : %d", commentLike.CommentId)
		}
		tx.Model(&FileCommentLike{}).Where("eth_addr = ? and comment_id = ? ", ethAddress, commentId).Count(&count)
		if count <= 0 {
			if err := tx.Create(&commentLike).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (model *Model) UnlikeFileComment(ethAddress string, commentId uint) error {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		tx.Model(&FileCommentLike{}).Where("eth_addr = ? and comment_id = ? ", ethAddress, commentId).Count(&count)
		if count <= 0 {
			return errors.New("the user" + ethAddress + " haven't clicked like yet:" + strconv.FormatUint(uint64(commentId), 10))
		}

		if err := tx.Where("eth_addr = ? and comment_id = ? ", ethAddress, commentId).Delete(&FileCommentLike{}).Error; err != nil {
			return err
		}

		return nil
	})
	return err
}
