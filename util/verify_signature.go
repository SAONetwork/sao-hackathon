package util

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"hash"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ipsn/go-secp256k1"
	"golang.org/x/crypto/sha3"

	"github.com/gin-gonic/gin"
)

const RecoveryIDOffset = 64
const SignatureLength = 64 + 1

func VerifySignature(c *gin.Context) {
	address := c.GetHeader("address")
	signature := c.GetHeader("signature")
	message := c.GetHeader("signatureMessage")
	c.Set("User", "")
	if address != "" && signature != "" && message != "" {
		message, _ = url.QueryUnescape(message)
		if verifyEthereumSignature(address, signature, message) {
			addr := common.HexToAddress(address)
			c.Set("User", addr.Hex())
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				failResponse("invalid.signature", "invalid signature"))
			return
		}
	}
	c.Next()
}

func TextAndHash(data []byte) ([]byte, string) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), string(data))
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(msg))
	return hasher.Sum(nil), msg
}

func TextHash(data []byte) []byte {
	hash, _ := TextAndHash(data)
	return hash
}

func verifyEthereumSignature(from, signature, message string) bool {
	sig, err := MustDecode(signature)
	if err != nil {
		return false
	}
	msg := TextHash([]byte(message))
	sig[RecoveryIDOffset] -= 27

	recovered, err := SigToPub(msg, sig)
	if err != nil {
		return false
	}
	recoveredAddr := PubkeyToAddress(*recovered)
	return strings.ToLower(from) == strings.ToLower(recoveredAddr.Hex())
}

func MustDecode(input string) ([]byte, error) {
	dec, err := hexutil.Decode(input)
	if err != nil {
		return nil, err
	}
	return dec, nil
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(S256(), pub.X, pub.Y)
}

type KeccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

func NewKeccakState() KeccakState {
	return sha3.NewLegacyKeccak256().(KeccakState)
}

func Keccak256(data ...[]byte) []byte {
	b := make([]byte, 32)
	d := NewKeccakState()
	for _, b := range data {
		d.Write(b)
	}
	d.Read(b)
	return b
}

func PubkeyToAddress(p ecdsa.PublicKey) common.Address {
	pubBytes := FromECDSAPub(&p)
	return common.BytesToAddress(Keccak256(pubBytes[1:])[12:])
}

func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	btcsig := make([]byte, SignatureLength)
	btcsig[0] = sig[64] + 27
	copy(btcsig[1:], sig)

	pub, _, err := btcec.RecoverCompact(btcec.S256(), btcsig, hash)
	return (*ecdsa.PublicKey)(pub), err
}

func failResponse(code string, message string) gin.H {
	return gin.H{
		"code":      code,
		"message":   message,
		"timestamp": time.Now().UnixMilli(),
	}
}

func S256() elliptic.Curve {
	return secp256k1.S256()
}
