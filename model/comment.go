package model

type FileComment struct {
	SaoModel
	EthAddr  string
	Comment   string
	ParentId  uint
}

type FileCommentLike struct {
	SaoModel
	EthAddr  string
	CommentId   uint
}

type CommentVO struct {
	Id uint
	DateTime int64
	EthAddr string
	Avatar string
	Comment string
	SubComments []SubCommentVO
}

type SubCommentVO struct {
	Id uint
	DateTime int64
	EthAddr string
	Avatar string
	Comment string
}