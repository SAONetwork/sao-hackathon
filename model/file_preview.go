package model

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gwaylib/log"
	"github.com/shopspring/decimal"
	"math/big"
	"path/filepath"

	"gorm.io/gorm"
)

type FilePreview struct {
	SaoModel
	FileId         uint
	EthAddr        string
	Preview        string `gorm:"type:longtext;"`
	Labels         string
	TmpPath        string
	Price          decimal.Decimal `gorm:"type:decimal(32,18);"`
	Title          string
	Description    string
	ContentType    string
	Type           int
	Filename       string
	Status         FilePreviewStatus
	FileCategory   FileCategory
	NftTokenId     int64
	AdditionalInfo string
}

type FileStar struct {
	SaoModel
	FilePreviewId uint
	FilePreview FilePreview `gorm:"foreignKey:Id;references:FilePreviewId"`
	EthAddr      string
}

type FilePreviewVO struct {
	ID          uint
	Preview     string
	Labels      string
	Title       string
	Description string
	Type        int
}

func (model *Model) GetFilePreviewById(Id uint) (*FilePreview, error) {
	var file FilePreview
	result := model.DB.First(&file, Id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &file, nil
}

func (model *Model) CreateFilePreview(preview *FilePreview) error {
	return model.DB.Create(preview).Error
}

func (model *Model) GetFilePreviewByFileId(fileId uint) (*FilePreview, error) {

	var file FilePreview
	result := model.DB.Model(&FilePreview{}).Where("Id", fileId).First(&file)
	if result.Error != nil {
		return nil, result.Error
	}
	return &file, nil
}

func (model *Model) GetFilePreviewByTokenId(tokenId int64) (*FilePreview, error) {

	var filePreview FilePreview
	result := model.DB.Model(&FilePreview{}).Where("nft_token_id", tokenId).First(&filePreview)
	if result.Error != nil {
		return nil, result.Error
	}
	return &filePreview, nil
}

func (model *Model) DeletePreview(filepreview *FilePreview) error {
	return model.DB.Delete(filepreview).Error
}

func (model *Model) GetSuccessUploadFileCount() int64 {
	var count int64
	model.DB.Model(&FilePreview{}).Where("status", PlacedToIpfs).Count(&count)
	return count
}

func (model *Model) GetFileInfo(fileId uint, ethAddress string) (*FileDetail, error) {
	filePreview, err := model.GetFilePreviewByFileId(fileId)
	if err != nil {
		return nil, errors.New("file id not found in system")
	}
	paid := false
	if filePreview.Price.Cmp(decimal.NewFromInt(0))> 0 && filePreview.EthAddr != ethAddress {
		purchaseOrder := model.GetPurchaseOrder(fileId, ethAddress)
		if purchaseOrder.FileId > 0 {
			paid = true
		}
	}
	star := false
	if ethAddress != "" {
		var starCount int64
		model.DB.Model(&CollectionFile{}).Where("eth_addr = ? and file_id = ? ", ethAddress, filePreview.Id).Count(&starCount)
		if starCount > 0 {
			star = true
		}
	}

	var ipfsFileInfo FileInfo
	if err := model.DB.Model(&FileInfo{}).Where("id = ?", filePreview.FileId).Find(&ipfsFileInfo).Error; err != nil {
		return nil, errors.New("ipfs file not found in system")
	}
	fileExtension := filepath.Ext(filePreview.Filename)
	if fileExtension != "" {
		fileExtension = fileExtension[1:]
	}

	var TotalComments int64
	model.DB.Model(&FileComment{}).Where("status <> 2 and file_id = ? ", filePreview.Id).Count(&TotalComments)

	var TotalCollections int64
	err = model.DB.Raw("select count(*) from collections c inner join collection_files f on c.id = f.collection_id where f.deleted_at is null and c.deleted_at is null and f.file_id = ? and (type = 0 or (type = 1 and c.eth_addr = ?))", filePreview.Id, ethAddress).Find(&TotalCollections).Error
	if err != nil {
		log.Error(err)
	}

	filesInfoInMarket := FileDetail{
		FileInfoInMarket: FileInfoInMarket{Id: filePreview.Id,
			CreatedAt:      filePreview.CreatedAt,
			UpdatedAt:      filePreview.UpdatedAt,
			EthAddr:        filePreview.EthAddr,
			Preview:        fmt.Sprintf("%s/previews/%s", model.Config.ApiServer.Host, filePreview.Preview),
			Labels:         filePreview.Labels,
			Price:          filePreview.Price,
			Title:          filePreview.Title,
			Description:    filePreview.Description,
			ContentType:    filePreview.ContentType,
			Type:           filePreview.Type,
			Status:         filePreview.Status,
			NftTokenId:     filePreview.NftTokenId,
			FileCategory:   filePreview.FileCategory,
			AdditionalInfo: filePreview.AdditionalInfo,
			AlreadyPaid:    paid,
			FileExtension:  fileExtension,
			Star:           star},
		IpfsHash:         ipfsFileInfo.IpfsHash,
		Size:             ipfsFileInfo.Size,
		Cid:              ipfsFileInfo.Cid,
		StorageProvider:  ipfsFileInfo.StorageProvider,
		TotalComments:    TotalComments,
		TotalCollections: TotalCollections}
	return &filesInfoInMarket, nil
}

func (model *Model) GetFileInfosByCollectionId(collectionId string, ethAddress string, offset int, limit int) []FileInfoInMarket {
	var filePreviews []FilePreview
	model.DB.Offset(offset).Limit(limit).Raw("select p.* from file_previews p, collection_files f where f.file_id = p.id and f.deleted_at is null and p.deleted_at is null and f.collection_id = ?", collectionId).Scan(&filePreviews)

	filesInfoInMarket := make([]FileInfoInMarket, 0)

	for _, filePreview := range filePreviews {
		paid := false
		if filePreview.Price.Cmp(decimal.NewFromInt(0))> 0 && ethAddress != "" {
			order := model.GetPurchaseOrder(filePreview.Id, ethAddress)
			if order.FileId > 0 {
				paid = true
			}
		}
		star := false
		if ethAddress != "" {
			var starCount int64
			model.DB.Model(&CollectionFile{}).Where("eth_addr = ? and file_id = ? ", ethAddress, filePreview.Id).Count(&starCount)
			if starCount > 0 {
				star = true
			}
		}
		fileExtension := filepath.Ext(filePreview.Filename)
		if fileExtension != "" {
			fileExtension = fileExtension[1:]
		}
		filesInfoInMarket = append(filesInfoInMarket, FileInfoInMarket{Id: filePreview.Id,
			CreatedAt:    filePreview.CreatedAt,
			UpdatedAt:    filePreview.UpdatedAt,
			EthAddr:      filePreview.EthAddr,
			Preview:      fmt.Sprintf("%s/previews/%s", model.Config.ApiServer.Host, filePreview.Preview),
			Labels:       filePreview.Labels,
			Price:        filePreview.Price,
			Title:        filePreview.Title,
			Description:  filePreview.Description,
			ContentType:  filePreview.ContentType,
			Type:         filePreview.Type,
			Status:       filePreview.Status,
			NftTokenId:   filePreview.NftTokenId,
			FileCategory: filePreview.FileCategory,
			AdditionalInfo: filePreview.AdditionalInfo,
			FileExtension: fileExtension,
			AlreadyPaid:  paid,
			Star: star})
	}
	return filesInfoInMarket
}

func (model *Model) GetSearchFileResult(key string, ethAddress string, offset int, limit int) []FileInfoInMarket {
	var filePreviews []FilePreview

	if common.IsHexAddress(key) {
		model.DB.Offset(offset).Limit(limit).Model(&FilePreview{}).Where("status = 2").Where("price = 0 or (price > 0 and nft_token_id > 0)").Where("eth_addr = ?", key).Order("created_at desc").Find(&filePreviews)
	} else {
		bindKey := "%" + key + "%"
		model.DB.Offset(offset).Limit(limit).Raw("select *,\n"+
			"       case when title like ? then 3 else 0 end + \n"+
			"       case when filename like ? then 2 else 0 end + \n"+
			"       case when labels like ? then 3 else 0 end + \n"+
			"       case when `description` like ? then 1 else 0 end as matches \n"+
			"  from file_previews \n"+
			" where (title like ?\n"+
			"    or labels like ?\n"+
			"    or `description` like ?)\n"+
			"    and status = 2\n"+
			" order by matches desc", bindKey, bindKey, bindKey, bindKey, bindKey, bindKey, bindKey, bindKey).Scan(&filePreviews)
	}

	filesInfoInMarket := make([]FileInfoInMarket, 0)

	for _, filePreview := range filePreviews {
		paid := false
		if filePreview.Price.Cmp(decimal.NewFromInt(0))> 0 && ethAddress != "" {
			order := model.GetPurchaseOrder(filePreview.Id, ethAddress)
			if order.FileId > 0 {
				paid = true
			}
		}
		star := false
		if ethAddress != "" {
			var starCount int64
			model.DB.Model(&CollectionFile{}).Where("eth_addr = ? and file_id = ? ", ethAddress, filePreview.Id).Count(&starCount)
			if starCount > 0 {
				star = true
			}
		}
		fileExtension := filepath.Ext(filePreview.Filename)
		if fileExtension != "" {
			fileExtension = fileExtension[1:]
		}
		filesInfoInMarket = append(filesInfoInMarket, FileInfoInMarket{Id: filePreview.Id,
			CreatedAt:    filePreview.CreatedAt,
			UpdatedAt:    filePreview.UpdatedAt,
			EthAddr:      filePreview.EthAddr,
			Preview:      fmt.Sprintf("%s/previews/%s", model.Config.ApiServer.Host, filePreview.Preview),
			Labels:       filePreview.Labels,
			Price:        filePreview.Price,
			Title:        filePreview.Title,
			Description:  filePreview.Description,
			ContentType:  filePreview.ContentType,
			Type:         filePreview.Type,
			Status:       filePreview.Status,
			NftTokenId:   filePreview.NftTokenId,
			FileCategory: filePreview.FileCategory,
			AdditionalInfo: filePreview.AdditionalInfo,
			FileExtension: fileExtension,
			AlreadyPaid:  paid,
		Star: star})
	}
	return filesInfoInMarket
}

func (model *Model) GetMarketFiles(limit int, offset int, ethAddress string, condition map[string]interface{}, price int) ([]FileInfoInMarket, int64) {
	var filePreviews []FilePreview
	if price > 0 {
		model.DB.Offset(offset).Limit(limit).Model(&FilePreview{}).Where("status = 2").Where("price > 0 and nft_token_id > 0").Where(condition).Order("created_at desc").Find(&filePreviews)
	} else if price == 0{
		model.DB.Offset(offset).Limit(limit).Model(&FilePreview{}).Where("status = 2").Where("price = 0").Where(condition).Order("created_at desc").Find(&filePreviews)
	} else {
		model.DB.Offset(offset).Limit(limit).Model(&FilePreview{}).Where("status = 2").Where("price = 0 or (price > 0 and nft_token_id > 0)").Where(condition).Order("created_at desc").Find(&filePreviews)
	}
	filesInfoInMarket := make([]FileInfoInMarket, 0)
	var count int64
	if price > 0 {
		model.DB.Model(&FilePreview{}).Where("status = 2").Where("price > 0 and nft_token_id > 0").Where(condition).Count(&count)
	} else if price == 0{
		model.DB.Model(&FilePreview{}).Where("status = 2").Where("price = 0").Where(condition).Count(&count)
	} else {
		model.DB.Model(&FilePreview{}).Where("status = 2").Where("price = 0 or (price > 0 and nft_token_id > 0)").Where(condition).Count(&count)
	}
	if count <= 0 {
		return nil, 0
	}
	for _, filePreview := range filePreviews {
		paid := false
		if filePreview.Price.Cmp(decimal.NewFromInt(0))> 0 && ethAddress != "" {
			order := model.GetPurchaseOrder(filePreview.Id, ethAddress)
			if order.FileId > 0 {
				paid = true
			}
		}
		star := false
		if ethAddress != "" {
			var starCount int64
			model.DB.Model(&CollectionFile{}).Where("eth_addr = ? and file_id = ? ", ethAddress, filePreview.Id).Count(&starCount)
			if starCount > 0 {
				star = true
			}
		}
		fileExtension := filepath.Ext(filePreview.Filename)
		if fileExtension != "" {
			fileExtension = fileExtension[1:]
		}
		filesInfoInMarket = append(filesInfoInMarket, FileInfoInMarket{Id: filePreview.Id,
			CreatedAt:      filePreview.CreatedAt,
			UpdatedAt:      filePreview.UpdatedAt,
			EthAddr:        filePreview.EthAddr,
			Preview:        fmt.Sprintf("%s/previews/%s", model.Config.ApiServer.Host, filePreview.Preview),
			Labels:         filePreview.Labels,
			Price:          filePreview.Price,
			Title:          filePreview.Title,
			Description:    filePreview.Description,
			ContentType:    filePreview.ContentType,
			Type:           filePreview.Type,
			Status:         filePreview.Status,
			NftTokenId:     filePreview.NftTokenId,
			FileCategory:   filePreview.FileCategory,
			AdditionalInfo: filePreview.AdditionalInfo,
			FileExtension:  fileExtension,
			AlreadyPaid:    paid,
			Star:           star})
	}
	return filesInfoInMarket, count
}

func (model *Model) UpdatePreviewLinkedWithIpfs(Id uint, updates map[string]interface{}) error {
	return model.DB.Model(&FilePreview{}).Where("id", Id).Updates(updates).Error
}

func (model *Model) UpdatePreview(Id uint, updates map[string]interface{}) error {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		if err = tx.Model(&FilePreview{}).Where("id", Id).Updates(updates).Error; err != nil {
			return err
		}
		if err = tx.Model(&FilePreview{}).Where("id", Id).Where("status != 2").Update("status", 1).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

func (model *Model) UpdatePreviewPriceAndTokenId(fileId int64, price *big.Int, tokenId *big.Int) error {
	ethPrice := decimal.NewFromBigInt(price, -18)
	return model.DB.Model(&FilePreview{}).Where("ID = ?", uint(fileId)).Update("price", ethPrice).Update("nft_token_id", tokenId.Int64()).Error
}

func (model *Model) StoreFileMetadata(chunkMetadatas []FileChunkMetadata, fileId uint, previewId int64) error {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		if err = tx.Create(chunkMetadatas).Error; err != nil {
			return err
		}
		updateMap := map[string]interface{}{
			"Status": PlacedToIpfs,
			"FileId": fileId,
		}
		if err = tx.Model(&FilePreview{}).Where("id", previewId).Updates(updateMap).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

func (model *Model) StarFile(ethAddress string, fileId uint) error {
	fileLike := FileStar{
		FilePreviewId: fileId,
		EthAddr:      ethAddress,
	}
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		tx.Model(&FileStar{}).Where("eth_addr = ? and file_preview_id = ? ", ethAddress, fileId).Count(&count)
		if count <= 0 {
			if err := tx.Create(&fileLike).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (model *Model) DeleteStarFile(ethAddress string, fileId uint) error {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		tx.Model(&FileStar{}).Where("eth_addr = ? and file_preview_id = ? ", ethAddress, fileId).Count(&count)
		if count <= 0 {
			return errors.New("the user" + ethAddress + " haven't clicked star yet:" + string(fileId))
		}

		if err := tx.Where("eth_addr = ? and file_preview_id = ? ", ethAddress, fileId).Delete(&FileStar{}).Error; err != nil {
			return err
		}

		return nil
	})
	return err
}