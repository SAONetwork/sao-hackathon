package model

type KeyStore struct {
	Id     string
	FileId string
	Offset uint64
	Size   uint64
	Key    []byte
	Nonce  []byte
}

func (model *Model) CreateKeyStore(keyStore *KeyStore) error {
	return model.DB.Create(keyStore).Error
}

func (model *Model) GetKeyStore(condition map[string]interface{}, offset uint64) (KeyStore, error) {
	var keyStore KeyStore
	err := model.DB.Model(&KeyStore{}).Where(condition).Where("Offset", offset).First(&keyStore).Error
	return keyStore, err
}