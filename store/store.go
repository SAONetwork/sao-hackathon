package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sao-datastore-storage/common"
	"sao-datastore-storage/model"
	"sao-datastore-storage/node"
	"sao-datastore-storage/proc"
	"sao-datastore-storage/util"
	"sao-datastore-storage/util/fileprocess"
	"sao-datastore-storage/util/transport"
	"sao-datastore-storage/util/transport/httptransport"
	"sao-datastore-storage/util/transport/types"
	"strings"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"golang.org/x/xerrors"
)

var log = logging.Logger("store")

const ORIGINAL_SUFFIX = ".original"
const DECRYPT_SUFFIX = ".decrypt"

// chain store interface.
// ipfs, filecoin, arweave should implement.
type Store interface {
	StoreFile(ctx context.Context, reader io.Reader, info map[string]string) (StoreRet, error)
	GetFile(ctx context.Context, info map[string]string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, info map[string]string) error
}

type FileInfoInMarket struct {
	ID           uint
	CreatedAt    time.Time
	UpdatedAt    time.Time
	EthAddr      string
	Preview      string
	Labels       string
	Price        float64
	Title        string
	Description  string
	ContentType  string
	FileCategory model.FileCategory
	Status       model.FilePreviewStatus
	NftTokenId   int64
	AlreadyPaid  bool
}

type PagedFileInfoInMarket struct {
	FileInfoInMarkets []FileInfoInMarket
	Total             int64
}

type StoreService struct {
	store     Store
	storeMap  map[string]Store
	m         *model.Model
	host      host.Host
	config    *common.Config
	repodir   string
	transport transport.Transport
}

type StoreRet struct {
	IpfsHash string
	McsInfo  *model.McsInfo
}

func NewStoreService(config *common.Config, m *model.Model, host host.Host, repodir string) (StoreService, error) {
	var store Store
	storeMap := make(map[string]Store)

	// ipfs
	if config.Ipfs.Ip != "" {
		ipfsUrl := fmt.Sprintf("%s:%d", config.Ipfs.Ip, config.Ipfs.Port)
		if config.Ipfs.ProjectId != "" {
			// infura
			store = NewIpfsStoreWithBasicAuth(ipfsUrl, config.Ipfs.ProjectId, config.Ipfs.ProjectSecret)
		} else {
			// local
			store = NewIpfsStore(ipfsUrl)
			storeMap["ipfs"] = store
		}
	}
	return StoreService{
		store:     store,
		storeMap: storeMap,
		m:         m,
		host:      host,
		config:    config,
		repodir:   repodir,
		transport: httptransport.New(host),
	}, nil
}

func (a StoreService) StoreFile(ctx context.Context, reader io.Reader, contentType string, size int64, dest string, duration int64, walletAddr string, filename string) (*model.FileInfo, error) {
	count, err := a.m.CountFileByFilenameAndStatus(dest, 0)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New(dest + " exists under ")
	}

	storeInfo := map[string]string{
		"address":  walletAddr,
		"filename": filename,
	}
	ret, err := a.store.StoreFile(ctx, reader, storeInfo)
	if err != nil {
		return nil, err
	}

	// TODO: allow user to upload two same files?
	expireAt := time.Now().UnixMilli()
	if duration < 0 {
		// -1 - forever
		expireAt = -1
	}

	file := model.FileInfo{
		Filename:    dest,
		ContentType: contentType,
		Size:        size,
		ExpireAt:    expireAt,
		Status:      0,
	}
	if ret.IpfsHash != "" {
		file.IpfsHash = ret.IpfsHash
	}
	returnFile, err := a.m.StoreFile(file, ret.McsInfo)
	if err != nil {
		return nil, err
	}

	return returnFile, nil
}

