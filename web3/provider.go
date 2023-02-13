package web3

import (
	"context"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"math/big"
)

type Provider struct {
	client  *ethclient.Client
	gclient *gethclient.Client
	URL     string
}

func NewProvider(url string) (*Provider, error) {
	ctx := context.Background()
	c, err := rpc.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	client := ethclient.NewClient(c)
	gclient := gethclient.New(c)
	return &Provider{
		client:  client,
		gclient: gclient,
		URL:     url,
	}, nil
}

func (p *Provider) ListenEvent(addresses []common.Address, event_chan chan types.Log, from *big.Int, done chan int) {

	//ctx := context.Background()
	//
	//query := ethereum.FilterQuery{
	//	Addresses: addresses,
	//	FromBlock: from,
	//}

	var channel = make(chan types.Log)
	//sub, err := p.client.SubscribeFilterLogs(ctx, query, channel)
	//if err != nil {
	//	log.Printf("failed to subscribe %s %s", err, p.URL)
	//	return
	//}

	for {
		latest := p.GetLatestBlock()
		logs := p.FilterLogs(context.Background(), addresses, from, new(big.Int).SetUint64(latest))
		go func(ch chan types.Log, logs []types.Log) {
			for _, log := range logs {
				ch <- log
			}
		}(channel, logs)

		select {
		case <-done:
			return
		//case err := <-sub.Err():
		//	log.Println(err, p.URL)
		//	os.Exit(0)
		//	break
		case log := <-channel:
			event_chan <- log
		}
	}

}

func (p *Provider) Call(to common.Address, method abi.Method, params []interface{}, block *big.Int) ([]byte, error) {
	ctx := context.Background()
	from := common.HexToAddress("0x0000000000000000000000000000000000000000")

	args, _ := method.Inputs.PackValues(params)
	msg := ethereum.CallMsg{
		From: from,
		To:   &to,
		Gas:  136484,
		Data: append(method.ID, args[:]...),
	}
	return p.client.CallContract(ctx, msg, block)
}

func (p *Provider) GetLatestBlock() uint64 {
	ctx := context.Background()
	latest, _ := p.client.BlockNumber(ctx)
	return latest
}

func (p *Provider) FilterLogs(ctx context.Context, addresses []common.Address, from *big.Int, to *big.Int) []types.Log {
	query := ethereum.FilterQuery{
		Addresses: addresses,
		FromBlock: from,
		ToBlock:   to,
	}

	logs, err := p.client.FilterLogs(ctx, query)
	if err != nil {
		log.Printf("failed to filter log from %d %d, err: %s, chainId: %s\n", from, to, err.Error(), p.URL)
	}
	return logs
}

func (p *Provider) SubscribePendingTransactions(ch chan common.Hash) {
	ctx := context.Background()
	p.gclient.SubscribePendingTransactions(ctx, ch)
}

func (p *Provider) SubscribeNewHead(ch chan *types.Header) {
	ctx := context.Background()
	p.client.SubscribeNewHead(ctx, ch)
}

func (p *Provider) SendTx(tx *types.Transaction) error {
	ctx := context.Background()
	return p.client.SendTransaction(ctx, tx)
}

func (p *Provider) GetNonce(address common.Address) (uint64, error) {
	ctx := context.Background()
	return p.client.PendingNonceAt(ctx, address)
}

func (p *Provider) EstamateGas(from, to common.Address, data []byte) (uint64, error) {
	ctx := context.Background()
	msg := ethereum.CallMsg{
		From: from,
		To:   &to,
		Data: data,
	}
	return p.client.EstimateGas(ctx, msg)
}

func (p *Provider) Getgasprice() (*big.Int, error) {
	ctx := context.Background()
	return p.client.SuggestGasPrice(ctx)
}
