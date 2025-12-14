package kalshi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
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
	var keyBb []byte
	if k.useFile {
		// open key file
		f, err := os.Open(k.filepath)
		if err != nil {
			return fmt.Errorf("os.Open: %w", err)
		}
		keyBb, err = io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("io.ReadAll: %w", err)
		}
	} else { // use key from env
		keyBb = []byte(k.key)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(keyBb)
	if err != nil {
		return fmt.Errorf("x509.ParsePKCS1PrivateKey: %w", err)
	}

	// get hash
	path := req.URL.Path
	method := req.Method
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	toSign := ts + method + path
	msgHash := sha256.New()
	if _, err := msgHash.Write([]byte(toSign)); err != nil {
		return fmt.Errorf("msgHash.Write: %w", err)
	}

	// get hash of request body
	// bb, err := io.ReadAll(req.Body)
	// if err != nil {
	// 	return "", fmt.Errorf("io.ReadAll: %w", err)
	// }
	// req.Body.Close()
	// req.Body = io.NopCloser(bytes.NewBuffer(bb))

	// sign key
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, msgHash.Sum(nil))
	if err != nil {
		return fmt.Errorf("rsa.SignPKCS1v15: %w", err)
	}

	// set auth headers on request
	req.Header.Add(HeaderAccessKey, k.keyId)
	req.Header.Add(HeaderAccessSignature, string(signature))
	req.Header.Add(HeaderAccessTimestamp, ts)

	return nil
}
