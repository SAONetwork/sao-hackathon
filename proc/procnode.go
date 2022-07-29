package proc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/filecoin-project/lotus/chain/wallet"
	"github.com/google/uuid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"os"
	"path/filepath"
	"sao-datastore-storage/common"
	"sao-datastore-storage/model"
	"sao-datastore-storage/node"
	"sao-datastore-storage/util/fileprocess"
	"sao-datastore-storage/util/transport"
	"sao-datastore-storage/util/transport/httptransport"
	"sao-datastore-storage/util/transport/types"
	"time"
)

var log = logging.Logger("proc")

const FileProcessEncryptionProtocolDraft = "/sao/file/encrypt/0.0.1"
const FileProcessDecryptionProtocolDraft = "/sao/file/decrypt/0.0.1"
const processReadDeadline = 10 * time.Second
const processWriteDeadline = 5 * 60 * time.Second

// TODO replace database
type ProcNode struct {
	ctx       context.Context
	host      host.Host
	wallet    *wallet.LocalWallet
	model     *model.Model
	config    *common.Config
	transport transport.Transport
	repodir   string
}

func NewProcNode(ctx context.Context, host host.Host, wallet *wallet.LocalWallet, m *model.Model, config *common.Config, repodir string) *ProcNode {
	return &ProcNode{
		ctx:       ctx,
		host:      host,
		wallet:    wallet,
		model:     m,
		config:    config,
		transport: httptransport.New(host),
		repodir:   repodir,
	}
}

func (p *ProcNode) Start() {
	p.host.SetStreamHandler(FileProcessEncryptionProtocolDraft, p.handleFileEncryptionRequest)
	p.host.SetStreamHandler(FileProcessDecryptionProtocolDraft, p.handleFileDecryptionRequest)
	p.serverAPI()
}

func (p *ProcNode) Stop(ctx context.Context) error {
	p.host.RemoveStreamHandler(FileProcessEncryptionProtocolDraft)
	p.host.RemoveStreamHandler(FileProcessDecryptionProtocolDraft)
	return nil
}

func (p *ProcNode) handleFileDecryptionRequest(s network.Stream) {
	defer s.Close()

	// Set a deadline on reading from the stream so it doesn't hang
	_ = s.SetReadDeadline(time.Now().Add(processReadDeadline))
	defer s.SetReadDeadline(time.Time{}) // nolint

	var req FileDecryptReq
	err := req.Unmarshal(s, "json")
	if err != nil {
		log.Warnw("reading fileproc req from stream", "err", err)
		var resp = &FileDecryptResp{
			Accepted: false,
		}

		resp.Marshal(s, "json")
		return
	}

	// lookup in database if this node has encrypted key for this offset.
	var keyStore model.KeyStore

	condition := map[string]interface{}{"file_id": req.FileId, "size": req.Size}
	keyStore, err = p.model.GetKeyStore(condition, req.Offset)
	if err != nil {
		log.Warnw("no key store record found", "err", err)
		var resp = &FileDecryptResp{
			Accepted: false,
		}

		resp.Marshal(s, "json")
		return
	}

	tctx, cancel := context.WithDeadline(p.ctx, time.Now().Add(p.config.Transport.MaxTransferDuration*time.Second))
	defer cancel()

	outFilePath := filepath.Join(node.StageProcPath(p.repodir), fmt.Sprintf("%s_%d", req.FileId, req.Offset))

	if _, err = os.Stat(outFilePath + ENCRYPT_SUFFIX); errors.Is(err, os.ErrNotExist) {
		// path does not exist
		err = fileprocess.TransferFile(tctx, p.transport, req.Transfer, req.FileId, outFilePath+ENCRYPT_SUFFIX)
		if err != nil {
			log.Warnw("transfer file error", "err", err)
			var resp = &FileDecryptResp{
				Accepted: false,
			}

			resp.Marshal(s, "json")
			return
		}
	}

	var size int
	if decryptFileInfo, err := os.Stat(outFilePath); errors.Is(err, os.ErrNotExist) {
		// path does not exist
		size, err = fileprocess.DecryptFile(outFilePath+ENCRYPT_SUFFIX, outFilePath, keyStore.Key)
		if err != nil {
			log.Warnw("decryption error", "err", err)
			var resp = &FileDecryptResp{
				Accepted: false,
			}

			resp.Marshal(s, "json")
			return
		}
	} else {
		size = int(decryptFileInfo.Size())
	}

	url := fmt.Sprintf("%s%s/api/v1/proc/decrypt/%s", p.config.ApiServer.ExposedPath, p.config.ApiServer.ContextPath, filepath.Base(outFilePath))
	paramsBytes, err := json.Marshal(types.HttpRequest{
		URL: url,
	})

	var resp = &FileDecryptResp{
		Transfer: types.Transfer{
			Type:   "http",
			Size:   uint64(size),
			Params: paramsBytes,
		},
		Accepted: true,
	}

	err = resp.Marshal(s, "json")
	if err != nil {
		log.Warnw("decryption error", "err", err)
		var resp = &FileDecryptResp{
			Accepted: false,
		}

		resp.Marshal(s, "json")
		return
	}
}

