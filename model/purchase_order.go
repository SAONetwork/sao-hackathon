package model

import "time"

type PurchaseOrder struct {
	Id             uint	`gorm:"autoIncrement:false"`
	FileId         int64
	BuyerAddr      string
	OrderTxHash    string
	CompleteTxHash string
	State          OrderState
	UpdatedAt    time.Time
}

func (model *Model) GetPurchaseOrder(fileId uint, ethAddress string) *PurchaseOrder {
	var purchaseOrder PurchaseOrder
	model.DB.Model(&PurchaseOrder{}).Where("file_id", fileId).Where("buyer_addr", ethAddress).First(&purchaseOrder)
	return &purchaseOrder
}

func (model *Model) GetNextPurchaseOrderToFinish() (*PurchaseOrder, bool) {
	var count int64
	model.DB.Model(&PurchaseOrder{}).Where("state = ?", FinishContractStarted).Where("updated_at > ?", time.Now().Add(-time.Minute * 5)).Count(&count)
	if count >0 {
		return nil, false
	}
	var purchaseOrder PurchaseOrder

	// the contract started but not finished - retry
	model.DB.Model(&PurchaseOrder{}).Where("state = ?", FinishContractStarted).Where("updated_at < ?", time.Now().Add(-time.Minute * 5)).First(&purchaseOrder)
	if purchaseOrder.Id > 0 {
		return &purchaseOrder, true
	}

	model.DB.Model(&PurchaseOrder{}).Where("state = ?", ReadyToDownload).First(&purchaseOrder)
	if purchaseOrder.Id > 0 {
		return &purchaseOrder, true
	} else {
		return nil, false
	}
}

func (model *Model) CreatePurchaseOrder(purchaseOrder map[string]interface{}) error {
	return model.DB.Model(&PurchaseOrder{}).Create(purchaseOrder).Error
}

func (model *Model) UpdatePurchaseOrderState(orderId uint, state OrderState) error {
	return model.DB.Model(&PurchaseOrder{}).Where("Id = ?", orderId).Update("state", state).Error
}