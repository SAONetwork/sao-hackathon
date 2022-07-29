package model

type OrderState string

const (
	ContractOrdered OrderState = "ContractOrdered"
	ReadyToDownload OrderState = "ReadyToDownload"
	FinishContractStarted OrderState = "FinishContractStarted"
	Finish OrderState = "Finish"
)

type FilePreviewStatus int

const (
	Uploading FilePreviewStatus = 0
	// file preview is stored in database, but not on chain.
	UploadSuccess FilePreviewStatus = 1
	// file is processed and uploaded to ipfs
	PlacedToIpfs FilePreviewStatus = 2
)

type FileCategory string

const (
	Video    FileCategory = "Video"
	Music    FileCategory = "Music"
	Document FileCategory = "Document"
	Software FileCategory = "Software"
	Image    FileCategory = "Image"
	Other    FileCategory = "Other"
)
