package util

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/fogleman/gg"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/nfnt/resize"
)

func ToJson(obj interface{}) (string, error) {
	resJson, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshalling json: %w", err)
	}

	return string(resJson), nil
}

func CreateFileIfNotExists(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

type CommonUnmarshaler interface {
	Unmarshal(io.Reader, string) error
}

type CommonMarshaler interface {
	Marshal(io.Writer, string) error
}

func DetectReaderType(reader io.Reader) (string, error) {
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	return http.DetectContentType(buffer), nil
}

func GenerateTags(contentType string, tempFileName string) (string, error) {
	if contentType == "image/png" || contentType == "image/jpeg" || contentType == "video/mp4" {
		ctx := context.Background()
		client, err := vision.NewImageAnnotatorClient(ctx)
		if err != nil {
			return "", err
		}
		defer client.Close()

		tempFile, err := os.Open(tempFileName)
		if err != nil {
			return "", err
		}

		image, err := vision.NewImageFromReader(tempFile)
		if err != nil {
			return "", err
		}

		labels, err := client.DetectLabels(ctx, image, nil, 10)
		if err != nil {
			return "", err
		}

		labelContents := make([]string, 4)
		labelIndex := 0
		for _, label := range labels {
			sp := strings.Split(label.Description, " ")
			if len(sp) >= 2 {
				continue
			}
			labelContents[labelIndex] = label.Description
			labelIndex++
			if labelIndex >= 4 {
				break
			}
		}
		return strings.Join(labelContents, ","), nil
	}
	return "", nil
}

func DoRpc(ctx context.Context, s network.Stream, req interface{}, resp interface{}, format string) error {
	errc := make(chan error)
	go func() {
		if m, ok := req.(CommonMarshaler); ok {
			if err := m.Marshal(s, format); err != nil {
				errc <- fmt.Errorf("failed to send request: %w", err)
				return
			}
		} else {
			errc <- fmt.Errorf("failed to send request")
			return
		}

		if m, ok := resp.(CommonUnmarshaler); ok {
			if err := m.Unmarshal(s, format); err != nil {
				errc <- fmt.Errorf("failed to read response: %w", err)
				return
			}
		} else {
			errc <- fmt.Errorf("failed to read response")
			return
		}

		errc <- nil
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func GenerateImgPreview(contentType string, tempFileName string) (string, string, error) {
	if contentType == "image/png" || contentType == "image/jpeg" {
		return GenerateImgFromImgFile(contentType, tempFileName)
	} else if contentType == "video/mp4" {
		previewFileName := fmt.Sprintf("%s.jpg", tempFileName)
		cmd := exec.Command("ffmpeg", "-i", tempFileName, "-vframes", "1", "-f", "image2", previewFileName)
		var buffer bytes.Buffer
		cmd.Stdout = &buffer
		if cmd.Run() != nil {
			return "", tempFileName, errors.New("could not generate frame")
		}
		return GenerateImgFromImgFile("image/jpeg", previewFileName)
	}
	return "", tempFileName, nil
}

func GenerateImgFromImgFile(contentType string, tempFileName string) (string, string, error) {
	// decode jpeg into image.Image
	var srcImage image.Image
	var err error
	var buf bytes.Buffer
	if contentType == "image/png" {
		srcImage, err = gg.LoadPNG(tempFileName)
	} else {
		srcImage, err = gg.LoadJPG(tempFileName)
	}
	if err != nil {
		return "",tempFileName, err
	}
	srcImage = resize.Thumbnail(256, 256, srcImage, resize.Lanczos3)
	dc := gg.NewContextForImage(srcImage)
	dc.EncodePNG(&buf)
	data := buf.Bytes()
	return base64.StdEncoding.EncodeToString(data), tempFileName, nil
}

