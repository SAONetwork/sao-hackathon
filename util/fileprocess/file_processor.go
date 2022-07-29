package fileprocess

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	logging "github.com/ipfs/go-log/v2"
	"io"
	"io/ioutil"
	"math"
	"os"
	"sao-datastore-storage/util"
	"sao-datastore-storage/util/transport"
	"sao-datastore-storage/util/transport/types"
	"time"
)

var log = logging.Logger("file")
const FileChunkThreshold = 512 * (1 << 20) // 500 MB, change this to your requirement

type SplitFileInfo struct {
	FilePath string
	Offset   int64
	Size     int64
}

func SplitFile(file *os.File, fileSize int64, fileBasePath string) ([]SplitFileInfo, error) {
	totalPartsNum := 2
	fileChunkFloat := float64(fileSize) / float64(totalPartsNum)
	fileChunk := int64(math.Ceil(fileChunkFloat))

	log.Infof("Splitting to %d chunks.", totalPartsNum)

	var splitFileInfos []SplitFileInfo
	var offset int64
	for i := 0; i < totalPartsNum; i++ {
		partSize := int(math.Min(float64(fileChunk), float64(fileSize-int64(i)*fileChunk)))
		filePath := fmt.Sprintf("%s_%d", fileBasePath, i)
		err := writeThisPartToFile(file, partSize, filePath)
		if err != nil {
			return splitFileInfos, err
		}

		log.Info("Split to:", filePath)
		splitFileInfos = append(splitFileInfos, SplitFileInfo{
			FilePath: filePath,
			Offset:   offset,
			Size:     int64(partSize),
		})
		offset += int64(partSize)
	}
	return splitFileInfos, nil
}

func writeThisPartToFile(file *os.File, originalPartSize int, filePath string) error {
	if originalPartSize > FileChunkThreshold {
		// calculate total number of parts the file will be chunked into
		totalPartsNum := uint64(math.Ceil(float64(originalPartSize) / float64(FileChunkThreshold)))

		log.Infof("Splitting to %d pieces.\n", totalPartsNum)
		partFile, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer partFile.Close()

		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			return err
		}
		defer f.Close()
		for i := uint64(0); i < totalPartsNum; i++ {
			partSize := int(math.Min(FileChunkThreshold, float64(int64(originalPartSize)-int64(i*FileChunkThreshold))))
			partBuffer := make([]byte, partSize)

			_, err = file.Read(partBuffer)
			if err != nil {
				return err
			}

			// write/save buffer to disk
			n, err := f.Write(partBuffer)
			if err != nil {
				return err
			}
			f.Sync()
			partBuffer = nil
			log.Debug("Written ", n, " bytes")
		}
	} else {
		partBuffer := make([]byte, originalPartSize)

		_, err := file.Read(partBuffer)
		if err != nil {
			return err
		}

		// write to disk
		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		func() {
			defer f.Close()
		}()

		// write/save buffer to disk
		err = ioutil.WriteFile(filePath, partBuffer, os.ModeAppend)
		if err != nil {
			return err
		}
	}
	return nil
}

