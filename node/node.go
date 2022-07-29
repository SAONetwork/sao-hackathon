package node

import (
	"context"
	crand "crypto/rand"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/wallet"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"io/ioutil"
	"os"
	"path/filepath"
	"sao-datastore-storage/cmd"
	"sao-datastore-storage/common"
	"sao-datastore-storage/util/keystore"
)

var log = logging.Logger("node")

type Node struct {
	Host   host.Host
	Wallet *wallet.LocalWallet
}

func Setup(ctx context.Context, cfgdir string, libp2pConfig common.Libp2p) (*Node, error) {
	peerkey, err := loadOrInitPeerKey(keyPath(cfgdir))
	if err != nil {
		return nil, err
	}

	var listenAddressesOption = libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0")
	if len(libp2pConfig.ListenAddresses) > 0 {
		listenAddressesOption = libp2p.ListenAddrStrings(libp2pConfig.ListenAddresses...)
	}
	h, err := libp2p.New(
		listenAddressesOption,
		libp2p.Identity(peerkey),
	)
	if err != nil {
		return nil, err
	}

	if len(libp2pConfig.DirectPeers) > 0 {
		for _, addr := range libp2pConfig.DirectPeers {
			a, err := ma.NewMultiaddr(addr)
			if err != nil {
				return nil, err
			}

			pi, err := peer.AddrInfoFromP2pAddr(a)
			if err != nil {
				return nil, err
			}

			log.Info("p2p connecting ", addr)
			if err = h.Connect(ctx, *pi); err != nil {
				log.Error(err)
			}
		}
	}

	wallet, err := setupWallet(walletPath(cfgdir))
	if err != nil {
		return nil, err
	}

	prepareStageDir(StageProcPath(cfgdir))

	return &Node{
		Host:   h,
		Wallet: wallet,
	}, nil
}

func prepareStageDir(dir string) error {
	if _, err := os.Stat(dir); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return nil
}

func StageProcPath(baseDir string) string {
	return filepath.Join(baseDir, cmd.FsStaging, "proc")
}

func keyPath(baseDir string) string {
	return filepath.Join(baseDir, "libp2p.key")
}

func walletPath(baseDir string) string {
	return filepath.Join(baseDir, "wallet")
}

func loadOrInitPeerKey(kf string) (crypto.PrivKey, error) {
	data, err := ioutil.ReadFile(kf)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		k, _, err := crypto.GenerateEd25519Key(crand.Reader)
		if err != nil {
			return nil, err
		}

		data, err := crypto.MarshalPrivateKey(k)
		if err != nil {
			return nil, err
		}

		if err := ioutil.WriteFile(kf, data, 0600); err != nil {
			return nil, err
		}

		return k, nil
	}
	return crypto.UnmarshalPrivateKey(data)
}

func setupWallet(dir string) (*wallet.LocalWallet, error) {
	kstore, err := keystore.OpenOrInitKeystore(dir)
	if err != nil {
		return nil, err
	}

	wallet, err := wallet.NewWallet(kstore)
	if err != nil {
		return nil, err
	}

	addrs, err := wallet.WalletList(context.TODO())
	if err != nil {
		return nil, err
	}

	if len(addrs) == 0 {
		_, err := wallet.WalletNew(context.TODO(), types.KTBLS)
		if err != nil {
			return nil, err
		}
	}

	return wallet, nil
}
