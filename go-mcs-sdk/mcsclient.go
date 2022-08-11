package go_mcs_sdk

import (
	"bytes"
	"encoding/json"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

type UploadResp struct {
	Status  string     `json:"status"`
	Data    UploadData `json:"data"`
	Message string     `json:"message"`
}

type UploadData struct {
	SourceFileUploadId int64  `json:"source_file_upload_id"`
	PayloadCid         string `json:"payload_cid"`
	IpfsUrl            string `json:"ipfs_url"`
	FileSize           int64  `json:"file_size"`
	WCid               string `json:"w_cid"`
}

type StatsResp struct {
	Data StatsData `json:"data"`
}

type BillingResp struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Data   int    `json:"data"`
}

type StatsData struct {
	AverageCostPushMessage           string
	AverageDataCostSealing1TB        string
	AverageGasCostSealing1TB         string
	AverageMinPieceSize              string
	AveragePricePerGBPerYear         string
	AverageVerifiedPricePerGBPerYear string
	Status                           string
}

type ParamResp struct {
	Status string    `json:"status"`
	Code   int       `json:"code"`
	Data   ParamData `json:"data"`
}
type ParamData struct {
	GasLimit                int     `json:"GAS_LIMIT"`
	LockTime                int     `json:"LOCK_TIME"`
	MintContractAddress     string  `json:"MINT_CONTRACT_ADDRESS"`
	PaymentContractAddress  string  `json:"PAYMENT_CONTRACT_ADDRESS"`
	PaymentRecipientAddress string  `json:"PAYMENT_RECIPIENT_ADDRESS"`
	PayMultiplyFactor       float64 `json:"PAY_MULTIPLY_FACTOR"`
	UsdcAddress             string  `json:"USDC_ADDRESS"`
}

type McsClient struct {
	McsEndpoint     string
	StorageEndpoint string
	Address         string
	PrivateKey      string
}

func (s McsClient) SetAccount(privateKey string) {
	s.PrivateKey = privateKey
	// TODO:
	//s.Address =
}

func (s McsClient) Upload(filename string, reader io.Reader, options map[string]string) (*UploadResp, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(fw, reader); err != nil {
		return nil, err
	}

	if err = writer.WriteField("duration", options["duration"]); err != nil {
		return nil, err
	}
	if err = writer.WriteField("file_type", options["fileType"]); err != nil {
		return nil, err
	}
	if err = writer.WriteField("wallet_address", s.Address); err != nil {
		return nil, err
	}

	if err = writer.Close(); err != nil {
		return nil, err
	}

	resp, err := http.Post(s.McsEndpoint+"/storage/ipfs/upload", writer.FormDataContentType(), body)
	if err != nil {
		return nil, err
	}
	resBody, err := ioutil.ReadAll(resp.Body)

	jsonResp := UploadResp{}
	if err = json.Unmarshal(resBody, &jsonResp); err != nil {
		return nil, err
	}

	if jsonResp.Status == "success" {
		return &jsonResp, nil
	} else {
		return nil, xerrors.New(jsonResp.Message)
	}
}

func (s McsClient) getParams() (*ParamData, error) {
	resp, err := http.Get(s.McsEndpoint + "/common/system/params")
	if err != nil {
		return nil, err
	}

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	jsonResp := ParamResp{}
	err = json.Unmarshal(resBody, &jsonResp)
	if err != nil {
		return nil, err
	}
	return &jsonResp.Data, nil
}

func (s McsClient) getAverageAmount(walletAddress string, fileSize int, duration int) (string, error) {
	fileSizeInGB := float64(fileSize) / math.Pow10(9)
	storageCostPerUnit := float64(0)

	storageRes, err := http.Get(s.StorageEndpoint + "/stats/storage?wallet_address=" + walletAddress)
	if err != nil {
		return "", err
	}

	resBody, err := ioutil.ReadAll(storageRes.Body)
	if err != nil {
		return "", err
	}

	jsonResp := StatsResp{}
	if err = json.Unmarshal(resBody, &jsonResp); err != nil {
		return "", err
	}

	var cost []string
	if jsonResp.Data.AveragePricePerGBPerYear != "" {
		cost = strings.Split(jsonResp.Data.AveragePricePerGBPerYear, " ")
	}
	if len(cost) > 0 {
		storageCostPerUnit, err = strconv.ParseFloat(cost[0], 64)
		if err != nil {
			return "", err
		}
	}

	// get FIL/USDC
	billingPrice := 1
	billingResp, err := http.Get(s.McsEndpoint + "/billing/price/filecoin?wallet_address=" + walletAddress)
	if err != nil {
		return "", err
	}

	resBody, err = ioutil.ReadAll(billingResp.Body)
	if err != nil {
		return "", err
	}

	billingJsonResp := BillingResp{}
	if err = json.Unmarshal(resBody, &billingJsonResp); err != nil {
		return "", err
	}

	billingPrice = billingJsonResp.Data

	price := decimal.NewFromFloat(fileSizeInGB * storageCostPerUnit * float64(duration*5*billingPrice) / 365)
	numberPrice := price.Truncate(9)
	if numberPrice.Cmp(decimal.Zero) > 0 {
		return price.Mul(decimal.NewFromInt(3)).StringFixed(9), nil
	} else {
		return "0.000000002", nil
	}
}

func (s McsClient) MakePayment(wCid string, size int, duration int) (string, error) {
	amount, err := s.getAverageAmount(s.Address, size, duration)
	if err != nil {
		return "", err
	}

	tx, err := s.lockToken(wCid, amount, size)
	if err != nil {
		return "", err
	}

	return tx, nil
}

func (s McsClient) lockToken(wCid string, amount string, size int) (string, error) {
	return "", nil
}