func CombineFile(chunkPaths []string, outFilePath string) error {
	f, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	outFile, err := os.OpenFile(outFilePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer outFile.Close()
	for j, currentChunkPath := range chunkPaths {
		err = combineChunk(outFilePath, currentChunkPath, outFile, j)
		if err != nil {
			return err
		}

		deleteChunkFile(currentChunkPath)
	}
	return nil
}

func combineChunk(outFilePath string, currentChunkPath string, outFile *os.File, index int) error {
	//read a chunk
	currentChunkFile, err := os.Open(currentChunkPath)
	if err != nil {
		return err
	}
	defer currentChunkFile.Close()

	chunkInfo, err := currentChunkFile.Stat()
	if err != nil {
		return err
	}

	var chunkSize = chunkInfo.Size()
	err = writeChunkToCombinedFile(chunkSize, currentChunkFile, outFile)
	if err != nil {
		return err
	}
	log.Debug("Recombining part [", index, "] into : ", outFilePath)
	return nil
}

func writeChunkToCombinedFile(chunkSize int64, currentChunkFile *os.File, outFile *os.File) error {
	if chunkSize > FileChunkThreshold {
		// calculate total number of parts the file will be chunked into
		totalPartsNum := uint64(math.Ceil(float64(chunkSize) / float64(FileChunkThreshold)))

		log.Infof("Splitting to %d pieces.\n", totalPartsNum)
		for i := uint64(0); i < totalPartsNum; i++ {
			partSize := int(math.Min(FileChunkThreshold, float64(chunkSize-int64(i*FileChunkThreshold))))
			chunkBufferBytes := make([]byte, partSize)

			_, err := currentChunkFile.Read(chunkBufferBytes)
			if err != nil {
				return err
			}

			n, err := outFile.Write(chunkBufferBytes)
			if err != nil {
				return err
			}

			outFile.Sync()
			chunkBufferBytes = nil

			log.Debug("Written ", n, " bytes")
		}
	} else {
		chunkBufferBytes := make([]byte, chunkSize)

		reader := bufio.NewReader(currentChunkFile)
		_, err := reader.Read(chunkBufferBytes)
		if err != nil {
			return err
		}

		n, err := outFile.Write(chunkBufferBytes)
		if err != nil {
			return err
		}

		outFile.Sync()
		chunkBufferBytes = nil

		log.Debug("Written ", n, " bytes")
	}
	return nil

}

func deleteChunkFile(currentChunkPath string) {
	err := os.Remove(currentChunkPath)
	if err != nil {
		log.Error(err)
	}
}

func TransferFile(ctx context.Context, transport transport.Transport, transfer types.Transfer, fileId string, outFilePath string) error {
	err := util.CreateFileIfNotExists(outFilePath)
	if err != nil {
		return err
	}

	log.Debugf("start transferring %s->%s", fileId, outFilePath)
	st := time.Now()
	handler, err := transport.Execute(ctx, transfer.Params, &types.TransportFileInfo{
		OutputFile: outFilePath,
		FileId:     fileId,
		FileSize:   int64(transfer.Size),
	})
	if err != nil {
		return err
	}

	// wait for data-transfer to finish
	if err = waitForTransferFinish(ctx, handler, int64(transfer.Size), fileId); err != nil {
		// Note that the data transfer has automatic retries built in, so if
		// it fails, it means it's already retried several times and we should
		// surface the problem to the user so they can decide manually whether
		// to keep retrying
		return err
	}
	log.Infof("file %s data-transfer completed successfully. time taken: %s", fileId, time.Since(st).String())
	return nil
}

func waitForTransferFinish(ctx context.Context, handler transport.Handler, size int64, fileId string) error {
	defer handler.Close()
	var lastOutputPct int64

	logTransferProgress := func(received int64) {
		pct := (100 * received) / size
		outputPct := pct / 10
		if outputPct != lastOutputPct {
			lastOutputPct = outputPct
			log.Infow(fileId, "transfer progress", "bytes received", received,
				"file size", size, "percent complete", pct)
		}
	}

	for {
		select {
		case evt, ok := <-handler.Sub():
			if !ok {
				return nil
			}
			if evt.Error != nil {
				return evt.Error
			}
			//deal.NBytesReceived = evt.NBytesReceived
			logTransferProgress(evt.NBytesReceived)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func DecryptFile(inputPath string, outPath string, key []byte) (int, error) {
	body, err := ioutil.ReadFile(inputPath)

	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, err
	}

	nonce := body[:gcm.NonceSize()]
	body = body[gcm.NonceSize():]
	originalData, err := gcm.Open(nil, nonce, body, nil)
	if err != nil {
		return 0, err
	}

	err = ioutil.WriteFile(outPath, originalData, 0777)
	if err != nil {
		return 0, err
	}

	return len(originalData), nil
}

func EncryptFile(inputPath string, outPath string) ([]byte, []byte, int, error) {
	key := make([]byte, 32)

	if _, err := rand.Read(key); err != nil {
		return nil, nil, 0, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, 0, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, 0, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, 0, err
	}

	body, err := ioutil.ReadFile(inputPath)

	encryptedData := gcm.Seal(nonce, nonce, body, nil)

	err = ioutil.WriteFile(outPath, encryptedData, 0777)
	if err != nil {
		return nil, nil, 0, err
	}

	return key, nonce, len(encryptedData), nil
}
