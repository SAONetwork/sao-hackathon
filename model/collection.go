package model

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
	EthAddr      string
}

type CollectionStar struct {
	SaoModel
	CollectionId uint
	EthAddr      string
}

type CollectionFiles struct {
	SaoModel
	FileId  uint
	EthAddr string
	Status  int
}

type CollectionVO struct {
	ID          uint
	Preview     string
	Labels      string
	Title       string
	Description string
	Type        int
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

func (model *Model) GetCollection(collectionId uint, ethAddr string, fileID uint) (*[]Collection, error) {
	var collections []Collection
	if collectionId > 0 {
		var collection Collection
		result := model.DB.First(&collection, collectionId)
		if result.Error != nil {
			return nil, result.Error
		}
		collections = append(collections, collection)
		return &collections, nil
	}

	if ethAddr != "" {
		model.DB.Where("eth_addr = ?", ethAddr).Find(&collections)
		return &collections, nil
	}

	if fileID > 0 {

	}
	return nil, nil
}