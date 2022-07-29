package proc

import (
	"bytes"
	"encoding/json"
	"io"
	"sao-datastore-storage/util/transport/types"

	cborutil "github.com/filecoin-project/go-cbor-util"
)

const ENCRYPT_SUFFIX = ".encrypt"
const DECRYPT_SUFFIX = ".decrypt"

type FileEncryptReq struct {
	// file unique id
	FileId string
	// file owner address
	ClientId string
	// process file start offset
	Offset uint64
	// process file total bytes
	Size uint64
	// transfer method
	Transfer types.Transfer
}

func (f *FileEncryptReq) Unmarshal(r io.Reader, format string) (err error) {
	if format == "json" {
		buf := &bytes.Buffer{}
		buf.ReadFrom(r)
		err = json.Unmarshal(buf.Bytes(), f)
	} else {
		// TODO: CBOR marshal
	}
	return nil
}

func (f *FileEncryptReq) Marshal(w io.Writer, format string) error {
	if format == "json" {
		bytes, err := json.Marshal(f)
		if err != nil {
			return err
		}
		w.Write(bytes)
		return nil
	} else {
		// TODO: CBOR marshal
	}
	return nil
}

type FileEncryptResp struct {
	FileKey  string
	Transfer types.Transfer
	Accepted bool
}

func (f *FileEncryptResp) Marshal(w io.Writer, format string) error {
	if format == "cbor" {
		err := cborutil.WriteCborRPC(w, f)
		return err
	} else {
		bytes, err := json.Marshal(f)
		if err != nil {
			return err
		}
		w.Write(bytes)
		return nil
	}
}
func (f *FileEncryptResp) Unmarshal(r io.Reader, format string) (err error) {
	if format == "json" {
		buf := &bytes.Buffer{}
		buf.ReadFrom(r)
		err = json.Unmarshal(buf.Bytes(), f)
	} else {
		// TODO: CBOR marshal
	}
	return nil
}

type FileDecryptReq struct {
	// file unique id
	FileId string
	// file owner address
	ClientId string
	// process file start offset
	Offset uint64
	// process file total bytes
	Size uint64
	// transfer method
	Transfer types.Transfer
}

func (f *FileDecryptReq) Unmarshal(r io.Reader, format string) (err error) {
	if format == "json" {
		buf := &bytes.Buffer{}
		buf.ReadFrom(r)
		err = json.Unmarshal(buf.Bytes(), f)
	} else {
		// TODO: CBOR marshal
	}
	return nil
}

func (f *FileDecryptReq) Marshal(w io.Writer, format string) error {
	if format == "json" {
		bytes, err := json.Marshal(f)
		if err != nil {
			return err
		}
		w.Write(bytes)
		return nil
	} else {
		// TODO: CBOR marshal
	}
	return nil
}

type FileDecryptResp struct {
	FileId   string
	Offset   uint64
	Size     uint64
	Transfer types.Transfer
	Accepted bool
}

func (f *FileDecryptResp) Marshal(w io.Writer, format string) error {
	if format == "cbor" {
		err := cborutil.WriteCborRPC(w, f)
		return err
	} else {
		bytes, err := json.Marshal(f)
		if err != nil {
			return err
		}
		w.Write(bytes)
		return nil
	}
}
func (f *FileDecryptResp) Unmarshal(r io.Reader, format string) (err error) {
	if format == "json" {
		buf := &bytes.Buffer{}
		buf.ReadFrom(r)
		err = json.Unmarshal(buf.Bytes(), f)
	} else {
		// TODO: CBOR marshal
	}
	return nil
}
