package monitor

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

func (m *Monitor) Finish(orderId int64) error {
	fullpath := fmt.Sprintf("m/44'/60'/0'/0/%d", 0)
	path := hdwallet.MustParseDerivationPath(fullpath)
	account, _ := m.Wallet.Derive(path, true)
	fmt.Println(account.Address.Hex())
	nonce, _ := m.provider.GetNonce(account.Address)
	fmt.Println(nonce)
	gas_price, _ := m.provider.Getgasprice()

	to := common.HexToAddress(m.cfg.Contract)
	var buf bytes.Buffer
	buf.Write(m.contract.Methods["finish"].ID)
	paramId, _ := hex.DecodeString(fmt.Sprintf("%064x", orderId))
	buf.Write(paramId)
	tx := types.NewTransaction(nonce, to, big.NewInt(0), big.NewInt(1000000).Uint64(), gas_price, buf.Bytes())
	signedTx, signErr := m.Wallet.SignTx(account, tx, nil)
	if signErr != nil {
		log.Panicln("signer with signature error:", signErr)
		return signErr
	}
	fmt.Println(signedTx.Hash().String())
	return m.provider.SendTx(signedTx)
}
