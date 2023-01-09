package model

import (
	"errors"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

type FileInfo struct {
	SaoModel
	Filename        string `json:"filename" gorm:"column:filename;type:varchar(255) ;default:''"`
	ContentType     string `json:"contentType" gorm:"column:contentType;type:varchar(255) ;default:''"`
	Size            int64
	ExpireAt        int64  `json:"-" gorm:"column:expireAt"`
	IpfsHash        string `json:"ipfsHash" gorm:"column:ipfsHash;type:varchar(255) ;default:''"`
	McsInfoId       uint   `json:"mcsInfoId" gorm:"column:mcsInfoId"`
	Cid             string `json:"cid" gorm:"column:cid;type:varchar(255) ;default:''"`
	StorageProvider string `json:"storageProvider" gorm:"column:storageProvider;type:varchar(255) ;default:''"`
	Status          uint   `json:"-" gorm:"column:status;type:int(11)"`
}

type McsInfo struct {
	SaoModel
	SourceFileUploadId int64
	PayloadCid         string
	IpfsUrl            string
	FileSize           int64
	WCid               string
	PaymentTxHash      string
}

type FileInfoInMarket struct {
	Id             uint
	CreatedAt      time.Time
	UpdatedAt      time.Time
	EthAddr        string
	Preview        string
	Labels         string
	Price          decimal.Decimal
	Title          string
	Description    string
	ContentType    string
	Type           int
	Status         FilePreviewStatus
	NftTokenId     int64
	FileCategory   FileCategory
	AlreadyPaid    bool
	AdditionalInfo string
	FileExtension  string
	WCid           string
	Star           bool
}

type FileDetail struct {
	FileInfoInMarket
	IpfsHash        string
	Size            int64
	Cid             string
	StorageProvider string
	TotalComments int64
	TotalCollections int64
}

type PagedFileInfoInMarket struct {
	FileInfoInMarkets []FileInfoInMarket
	Total             int64
}

func (model *Model) CountFileByFilenameAndStatus(dest string, status int) (int64, error) {
	var count int64
	result := model.DB.Model(&FileInfo{}).Where("filename = ? and status = ?", dest, status).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

func (model *Model) GetFileInfoByPreviewId(fileId uint) *FileInfo {
	var file FileInfo
	model.DB.Raw("SELECT i.id, i.contentType, i.ipfsHash, p.filename, i.mcsInfoId FROM file_infos i, file_previews p WHERE p.file_id = i.id and p.id = ?", fileId).Scan(&file)
	return &file
}

func (model *Model) StoreFile(file FileInfo, mcsInfo *McsInfo) (*FileInfo, error) {
	var count int64
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		exists := false
		if file.IpfsHash != "" {
			tx.Model(&FileInfo{}).Where("ipfsHash = ? and status = 0", file.IpfsHash).Count(&count)
			exists = count > 0
		} else if mcsInfo != nil {
			// TODO: check mcsinfo exists.
		}
		if !exists {
			if mcsInfo != nil {
				if err := model.DB.Create(mcsInfo).Error; err != nil {
					return err
				}

				file.McsInfoId = mcsInfo.Id
			}
			if err := tx.Create(&file).Error; err != nil {
				return err
			}
		} else {
			if err := model.DB.Model(&FileInfo{}).Where("ipfsHash = ?", file.IpfsHash).Update("filename", file.Filename).Error; err != nil {
				return err
			}
			model.DB.Where("ipfsHash = ? and status = 0", file.IpfsHash).First(&file)
		}

		return nil
	})
	return &file, err
}

func (model *Model) StoreMcsInfo(info *McsInfo) (*McsInfo, error) {
	err := model.DB.Create(info).Error
	return info, err
}

func (model *Model) GetMcsInfoById(id uint) (*McsInfo, error) {
	var info McsInfo
	result := model.DB.First(&info, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &info, nil
}

func (model *Model) DeleteFile(preview *FilePreview) (string, error) {
	var ipfsHash string
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var ipfsFileInfo FileInfo
		if err := tx.Model(&FileInfo{}).Where("id = ?", preview.FileId).Find(&ipfsFileInfo).Error; err != nil {
			return errors.New("ipfs file not found in system")
		}
		ipfsHash = ipfsFileInfo.IpfsHash

		if err := tx.Model(&CollectionFile{}).Where("file_id = ?", preview.Id).Delete(&CollectionFile{}).Error; err != nil {
			return err
		}

		if err := tx.Raw("delete l from file_comment_likes l INNER JOIN file_comments c ON l.comment_id = c.id WHERE c.file_id =  ?", preview.Id).Find(&FileCommentLike{}).Error; err != nil {
			return err
		}

		if err := tx.Model(&FileComment{}).Where("file_id = ?", preview.Id).Delete(&FileComment{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&preview).Error; err != nil {
			return err
		}

		if err := tx.Delete(&ipfsFileInfo).Error; err != nil {
			return err
		}
		return nil
	})
	return ipfsHash, err
}