package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sao-datastore-storage/cmd"
	"sao-datastore-storage/model"
	"sao-datastore-storage/proc"
	"sao-datastore-storage/util"
	"sao-datastore-storage/util/fileprocess"
	"strings"
	"time"
)

var fileCategories = map[string]model.FileCategory{
	"image/avif":      model.Image,
	"image/gif":       model.Image,
	"image/jpeg":      model.Image,
	"application/pdf": model.Document,
	"image/png":       model.Image,
	"image/svg+xml":   model.Image,
	"audio/mpeg":      model.Music,
	"audio/mp3":       model.Music,
	"video/mp4":       model.Video,
	"application/zip": model.Document,
}

var formatContentTypeMaps = map[string]string {
	"CSV" : "text/csv",
	"JPG":"image/jpeg",
	"SVG":"image/svg+xml",
	"MP3":"audio/mpeg",
	"MP4":"video/mp4",
}

var fileExtensionCategories = map[string]model.FileCategory{
	".aiff": model.Music,
	".mp3":  model.Music,
	".m3u":  model.Music,
	".wav":  model.Music,
	".wma":  model.Music,
	".bwf":  model.Music,
	".dat":  model.Music,

	".mpeg": model.Video,
	".wmv":  model.Video,
	".flv":  model.Video,
	".avi":  model.Video,
	".mp4":  model.Video,
	".asf":  model.Video,
	".divx": model.Video,

	//".jpeg": model.Image,
	//".jpg":  model.Image,
	//".gif":  model.Image,
	//".png":  model.Image,
	//".bmp":  model.Image,
	//".tiff": model.Image,
	//".svg":  model.Image,

	".pdf":  model.Document,
	".doc":  model.Document,
	".txt":  model.Document,
	".docx": model.Document,
	".ndoc": model.Document,
	".fodt": model.Document,
	".sub":  model.Document,
	".me":   model.Document,
	".rtf":  model.Document,
	".fdr":  model.Document,
	".zip":  model.Document,
	".xlsx":  model.Document,
	".csv":  model.Document,
}

func (s *Server) uploadFile(reader io.Reader, filename string, contentType string, ethAddress string, additionalInfo string) (*model.FileInfoInMarket, error) {
	tmpPath := filepath.Join(s.Repodir, cmd.FsStaging, ethAddress)
	var filInfo model.FileInfoInMarket

	// create dir
	err := os.MkdirAll(tmpPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile(tmpPath, uuid.New().String())
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)
	// return that we have successfully uploaded our file!
	log.Info("Successfully staged File\n")

	tempFileName := tempFile.Name()

	fileCategory := getFileCategory(filepath.Ext(filename))

	filePreview := model.FilePreview{
		EthAddr:      ethAddress,
		TmpPath:      tempFileName,
		ContentType:  contentType,
		Status:       model.Uploading,
		Filename:     filename,
		FileCategory: fileCategory,
		AdditionalInfo: additionalInfo,
	}

	preview, imgFilePath, err := util.GenerateImgPreview(contentType, tempFileName)
	if err != nil {
		log.Error(err)
	}
	filePreview.Preview = preview

	tags, err := util.GenerateTags(contentType, imgFilePath)
	if err != nil {
		log.Error(err)
	}

	if err = s.Model.CreateFilePreview(&filePreview); err != nil {
		return nil, xerrors.New("database error")
	}

	fileExtension := filepath.Ext(filePreview.Filename)
	if fileExtension != "" {
		fileExtension = fileExtension[1:]
	}

	filInfo = model.FileInfoInMarket{
		Id:             filePreview.Id,
		EthAddr:        filePreview.EthAddr,
		Preview:        filePreview.Preview,
		Price:          filePreview.Price,
		Labels:         tags,
		Title:          filePreview.Title,
		Description:    filePreview.Description,
		ContentType:    filePreview.ContentType,
		Type:           filePreview.Type,
		Status:         filePreview.Status,
		FileCategory:   fileCategory,
		AdditionalInfo: additionalInfo,
		AlreadyPaid:    false,
		FileExtension:  fileExtension}

	return &filInfo, nil
}