func (a StoreService) EncryptFileChunk(ctx context.Context, preview *model.FilePreview, splitFile fileprocess.SplitFileInfo, offset int64) (string, uint64, error) {
	peerIds := a.host.Peerstore().Peers()
	tryTime := 50
	var peerId peer.ID
	for findPeer := false; findPeer == false && tryTime > 0; tryTime-- {
		randomIndex := rand.Intn(len(peerIds))
		peerId = peerIds[randomIndex]
		for _, configPeer := range a.config.Libp2p.DirectPeers {
			if strings.Contains(configPeer, peerId.String()) {
				findPeer = true
				break
			}
		}
	}
	addrInfo := a.host.Peerstore().PeerInfo(peerId)

	transferParams := &types.HttpRequest{URL: a.config.ApiServer.ExposedPath + a.config.ApiServer.ContextPath + "/api/v1/proc/file/" + filepath.Base(splitFile.FilePath)}

	paramsBytes, err := json.Marshal(transferParams)
	if err != nil {
		return "", 0, xerrors.Errorf("marshalling request parameters: %v", err)
	}
	transfer := types.Transfer{
		Size:   uint64(splitFile.Size),
		Type:   "http",
		Params: paramsBytes,
	}

	s, err := a.host.NewStream(ctx, addrInfo.ID, proc.FileProcessEncryptionProtocolDraft)
	if err != nil {
		return "", transfer.Size, xerrors.Errorf("failed to open stream to peer %s: %w", addrInfo.ID, err)
	}
	defer s.Close()

	req := proc.FileEncryptReq{
		FileId:   fmt.Sprintf("%d", preview.Id),
		ClientId: preview.EthAddr,
		Offset:   uint64(offset),
		Size:     uint64(splitFile.Size),
		Transfer: transfer,
	}

	var resp proc.FileEncryptResp
	if err = util.DoRpc(ctx, s, &req, &resp, "json"); err != nil {
		return "", transfer.Size, xerrors.Errorf("send proposal rpc: %w", err)
	}

	if resp.Accepted {
		outFilePath := filepath.Join(node.StageProcPath(a.repodir), fmt.Sprintf("%s.encrypt", filepath.Base(splitFile.FilePath)))
		err = fileprocess.TransferFile(ctx, a.transport, resp.Transfer, req.FileId, outFilePath)
		if err != nil {
			return "", transfer.Size, xerrors.Errorf("transfer file error: %w", err)
		}
		return outFilePath, resp.Transfer.Size, nil
	} else {
		return "", transfer.Size, xerrors.Errorf("failed to get encrypted file chunk from peer: %w", err)
	}
}

func (a StoreService) GetFile(ctx context.Context, previewId uint, ethAddr string) (*model.FileInfo, io.ReadCloser, error) {
	filePreview, err := a.m.GetFilePreviewById(previewId)
	if err != nil {
		return nil, nil, err
	}

	file := a.m.GetFileInfoByPreviewId(previewId)
	ipfsHash := file.IpfsHash

	var storeService Store
	if file.McsInfoId > 0 {
		mcsInfo, err := a.m.GetMcsInfoById(file.McsInfoId)
		if err != nil {
			return file, nil, errors.New("missing ipfs hash")
		}
		ipfsHash = mcsInfo.IpfsUrl
		storeService = a.storeMap["mcs"]
	} else {
		storeService = a.storeMap["ipfs"]
	}

	filePath := filepath.Join(node.StageProcPath(a.repodir), filepath.Base(file.Filename))
	var originalFile *os.File
	if _, err = os.Stat(filePath + ORIGINAL_SUFFIX); errors.Is(err, os.ErrNotExist) {
		read, err := storeService.GetFile(ctx, map[string]string{
			"hash": ipfsHash,
		})
		if err != nil {
			return nil, nil, err
		}

		willDecrypt := filePreview.Price.Cmp(decimal.NewFromInt(0)) > 0
		if willDecrypt {
			defer read.Close()
			fileChunkMetadatas := a.m.GetFileChunkMetadatasByFileId(previewId)

			var splitFileInfos []fileprocess.SplitFileInfo

			for i, fileChunkMetadata := range fileChunkMetadatas {
				partBuffer := make([]byte, fileChunkMetadata.EncryptedSize)

				read.Read(partBuffer)

				// write to disk
				splitFilePath := fmt.Sprintf("%s_%d.encrypt", filePath, i)
				f, err := os.Create(splitFilePath)
				if err != nil {
					return nil, nil, err
				}

				// write/save buffer to disk
				err = ioutil.WriteFile(splitFilePath, partBuffer, os.ModeAppend)
				if err != nil {
					return nil, nil, err
				}

				func() {
					defer f.Close()
				}()

				log.Info("Split to:", splitFilePath)

				splitFileInfos = append(splitFileInfos, fileprocess.SplitFileInfo{
					FilePath: splitFilePath,
					Offset:   fileChunkMetadata.EncryptedOffset,
					Size:     fileChunkMetadata.EncryptedSize,
				})
			}

			var decryptFilePaths []string
			for i, splitFileInfo := range splitFileInfos {
				outFilePath := filepath.Join(node.StageProcPath(a.repodir), fmt.Sprintf("%s_%d%s", filepath.Base(filePath), i, DECRYPT_SUFFIX))
				decryptFilePath, err := a.decryptFileChunk(ctx, splitFileInfo, ethAddr, previewId, outFilePath)
				if err != nil {
					return nil, nil, err
				}
				decryptFilePaths = append(decryptFilePaths, decryptFilePath)
			}

			log.Infof("start combining decrypted chunks into %s", filePath+ORIGINAL_SUFFIX)
			if err = fileprocess.CombineFile(decryptFilePaths, filePath+ORIGINAL_SUFFIX); err != nil {
				return nil, nil, err
			}

			log.Infof("finished combining decrypted chunks into %s", filePath+ORIGINAL_SUFFIX)

			originalFile, _ = os.Open(filePath + ORIGINAL_SUFFIX)
			if err != nil {
				return nil, nil, err
			}

			return file, originalFile, err
		} else {
			return file, read, nil
		}
	} else {
		originalFile, _ = os.Open(filePath + ORIGINAL_SUFFIX)
		if err != nil {
			return nil, nil, err
		}
		return file, originalFile, err
	}
}

