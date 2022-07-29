package monitor

import (
	"context"
	"fmt"
	"math/big"
	"sao-datastore-storage/common"
	"sao-datastore-storage/model"
	"sao-datastore-storage/web3"
	"strings"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/go-co-op/gocron"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gwaylib/log"
)

type Monitor struct {
	cfg      common.MonitorInfo
	provider *web3.Provider
	Model    *model.Model
	Wallet   *hdwallet.Wallet
	contract abi.ABI
}

func NewMonitor(cfg common.MonitorInfo, model *model.Model) (*Monitor, error) {
	provider, _ := web3.NewProvider(cfg.Provider)
	wallet, err := hdwallet.NewFromMnemonic(cfg.Mnemonic)
	if err != nil {
		fmt.Println("invalid key")
		return nil, err
	}
	monitor := Monitor{
		provider: provider,
		Model:    model,
		cfg:      cfg,
		Wallet:   wallet,
	}
	return &monitor, nil
}

func (m *Monitor) Run() {
	ch := make(chan types.Log, 100)
	addresses := make([]ethcommon.Address, 0)
	addr := ethcommon.HexToAddress(m.cfg.Contract)
	addresses = append(addresses, addr)
	contract, _ := abi.JSON(strings.NewReader(web3.NFTABI))
	m.contract = contract

	latest := m.provider.GetLatestBlock()

	done := make(chan int, 1)
	fmt.Println("filter logs")
	logs := m.provider.FilterLogs(context.Background(), addresses, new(big.Int).SetInt64(m.cfg.BlockNumber), new(big.Int).SetUint64(latest))
	go func(ch chan types.Log, logs []types.Log) {
		for _, log := range logs {
			ch <- log
		}
	}(ch, logs)

	fmt.Println("listen download status")
	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Seconds().Do(func() {
		finishPurchases := m.Model.GetFinishPurchaseOrders()
		for _, p := range *finishPurchases {
			m.Finish(int64(p.Id))
			if err := m.Model.UpdatePurchaseOrderState(p.Id, model.FinishContractStarted); err != nil {
				log.Error(err)
			}
		}
	})

	s.StartAsync()


	fmt.Println("listen event")
	go m.provider.ListenEvent(addresses, ch, new(big.Int).SetInt64(m.cfg.BlockNumber), done)
	for {
		logger := <-ch
		event := logger.Topics[0].Hex()
		var buyer ethcommon.Address
		var tokenId *big.Int
		var price *big.Int
		var eventName string
		var timestamp int64
		var fileId int64
		var orderId *big.Int
		switch event {
		case "0xe96fc9a98f9b2e37b107acde0c3ba4ba52a1fb575da1acd1f15f0b6b1a35b8a9":
			tokenId, fileId, price, timestamp = m.parseListEvent(contract, logger)
			eventName = "ListingNFT"
		case "0xaf6c8baebc55f540953b129965caf1e3f48010ed731c2cea5ed61e71ebc57153":
			tokenId, buyer, orderId, price, timestamp = m.parseBuyEvent(contract, logger)
			eventName = "BuyNFT"
		case "0xfc677c6d4fb41b52076143bd16b0f032d45adbe706f64c08956a1829cd048986":
			orderId, timestamp = m.parseDownloadEvent(contract, logger)
			eventName = "DownloadFile"
		}

		switch eventName {
		case "ListingNFT":
			// set nft/file price
			if err := m.Model.UpdatePreviewPriceAndTokenId(fileId, price, tokenId); err != nil {
				log.Error(err)
				continue
			}
			fmt.Println(tokenId, fileId, price, timestamp)
		case "BuyNFT":
			filePreview, err := m.Model.GetFilePreviewByTokenId(tokenId.Int64())
			if err != nil {
				log.Error(err)
				continue
			}
			purchaseOrder := model.PurchaseOrder{
				Id:        uint(orderId.Int64()),
				FileId:    int64(filePreview.Id),
				BuyerAddr: buyer.Hex(),
				//OrderTxHash: orderId,
				State: model.ContractOrdered,
			}
			if err = m.Model.CreatePurchaseOrder(&purchaseOrder); err != nil {
				log.Error(err)
			}
			fmt.Println(tokenId, buyer, orderId, price, timestamp)
		case "DownloadFile":
			if err := m.Model.UpdatePurchaseOrderState(uint(orderId.Int64()), model.Finish); err != nil {
				log.Error(err)
			}
			fmt.Println("DownloadFile", orderId, timestamp)
		}
	}
}

func (m *Monitor) parseBuyEvent(contractABI abi.ABI, log types.Log) (*big.Int, ethcommon.Address, *big.Int, *big.Int, int64) {
	tokenId := new(big.Int)
	orderId := new(big.Int)
	var buyer ethcommon.Address
	tokenId.SetBytes(log.Topics[1].Bytes())
	buyer = ethcommon.BytesToAddress(log.Topics[2].Bytes())
	orderId.SetBytes(log.Topics[3].Bytes())
	var data []interface{}
	data, _ = contractABI.Unpack("Bought", log.Data)
	return tokenId, buyer, orderId, data[0].(*big.Int), data[1].(*big.Int).Int64()
}

func (m *Monitor) parseListEvent(contractABI abi.ABI, log types.Log) (*big.Int, int64, *big.Int, int64) {
	tokenId := new(big.Int)
	fmt.Println(log.Topics[1])
	tokenId.SetBytes(log.Topics[1].Bytes())
	var data []interface{}
	data, _ = contractABI.Unpack("Listing", log.Data)
	return tokenId, data[0].(*big.Int).Int64(), data[1].(*big.Int), data[2].(*big.Int).Int64()
}

func (m *Monitor) parseDownloadEvent(contractABI abi.ABI, log types.Log) (*big.Int, int64) {
	orderId := new(big.Int)
	orderId.SetBytes(log.Topics[1].Bytes())
	var data []interface{}
	data, _ = contractABI.Unpack("Download", log.Data)
	return orderId, data[0].(*big.Int).Int64()
}

func (m *Monitor) parseTransferEvent(contractABI abi.ABI, log types.Log) (ethcommon.Address, ethcommon.Address, *big.Int) {
	value := new(big.Int)
	// value indexed
	from := ethcommon.BytesToAddress(log.Topics[1].Bytes())
	to := ethcommon.BytesToAddress(log.Topics[2].Bytes())
	if len(log.Topics) == 4 {
		value.SetBytes(log.Topics[3].Bytes())
	} else if len(log.Topics) == 3 {
		var data []interface{}
		data, _ = contractABI.Unpack("Transfer", log.Data)
		if len(data) > 0 {
			value = data[0].(*big.Int)
		}
	}
	return from, to, value
}