func (s *Server) StoreFileWithPreview(ctx context.Context, preview model.FilePreview, ethAddress string) (*model.FileInfoInMarket, error) {
	filePreview, err := s.Model.GetFilePreviewById(preview.Id)
	if err != nil {
		return nil, xerrors.New("database error")
	}
	dir := path.Dir(filePreview.TmpPath)
	os.MkdirAll(dir, 0666)

	willEncrypt := preview.Price.Cmp(decimal.NewFromInt(0))> 0

	// TODO
	if willEncrypt {
		go s.processFileSplitAndEncryption(filePreview)
	} else {
		go s.processFreeFile(ctx, filePreview)
	}

	updateMap := map[string]interface{}{
		"Labels":         preview.Labels,
		"Price":          preview.Price,
		"Preview":        preview.Preview,
		"Title":          preview.Title,
		"Description":    preview.Description,
		"Type":           preview.Type,
		"AdditionalInfo": preview.AdditionalInfo,
		"Status":         model.UploadSuccess,
	}
	if err = s.Model.UpdatePreview(preview.Id, updateMap); err != nil {
		return nil, xerrors.New("database error")
	}

	fileInfoInMarket := model.FileInfoInMarket{Id: filePreview.Id,
		CreatedAt:    filePreview.CreatedAt,
		UpdatedAt:    filePreview.UpdatedAt,
		EthAddr:      filePreview.EthAddr,
		Preview:      preview.Preview,
		Labels:       preview.Labels,
		Price:        preview.Price,
		Title:        preview.Title,
		Description:  preview.Description,
		ContentType:  filePreview.ContentType,
		Type:         preview.Type,
		Status:       model.UploadSuccess,
		NftTokenId:   filePreview.NftTokenId,
		FileCategory: filePreview.FileCategory,
		AdditionalInfo: filePreview.AdditionalInfo,
		AlreadyPaid:  false}

	return &fileInfoInMarket, nil
}

func (s *Server) processFileSplitAndEncryption(filePreview *model.FilePreview) {
	file, err := os.Open(filePreview.TmpPath)
	if err != nil {
		log.Error(err)
		return
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		log.Error(err)
		return
	}
	fileSize := fileStat.Size()

	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	// split file
	log.Infof("start split %s ...", filePreview.TmpPath)
	splitFileBasePath := filepath.Join(s.Repodir, cmd.FsStaging, "proc", fmt.Sprintf("%d_%s", filePreview.Id, filepath.Base(file.Name())))
	splitFileInfos, err := fileprocess.SplitFile(file, fileSize, splitFileBasePath)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("complete split file into %d chunks", len(splitFileInfos))

	// encrypt file chunks
	log.Infof("start encrypting chunks...")
	var encryptedFileChunkPaths []string
	var chunkMetadatas []model.FileChunkMetadata
	var encryptedOffset int64 = 0
	for _, splitFileInfo := range splitFileInfos {
		encryptFilePath, encryptedSize, err := s.StoreService.EncryptFileChunk(ctx, filePreview, splitFileInfo, encryptedOffset)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("chunk %s is encrypted into %s", splitFileInfo.FilePath, encryptFilePath)

		encryptedFileChunkPaths = append(encryptedFileChunkPaths, encryptFilePath)
		chunkMetadatas = append(chunkMetadatas, model.FileChunkMetadata{
			FileId:          filePreview.Id,
			Offset:          splitFileInfo.Offset,
			Size:            splitFileInfo.Size,
			EncryptedOffset: encryptedOffset,
			EncryptedSize:   int64(encryptedSize),
		})
		encryptedOffset += int64(encryptedSize)
	}
	log.Infof("complete encrypting chunks")

	// combine encrypted file chunks
	combinedEncryptedPath := filePreview.TmpPath + proc.ENCRYPT_SUFFIX
	log.Infof("start combining ecnrypted chunks into %s", combinedEncryptedPath)
	if err = fileprocess.CombineFile(encryptedFileChunkPaths, combinedEncryptedPath); err != nil {
		log.Error(err)
		return
	}
	log.Infow("complete combining.")

	//  start storing combined encrypted file.
	storeFile, err := os.Open(combinedEncryptedPath)
	if err != nil {
		log.Error(err)
		return
	}
	encryptFileStat, err := storeFile.Stat()
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("uploading to ipfs/filecoin...")
	duration := int64(-1)
	dsFile, err := s.StoreService.StoreFile(ctx, storeFile, filePreview.ContentType, encryptFileStat.Size(), file.Name(), duration)
	if err != nil {
		log.Error(err)
		return
	}

	err = s.Model.StoreFileMetadata(chunkMetadatas, dsFile.Id, int64(filePreview.Id))
	if err != nil {
		log.Error(err)
		return
	}

	defer func() {
		err = os.Remove(filePreview.TmpPath)
		if err != nil {
			log.Error(err)
		}
		err = os.Remove(encryptFileStat.Name())
		if err != nil {
			log.Error(err)
		}
	}()

	return
}

