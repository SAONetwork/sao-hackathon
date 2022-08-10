package store

import (
	"bytes"
	"context"
	"encoding/json"
	"golang.org/x/xerrors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"sao-datastore-storage/model"
)

type UploadResp struct {
	Status  string     `json:"status"`
	Data    UploadData `json:"data"`
	Message string     `json:"message"`
}

type UploadData struct {
	SourceFileUploadId string `json:"source_file_upload_id"`
	PayloadCid         string `json:"payload_cid"`
	IpfsUrl            string `json:"ipfs_url"`
	FileSize           int64  `json:"file_size"`
	WCid               string `json:"w_cid"`
}

const MCS_DURATION = "525"
const MCS_FILE_TYPE = "0"

type McsStore struct {
	endpoint string
}

func NewMcsStore(endpoint string) McsStore {
	return McsStore{
		endpoint: endpoint,
	}
}

func (s McsStore) StoreFile(ctx context.Context, reader io.Reader, info map[string]string) (StoreRet, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("file", "cover.jpg")
	if err != nil {
		return StoreRet{}, err
	}

	_, err = io.Copy(fw, reader)
	if err != nil {
		return StoreRet{}, err
	}
	writer.WriteField("duration", MCS_DURATION)
	writer.WriteField("file_type", MCS_FILE_TYPE)
	writer.WriteField("wallet_address", info["address"])

	err = writer.Close()
	if err != nil {
		return StoreRet{}, err
	}

	resp, err := http.Post(s.endpoint+"/storage/ipfs/upload", writer.FormDataContentType(), body)
	if err != nil {
		return StoreRet{}, err
	}
	resBody, err := ioutil.ReadAll(resp.Body)

	jsonResp := UploadResp{}
	if err = json.Unmarshal(resBody, &jsonResp); err != nil {
		return StoreRet{}, err
	}

	if jsonResp.Status == "success" {
		mcsInfo := model.McsInfo{
			SourceFileUploadId: jsonResp.Data.SourceFileUploadId,
			PayloadCid:         jsonResp.Data.PayloadCid,
			IpfsUrl:            jsonResp.Data.IpfsUrl,
			FileSize:           jsonResp.Data.FileSize,
			WCid:               jsonResp.Data.WCid,
		}

		return StoreRet{
			McsInfo: &mcsInfo,
		}, nil
	} else {
		return StoreRet{}, xerrors.New(jsonResp.Message)
	}
}
func (s McsStore) GetFile(ctx context.Context, info map[string]string) (io.ReadCloser, error) {
	return nil, nil
}

func (s McsStore) DeleteFile(ctx context.Context, info map[string]string) error {
	return nil
}