func (p *ProcNode) handleFileEncryptionRequest(s network.Stream) {
	defer s.Close()

	// Set a deadline on reading from the stream so it doesn't hang
	_ = s.SetReadDeadline(time.Now().Add(processReadDeadline))
	defer s.SetReadDeadline(time.Time{}) // nolint

	// decode request
	var req FileEncryptReq
	err := req.Unmarshal(s, "json")
	if err != nil {
		log.Warnw("reading fileproc req from stream", "err", err)
		var resp = &FileEncryptResp{
			Accepted: false,
		}
		resp.Marshal(s, "json")
		return
	}
	log.Infow("received file encryption request", "id", req.FileId, "client-peer", s.Conn().RemotePeer())

	// transfer raw file chunk.
	tctx, cancel := context.WithDeadline(p.ctx, time.Now().Add(p.config.Transport.MaxTransferDuration*time.Second))
	defer cancel()

	outFilePath := filepath.Join(node.StageProcPath(p.repodir), fmt.Sprintf("%s_%d", req.FileId, req.Offset))
	err = fileprocess.TransferFile(tctx, p.transport, req.Transfer, req.FileId, outFilePath)
	if err != nil {
		log.Warnw("transfer file", "err", err)
		var resp = &FileEncryptResp{
			Accepted: false,
		}
		resp.Marshal(s, "json")
		return
	}

	key, nonce, encryptedSize, err := fileprocess.EncryptFile(outFilePath, outFilePath+ENCRYPT_SUFFIX)
	if err != nil {
		var resp = &FileEncryptResp{
			Accepted: false,
		}
		resp.Marshal(s, "json")
		return
	}

	keyId := uuid.New().String()
	keyStore := model.KeyStore{
		Id:     keyId,
		FileId: req.FileId,
		Offset: req.Offset,
		Size:   uint64(encryptedSize),
		Key:    key,
		Nonce:  nonce,
	}
	if err = p.model.CreateKeyStore(&keyStore); err != nil {
		log.Error(err)
		var resp = &FileEncryptResp{
			Accepted: false,
		}
		resp.Marshal(s, "json")
		return
	}

	url := fmt.Sprintf("%s%s/api/v1/proc/encrypt/%s", p.config.ApiServer.ExposedPath, p.config.ApiServer.ContextPath, fmt.Sprintf("%s_%d%s", req.FileId, req.Offset, ENCRYPT_SUFFIX))
	paramsBytes, err := json.Marshal(types.HttpRequest{
		URL: url,
	})
	if err != nil {
		log.Errorf("marshalling request parameters: %w", err)
		var resp = &FileEncryptResp{
			Accepted: false,
		}
		resp.Marshal(s, "json")
		return
	}

	var resp = &FileEncryptResp{
		FileKey: keyId,
		Transfer: types.Transfer{
			Type:   "http",
			Size:   uint64(encryptedSize),
			Params: paramsBytes,
		},
		Accepted: true,
	}

	// Set a deadline on writing to the stream so it doesn't hang
	_ = s.SetWriteDeadline(time.Now().Add(processWriteDeadline))
	defer s.SetWriteDeadline(time.Time{}) // nolint

	// Write the response to the client
	resp.Marshal(s, "json")
	if err != nil {
		var resp = &FileEncryptResp{
			Accepted: false,
		}
		resp.Marshal(s, "json")
		return
	}
}
