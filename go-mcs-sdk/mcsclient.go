package go_mcs_sdk

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"mime/multipart"
	"net/http"
	"sao-datastore-storage/web3"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
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
	Address         common.Address
	PrivateKey      *ecdsa.PrivateKey
	USDC            abi.ABI
	Payment         abi.ABI
	Provider        *web3.Provider
	ParamData       *ParamData
}

func NewMscClient(url string) *McsClient {
	provider, _ := web3.NewProvider(url)
	s := McsClient{
		McsEndpoint:     "https://mcs-api.filswan.com/api/v1",
		StorageEndpoint: "https://api.filswan.com",
		Provider:        provider,
	}
	s.USDC, _ = abi.JSON(strings.NewReader(ERC20_ABI))
	s.Payment, _ = abi.JSON(strings.NewReader(SWAN_PAYMENT_ABI))
	s.ParamData, _ = s.getParams()
	return &s
}

func (s *McsClient) SetAccount(privateKey string) (err error) {
	s.PrivateKey, err = crypto.HexToECDSA(privateKey)
	if err != nil {
		return err
	}
	s.Address = crypto.PubkeyToAddress(s.PrivateKey.PublicKey)
	return nil
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
	if err = writer.WriteField("wallet_address", s.Address.Hex()); err != nil {
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
	amount, err := s.getAverageAmount(s.Address.Hex(), size, duration)
	if err != nil {
		return "", err
	}
	payAmount, _ := new(big.Float).SetString(amount)
	var minPayment *big.Int
	minPayment, _ = new(big.Float).Mul(payAmount, big.NewFloat(1000000000000000000.0)).Int(minPayment)
	fmt.Println(minPayment, amount)
	allowance := s.queryAllowance()
	if allowance.Cmp(minPayment) <= 0 {
		s.approve(new(big.Int).Mul(minPayment, big.NewInt(10)))
	}
	tx, err := s.lockToken(wCid, minPayment, size)
	if err != nil {
		return "", err
	}
	return tx.Hex(), nil
}

func (s *McsClient) queryAllowance() *big.Int {
	result, err := s.Provider.Call(common.HexToAddress(s.ParamData.UsdcAddress), s.USDC.Methods["allowance"], []interface{}{s.Address, common.HexToAddress(s.ParamData.PaymentContractAddress)}, nil)
	if err != nil {
		return nil
	}
	return new(big.Int).SetBytes(result)
}

type Payment struct {
	Id         string
	MinPayment *big.Int
	Amount     *big.Int
	LockTime   *big.Int
	Recipient  common.Address
	Size       *big.Int
	CopyLimit  uint8
}

func (s *McsClient) CallContract(to common.Address, method abi.Method, params []interface{}) (*common.Hash, error) {
	nonce, _ := s.Provider.GetNonce(s.Address)
	gasPrice, _ := s.Provider.Getgasprice()
	var buf bytes.Buffer
	buf.Write(method.ID)
	payload, err := method.Inputs.Pack(params...)
	if err != nil {
		return nil, err
	}
	buf.Write(payload)
	/*
		gasLimit, err := s.Provider.EstamateGas(s.Address, to, payload)
		if err != nil {
			return nil, err
		}*/
	gasLimit := uint64(500000)
	tx := types.NewTransaction(nonce, to, big.NewInt(0), gasLimit, gasPrice, buf.Bytes())
	signed, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(80001)), s.PrivateKey)
	if err != nil {
		return nil, err
	}
	txhash := signed.Hash()
	return &txhash, s.Provider.SendTx(signed)
}

func (s *McsClient) approve(amount *big.Int) (*common.Hash, error) {
	approveMethod := s.USDC.Methods["approve"]
	params := []interface{}{common.HexToAddress(s.ParamData.PaymentContractAddress), amount}
	return s.CallContract(common.HexToAddress(s.ParamData.UsdcAddress), approveMethod, params)
}

func (s *McsClient) lockToken(cid string, min_amount *big.Int, size int) (*common.Hash, error) {

	paymentMethod := s.Payment.Methods["lockTokenPayment"]
	var amount *big.Int
	amount, _ = new(big.Float).Mul(new(big.Float).SetInt(min_amount), big.NewFloat(s.ParamData.PayMultiplyFactor)).Int(amount)
	payment := Payment{
		Id:         cid,
		MinPayment: min_amount,
		Amount:     amount,
		LockTime:   big.NewInt(int64(86400 * s.ParamData.LockTime)),
		Recipient:  common.HexToAddress(s.ParamData.PaymentRecipientAddress),
		Size:       big.NewInt(int64(size)),
		CopyLimit:  5,
	}
	return s.CallContract(common.HexToAddress(s.ParamData.PaymentContractAddress), paymentMethod, []interface{}{payment})
}
