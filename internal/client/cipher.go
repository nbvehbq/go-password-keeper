package client

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

const blockLength = 128

func setupKeyPair() ([]byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return privateKeyPEM.Bytes(), nil
}

func encrypt(key, data []byte) ([]byte, error) {
	privateKeyBlock, _ := pem.Decode(key)
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	encryptedData := make([]byte, 0, len(data))
	var nextBlockLength int
	for i := 0; i < len(data); i += blockLength {
		nextBlockLength = i + blockLength
		if nextBlockLength > len(data) {
			nextBlockLength = len(data)
		}
		block, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &privateKey.PublicKey, data[i:nextBlockLength], []byte("yandex"))
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt data '%s': %v", data, err)
		}
		encryptedData = append(encryptedData, block...)
	}
	return encryptedData, nil
}

func decrypt(key, data []byte) ([]byte, error) {
	privateKeyBlock, _ := pem.Decode(key)
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	decryptedData := make([]byte, 0, len(data))
	var nextBlockLength int
	for i := 0; i < len(data); i += privateKey.PublicKey.Size() {
		nextBlockLength = i + privateKey.PublicKey.Size()
		if nextBlockLength > len(data) {
			nextBlockLength = len(data)
		}
		block, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, data[i:nextBlockLength], []byte("yandex"))
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt data: %v", err)
		}
		decryptedData = append(decryptedData, block...)
	}

	return decryptedData, nil
}
