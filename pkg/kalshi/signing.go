package kalshi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	HeaderAccessKey       = "KALSHI-ACCESS-KEY"
	HeaderAccessSignature = "KALSHI-ACCESS-SIGNATURE"
	HeaderAccessTimestamp = "KALSHI-ACCESS-TIMESTAMP"
)

//go:generate mockgen -destination ../mocks/signing.go -package=mocks . KeySignerLogic
type KeySignerLogic interface {
	SignRequestWithRSAKey(req *http.Request) error
}

type KeySigner struct {
	filepath string
	key      string
	keyId    string
	useFile  bool
}

func NewKeySigner(filepath, key, keyId string, useFile bool) *KeySigner {
	return &KeySigner{
		filepath: filepath,
		key:      key,
		keyId:    keyId,
		useFile:  useFile,
	}
}

func (k *KeySigner) SignRequestWithRSAKey(req *http.Request) error {
	// get key
	var decoded *pem.Block
	if k.useFile {
		// open key file
		f, err := os.Open(k.filepath)
		if err != nil {
			return fmt.Errorf("os.Open: %w", err)
		}
		keyBb, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("io.ReadAll: %w", err)
		}
		decoded, _ = pem.Decode(keyBb)
		if decoded == nil {
			return errors.New("pem.Decode: nil block")
		}
	} else { // use key from env
		decoded, _ = pem.Decode([]byte(k.key))
		if decoded == nil {
			return errors.New("pem.Decode: nil block")
		}
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(decoded.Bytes)
	if err != nil {
		return fmt.Errorf("x509.ParsePKCS1PrivateKey: %w", err)
	}

	// get hash
	path := req.URL.Path
	method := req.Method
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)

	toSign := ts + method + path
	msgHash := sha256.New()
	if _, err := msgHash.Write([]byte(toSign)); err != nil {
		return fmt.Errorf("msgHash.Write: %w", err)
	}

	// sign key
	signature, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, msgHash.Sum(nil), nil)
	if err != nil {
		return fmt.Errorf("rsa.SignPKCS1v15: %w", err)
	}
	encodedSignature := base64.StdEncoding.EncodeToString(signature)

	// set auth headers on request
	req.Header.Add(HeaderAccessKey, k.keyId)
	req.Header.Add(HeaderAccessSignature, encodedSignature)
	req.Header.Add(HeaderAccessTimestamp, ts)

	return nil
}
