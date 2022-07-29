package store

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/fogleman/gg"
)

func TestBase64JPG(t *testing.T) {
	img, _ := gg.LoadJPG("./test.jpg")
	dc := gg.NewContextForImage(img)
	var buf bytes.Buffer
	dc.EncodePNG(&buf)
	data := buf.Bytes()
	fmt.Println(base64.StdEncoding.EncodeToString(data))
}

func TestBase64PNG(t *testing.T) {
	img, _ := gg.LoadPNG("./test.png")
	dc := gg.NewContextForImage(img)
	var buf bytes.Buffer
	dc.EncodePNG(&buf)
	data := buf.Bytes()
	fmt.Println(base64.StdEncoding.EncodeToString(data))
}
