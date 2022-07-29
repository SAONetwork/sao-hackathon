package model

type FileChunkMetadata struct {
	SaoModel
	FileId          uint
	Offset          int64
	Size            int64
	EncryptedOffset int64
	EncryptedSize   int64
}

func (model *Model) GetFileChunkMetadatasByFileId(fileId uint) []FileChunkMetadata {
	var fileChunkMetadatas []FileChunkMetadata
	model.DB.Where("file_id", fileId).Find(&fileChunkMetadatas)
	return fileChunkMetadatas
}