func (s *Server) processFreeFile(ctx context.Context, filePreview *model.FilePreview) {
	file, err := os.Open(filePreview.TmpPath)
	if err != nil {
		log.Error(err)
		return
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		log.Error(err)
		return
	}
	fileSize := fileStat.Size()

	log.Infof("uploading to ipfs/filecoin...")
	duration := int64(-1)
	dsFile, err := s.StoreService.StoreFile(ctx, file, filePreview.ContentType, fileSize, file.Name(), duration)
	if err != nil {
		log.Error(err)
		return
	}

	updateMap := map[string]interface{}{
		"Status": model.PlacedToIpfs,
		"FileId": dsFile.Id,
	}
	err = s.Model.UpdatePreviewLinkedWithIpfs(filePreview.Id, updateMap)
	if err != nil {
		log.Error(err)
		return
	}

	defer func() {
		err = os.Remove(filePreview.TmpPath)
		if err != nil {
			log.Error(err)
		}
	}()
}

func (s *Server) deleteUploaded(previewId uint, ethAddress string) error {
	filePreview, err := s.Model.GetFilePreviewById(previewId)
	if err != nil {
		return errors.New("get file failed")
	}
	if filePreview.EthAddr != ethAddress {
		return errors.New("invalid previewId")
	}

	if filePreview.Status != model.Uploading {
		return errors.New("already uploaded success, can't delete it")
	}

	err = s.Model.DeletePreview(filePreview)
	if err != nil {
		return errors.New("delete preview failed")
	}

	err = os.Remove(filePreview.TmpPath)
	if err != nil {
		log.Error(err)
	}
	return nil
}

func (s *Server) getFileInfos(ethAddress string, offset int, limit int, category string, format string, price int) *model.PagedFileInfoInMarket {
	contentType, _ := formatContentTypeMaps[strings.ToUpper(format)]
	condition := map[string]interface{}{}
	if category != "" {
		condition["file_category"] = category
	}
	if contentType != "" {
		condition["content_type"] = contentType
	}
	files, count := s.Model.GetMarketFiles(limit, offset, ethAddress, condition, price)
	return &model.PagedFileInfoInMarket{
		FileInfoInMarkets: files,
		Total:             count,
	}
}

func (s *Server) getFileInfo(fileId uint, ethAddress string) (*model.FileDetail, error) {
	files, err := s.Model.GetFileInfo(fileId, ethAddress)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (s *Server) checkFileStatus(fileId uint, ethAddress string) error {

	filePreview, err := s.Model.GetFilePreviewByFileId(fileId)
	if err != nil {
		return errors.New("file id not found in system")
	}
	if filePreview.Price.Cmp(decimal.NewFromInt(0))> 0 && filePreview.EthAddr != ethAddress {
		purchaseOrder := s.Model.GetPurchaseOrder(fileId, ethAddress)
		if purchaseOrder.Id == 0 {
			return errors.New("not purchased")
		}
		if purchaseOrder.State == model.ContractOrdered {
			if err := s.Model.UpdatePurchaseOrderState(purchaseOrder.Id, model.ReadyToDownload); err != nil {
				log.Error(err)
			}
			//return errors.New("already purchased, we are preparing the download file")
		}
	}
	return nil
}

func getFileCategory(ext string) model.FileCategory {
	fileCategory, exists := fileExtensionCategories[strings.ToLower(ext)]
	if !exists {
		fileCategory = model.Other
	}
	return fileCategory
}
