package store

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sao-datastore-storage/common"
	go_mcs_sdk "sao-datastore-storage/go-mcs-sdk"
	"sao-datastore-storage/model"
)

const MCS_DURATION = 525
const MCS_FILE_TYPE = 0

type McsStore struct {
	enableFilecoin bool
	mcsClient      *go_mcs_sdk.McsClient
}

func NewMcsStore(config common.McsInfo) McsStore {
	mcsClient := go_mcs_sdk.NewMcsClient(config.ProviderRpc, config.McsEndpoint, config.StorageEndpoint)
	mcsClient.SetAccount(config.PrivateKey)
	return McsStore{
		mcsClient:      mcsClient,
		enableFilecoin: config.EnableFilecoin,
	}
}

func (s McsStore) StoreFile(ctx context.Context, reader io.Reader, info map[string]string) (StoreRet, error) {
	jsonResp, err := s.mcsClient.Upload(info["filename"], reader, map[string]string{
		"duration": fmt.Sprintf("%d", MCS_DURATION),
		"fileType": fmt.Sprintf("%d", MCS_FILE_TYPE),
	})
	if err != nil {
		return StoreRet{}, err
	}

	mcsInfo := model.McsInfo{
		SourceFileUploadId: jsonResp.Data.SourceFileUploadId,
		PayloadCid:         jsonResp.Data.PayloadCid,
		IpfsUrl:            jsonResp.Data.IpfsUrl,
		FileSize:           jsonResp.Data.FileSize,
		WCid:               jsonResp.Data.WCid,
	}
	if s.enableFilecoin {
		tx, err := s.mcsClient.MakePayment(jsonResp.Data.WCid, jsonResp.Data.FileSize, MCS_DURATION)
		if err != nil {
			return StoreRet{}, err
		}
		mcsInfo.PaymentTxHash = tx
	}
	return StoreRet{
		McsInfo: &mcsInfo,
	}, nil
}
func (s McsStore) GetFile(ctx context.Context, info map[string]string) (io.ReadCloser, error) {
	resp, err := http.Get(info["hash"])
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (s McsStore) DeleteFile(ctx context.Context, info map[string]string) error {
	return nil
}
