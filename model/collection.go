package model

import (
	"errors"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
)

type Collection struct {
	SaoModel
	EthAddr     string
	Preview     string `gorm:"varchar(255);"`
	Labels      string
	Title       string
	Description string
	Type        int
}

type CollectionLike struct {
	SaoModel
	CollectionId uint
	Collection Collection `gorm:"foreignKey:Id;references:CollectionId"`
	EthAddr      string
}

type CollectionStar struct {
	SaoModel
	CollectionId uint
	Collection Collection `gorm:"foreignKey:Id;references:CollectionId"`
	EthAddr      string
}

type CollectionFile struct {
	SaoModel
	Collection Collection  `gorm:"foreignKey:Id;references:CollectionId"`
	CollectionId uint
	FileId       uint
	EthAddr      string
	Status       int
}

type CollectionRequest struct {
	CollectionIds []uint
	FileId       uint
	Status       int
}

type CollectionVO struct {
	ID          uint
	Preview     string
	Labels      string
	Title       string
	Description string
	Type        int
	Liked		bool
}

func (model *Model) CreateCollection(collection *Collection) error {
	return model.DB.Create(collection).Error
}

func (model *Model) UpsertCollection(collection *Collection) (*Collection, error) {
	var c Collection
	result := model.DB.Where("id = ?", collection.Id).Assign(collection).FirstOrCreate(&c)
	if result.Error != nil {
		return nil, result.Error
	}
	return &c, nil
}

func (model *Model) DeleteCollection(collectionId uint) error {
	return model.DB.Where("id = ?", collectionId).Delete(&Collection{}).Error
}

func (model *Model) GetSearchCollectionResult(key string) (*[]Collection, error) {
	var collections []Collection
	model.DB.Where("title like '%?%' or labels like '%?%' or `description` like '%?%'", key, key, key).Find(&collections)
	return &collections, nil
}

func (model *Model) GetCollection(collectionId uint, ethAddr string, fileID uint) (*[]Collection, error) {
	var collections []Collection
	if collectionId > 0 {
		var collection Collection
		result := model.DB.First(&collection, collectionId)
		if result.Error != nil {
			return &collections, result.Error
		}
		collections = append(collections, collection)
		return &collections, nil
	}

	if ethAddr != "" {
		model.DB.Where("eth_addr = ?", ethAddr).Find(&collections)
		return &collections, nil
	}

	if fileID > 0 {
		var collectionFiles []CollectionFile
		model.DB.Where("file_id = ?", fileID).Find(&collectionFiles)
	}
	return &collections, nil
}

func (model *Model) AddFileToCollection(fileId uint, collectionId uint, ethAddr string) error {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		tx.Model(&CollectionFile{}).Where("file_id = ? and collection_id = ? ", fileId, collectionId).Count(&count)
		if count <= 0 {
			tx.Model(&FilePreview{}).Where("id = ? ", fileId).Count(&count)
			if count <= 0 {
				return xerrors.Errorf("file id not exist: %d", fileId)
			}

			collectionFile := CollectionFile{
				CollectionId: collectionId,
				FileId:  fileId,
				EthAddr: ethAddr,
			}
			if err := tx.Create(&collectionFile).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (model *Model) RemoveFileFromCollection(ethAddress string, fileId uint, collectionId uint) error {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		tx.Model(&CollectionFile{}).Where("eth_addr = ? and file_id = ? and collection_id = ? ", ethAddress, fileId, collectionId).Count(&count)
		if count <= 0 {
			return errors.New("the file is not added to the collection by eth address:" + ethAddress)
		}

		if err := tx.Where("eth_addr = ? and file_id = ? and collection_id = ? ", ethAddress, fileId, collectionId).Delete(&CollectionFile{}).Error; err != nil {
			return err
		}

		return nil
	})
	return err
}

func (model *Model) LikeCollection(ethAddress string, collectionId uint) error {
	collectionLike := CollectionLike{
		CollectionId: collectionId,
		EthAddr: ethAddress,
	}
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		tx.Model(&CollectionLike{}).Where("eth_addr = ? and collection_id = ? ", ethAddress, collectionId).Count(&count)
		if count <= 0 {
			if err := tx.Create(&collectionLike).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (model *Model) UnlikeCollection(ethAddress string, collectionId uint) error {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		tx.Model(&CollectionLike{}).Where("eth_addr = ? and collection_id = ? ", ethAddress, collectionId).Count(&count)
		if count <= 0 {
			return errors.New("the user" + ethAddress +" haven't clicked like yet:" + string(collectionId))
		}

		if err := tx.Where("eth_addr = ? and collection_id = ? ", ethAddress, collectionId).Delete(&CollectionLike{}).Error; err != nil {
			return err
		}

		return nil
	})
	return err
}

func (model *Model) StarCollection(ethAddress string, collectionId uint) error {
	collectionLike := CollectionStar{
		CollectionId: collectionId,
		EthAddr: ethAddress,
	}
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		tx.Model(&CollectionStar{}).Where("eth_addr = ? and collection_id = ? ", ethAddress, collectionId).Count(&count)
		if count <= 0 {
			if err := tx.Create(&collectionLike).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (model *Model) DeleteStarCollection(ethAddress string, collectionId uint) error {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		tx.Model(&CollectionStar{}).Where("eth_addr = ? and collection_id = ? ", ethAddress, collectionId).Count(&count)
		if count <= 0 {
			return errors.New("the user" + ethAddress + " haven't clicked like yet:" + string(collectionId))
		}

		if err := tx.Where("eth_addr = ? and collection_id = ? ", ethAddress, collectionId).Delete(&CollectionStar{}).Error; err != nil {
			return err
		}

		return nil
	})
	return err
}