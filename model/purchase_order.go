package model

type PurchaseOrder struct {
	Id             uint
	FileId         int64
	BuyerAddr      string
	OrderTxHash    string
	CompleteTxHash string
	State          OrderState
}

func (model *Model) GetPurchaseOrder(fileId uint, ethAddress string) *PurchaseOrder {
	var purchaseOrder PurchaseOrder
	model.DB.Model(&PurchaseOrder{}).Where("file_id", fileId).Where("buyer_addr", ethAddress).First(&purchaseOrder)
	return &purchaseOrder
}

func (model *Model) GetFinishPurchaseOrders() *[]PurchaseOrder {
	var purchaseOrders []PurchaseOrder
	model.DB.Model(&PurchaseOrder{}).Where("state = ?", ReadyToDownload).Find(&purchaseOrders)
	return &purchaseOrders
}

func (model *Model) CreatePurchaseOrder(purchaseOrder *PurchaseOrder) error {
	return model.DB.Create(purchaseOrder).Error
}

func (model *Model) UpdatePurchaseOrderState(orderId uint, state OrderState) error {
	return model.DB.Model(&PurchaseOrder{}).Where("Id = ?", orderId).Update("state", state).Error
}