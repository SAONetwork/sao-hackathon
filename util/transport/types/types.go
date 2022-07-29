package types

import (
	"github.com/ipfs/go-cid"
)

// Transfer has the parameters for a data transfer
type Transfer struct {
	// The type of transfer eg "http"
	Type string
	// An optional ID that can be supplied by the client to identify the deal
	ClientID string
	// A byte array containing marshalled data specific to the transfer type
	// eg a JSON encoded struct { URL: "<url>", Headers: {...} }
	Params []byte
	// The size of the data transferred in bytes
	Size uint64
}

const DataTransferProtocol = "/fil/storage/transfer/1.0.0"

// HttpRequest has parameters for an HTTP transfer
type HttpRequest struct {
	// URL can be
	// - an http URL:
	//   "https://example.com/path"
	// - a libp2p URL:
	//   "libp2p:///ip4/xxx.xxx.xxx.xxx/tcp/4001/ipfs/QmXXX..."
	//   Must include a Peer ID
	URL string
	// Headers are the HTTP headers that are sent as part of the request,
	// eg "Authorization"
	Headers map[string]string
}

// TransportFileInfo has parameters for a transfer to be executed
type TransportFileInfo struct {
	OutputFile string
	FileId     string
	FileSize   int64
}

// TransportEvent is fired as a transfer progresses
type TransportEvent struct {
	NBytesReceived int64
	Error          error
}

// TransferStatus describes the status of a transfer (started, completed etc)
type TransferStatus string

const (
	// TransferStatusStarted is set when the transfer starts
	TransferStatusStarted TransferStatus = "TransferStatusStarted"
	// TransferStatusStarted is set when the transfer restarts after previously starting
	TransferStatusRestarted TransferStatus = "TransferStatusRestarted"
	TransferStatusOngoing   TransferStatus = "TransferStatusOngoing"
	TransferStatusCompleted TransferStatus = "TransferStatusCompleted"
	TransferStatusFailed    TransferStatus = "TransferStatusFailed"
)

// TransferState describes a transfer's current state
type TransferState struct {
	ID         string
	LocalAddr  string
	RemoteAddr string
	Status     TransferStatus
	Sent       uint64
	Received   uint64
	Message    string
	PayloadCid cid.Cid
}