func (a StoreService) decryptFileChunk(ctx context.Context, splitFile fileprocess.SplitFileInfo, ethAddr string, fileId uint, outFilePath string) (string, error) {
	peerIds := a.host.Peerstore().Peers()
	for _, peerId := range peerIds {
		findPeer := false
		for _, configPeer := range a.config.Libp2p.DirectPeers {
			if strings.Contains(configPeer, peerId.String()) {
				findPeer = true
				break
			}
		}

		if !findPeer {
			log.Debugf("peer id %s is not in config file", peerId)
			continue
		}

		decryptedChunkPath, err, done := a.tryDecryptFromPeer(ctx, splitFile, ethAddr, fileId, outFilePath, peerId)
		if done {
			return decryptedChunkPath, err
		}
		if err != nil {
			log.Error(err)
		}
	}

	return "", errors.New("missing decrypt file part")
}

func (a StoreService) tryDecryptFromPeer(ctx context.Context, splitFile fileprocess.SplitFileInfo, ethAddr string, fileId uint, outFilePath string, peerId peer.ID) (string, error, bool) {
	addrInfo := a.host.Peerstore().PeerInfo(peerId)
	transferParams := &types.HttpRequest{URL: a.config.ApiServer.ExposedPath + a.config.ApiServer.ContextPath + "/api/v1/proc/file/" + filepath.Base(splitFile.FilePath)}

	paramsBytes, err := json.Marshal(transferParams)
	if err != nil {
		return "", xerrors.Errorf("marshalling request parameters: %v", err), true
	}
	transfer := types.Transfer{
		Size:   uint64(splitFile.Size),
		Type:   "http",
		Params: paramsBytes,
	}

	s, err := a.host.NewStream(ctx, addrInfo.ID, proc.FileProcessDecryptionProtocolDraft)
	if err != nil {
		return "", xerrors.Errorf("failed to open stream to peer %s: %w", addrInfo.ID, err), false
	}
	defer s.Close()

	req := proc.FileDecryptReq{
		FileId:   fmt.Sprintf("%d", fileId),
		ClientId: ethAddr,
		Offset:   uint64(splitFile.Offset),
		Size:     uint64(splitFile.Size),
		Transfer: transfer,
	}

	var resp proc.FileDecryptResp
	if err = util.DoRpc(ctx, s, &req, &resp, "json"); err != nil {
		return "", xerrors.Errorf("send proposal rpc: %w", err), false
	}

	if resp.Accepted {
		err = fileprocess.TransferFile(ctx, a.transport, resp.Transfer, req.FileId, outFilePath)
		if err != nil {
			return "", xerrors.Errorf("transfer file error: %w", err), false
		}
		return outFilePath, nil, true
	}
	return "", nil, false
}
