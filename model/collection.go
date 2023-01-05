package model

import (
	"errors"
	"fmt"
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
	Collection   Collection `gorm:"foreignKey:Id;references:CollectionId"`
	EthAddr      string
}

type CollectionStar struct {
	SaoModel
	CollectionId uint
	Collection   Collection `gorm:"foreignKey:Id;references:CollectionId"`
	EthAddr      string
}

type CollectionFile struct {
	SaoModel
	Collection   Collection `gorm:"foreignKey:Id;references:CollectionId"`
	CollectionId uint
	FileId       uint
	EthAddr      string
	Status       int
}

type CollectionRequest struct {
	CollectionIds []uint
	FileId        uint
	Status        int
}

type CollectionVO struct {
	Id           uint
	CreatedAt    int64
	UpdatedAt    int64
	EthAddr      string
	Preview      string
	Labels       string
	Title        string
	Description  string
	TotalFiles   int64
	MaxFiles     int64
	Type         int
	Liked        bool
	TotalLikes   int64
	FileIncluded bool
}

func (model *Model) CreateCollection(collection *Collection) error {
	return model.DB.Create(collection).Error
}

func (model *Model) UpsertCollection(collection *Collection) error {
	if collection.Id <= 0 {
		return model.DB.Create(collection).Error
	}
	var c Collection
	result := model.DB.Where("id = ?", collection.Id).First(&c)
	if result.Error != nil {
		return result.Error
	}
	if c.Id > 0 {
		return model.DB.Where("id = ?", collection.Id).Updates(collection).Error
	}
	return nil
}

func (model *Model) DeleteCollection(collectionId uint) error {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Collection{}).Where("id = ?", collectionId).Delete(&Collection{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&CollectionFile{}).Where("collection_id = ?", collectionId).Delete(&CollectionFile{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&CollectionLike{}).Where("collection_id = ?", collectionId).Delete(&CollectionLike{}).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

func (model *Model) GetSearchCollectionResult(key string) (*[]CollectionVO, error) {
	var collections []Collection
	bindKey := "%" + key + "%"
	model.DB.Where("(title like ? or labels like ? or `description` like ?) and type = 0", bindKey, bindKey, bindKey).Find(&collections)

	var result []CollectionVO
	for _, c := range collections {
		result = append(result, CollectionVO{
			Id:           c.Id,
			CreatedAt:    c.CreatedAt.UnixMilli(),
			UpdatedAt:    c.UpdatedAt.UnixMilli(),
			EthAddr:      c.EthAddr,
			Preview:      fmt.Sprintf("%s/%s/%s", model.Config.ApiServer.Host, "previews", c.Preview),
			Title:        c.Title,
			Labels:       c.Labels,
			Description:  c.Description,
			Type:         c.Type,
			MaxFiles:     100,
		})
	}
	return &result, nil
}

func (model *Model) GetCollection(collectionId uint, ethAddr string, fileID uint, address string, offset int, limit int) (*[]CollectionVO, error) {
	var collections []Collection
	if collectionId > 0 {
		var collection Collection
		result := model.DB.First(&collection, collectionId)
		if result.Error != nil {
			return nil, result.Error
		}
		collections = append(collections, collection)
	} else if ethAddr != "" && fileID > 0 {
		model.DB.Where("eth_addr = ?", ethAddr).Limit(limit).Offset(offset).Find(&collections)
	} else if ethAddr != "" {
		model.DB.Where("eth_addr = ?", ethAddr).Limit(limit).Offset(offset).Find(&collections)
	} else if fileID > 0 {
		model.DB.Raw("select c.* from collections c inner join collection_files f on c.id = f.collection_id where f.deleted_at is null and c.deleted_at is null and f.file_id = ? and （type = 0 or (type = 1 and eth_addr = ?))", fileID, address).Limit(limit).Offset(offset).Find(&collections)
	}

	var result []CollectionVO
	for _, c := range collections {
		var totalFiles int64
		model.DB.Model(&CollectionFile{}).Where("collection_id = ? ", c.Id).Count(&totalFiles)

		var totalLikes int64
		model.DB.Model(&CollectionLike{}).Where("collection_id = ? ", c.Id).Count(&totalLikes)

		fileIncluded := false
		if fileID > 0 {
			var count int64
			model.DB.Model(&CollectionFile{}).Where("file_id = ? and collection_id = ? ", fileID, c.Id).Count(&count)
			if count > 0 {
				fileIncluded = true
			}
		}

		liked := false
		if collectionId > 0 {
			var count int64
			model.DB.Model(&CollectionLike{}).Where("eth_addr = ? and collection_id = ? ", address, c.Id).Count(&count)
			if count > 0 {
				liked = true
			}
		}
		result = append(result, CollectionVO{
			Id:           c.Id,
			CreatedAt:    c.CreatedAt.UnixMilli(),
			UpdatedAt:    c.UpdatedAt.UnixMilli(),
			EthAddr:      c.EthAddr,
			Preview:      fmt.Sprintf("%s/%s/%s", model.Config.ApiServer.Host, "previews", c.Preview),
			Title:        c.Title,
			Labels:       c.Labels,
			Description:  c.Description,
			Type:         c.Type,
			TotalFiles:   totalFiles,
			MaxFiles:     100,
			FileIncluded: fileIncluded,
			TotalLikes:   totalLikes,
			Liked:        liked,
		})
	}
	return &result, nil
}

func (model *Model) AddFileToCollections(fileId uint, collectionIds []uint, ethAddr string) error {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		tx.Model(&FilePreview{}).Where("id = ? ", fileId).Count(&count)
		if count <= 0 {
			return xerrors.Errorf("file id not exist: %d", fileId)
		}
		tx.Model(&CollectionFile{}).Where("file_id = ?", fileId).Delete(&CollectionFile{})
		for _, collectionId := range collectionIds {
			collectionFile := CollectionFile{
				CollectionId: collectionId,
				FileId:       fileId,
				EthAddr:      ethAddr,
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
		EthAddr:      ethAddress,
	}
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		tx.Model(&Collection{}).Where("id = ? ", collectionLike.CollectionId).Count(&count)
		if count <= 0 {
			return xerrors.Errorf("the collection not exist : %d", collectionLike.CollectionId)
		}
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
			return errors.New("the user" + ethAddress + " haven't clicked like yet:" + string(collectionId))
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
		EthAddr:      ethAddress,
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
