package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	mrand "math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	cInstance     []byte = nil
	aesKeyVersion int64
	cOnce         sync.Once
)

func GetAESKey() ([]byte, int64) {
	created := false
	aesKeyVersion = time.Now().Unix()
	if cInstance == nil {
		cOnce.Do(func() {
			cInstance = GenAESKey()
			created = true
		})
	}

	if created {
		log.Debugf("create a aes key %s\n", string(cInstance))

	}
	return cInstance, aesKeyVersion
}

func SetAESKey(aesString []byte, version int64) {
	cInstance = aesString
	aesKeyVersion = version
}

func GenRSAKeyPairs(keyDir string, keyName string, bitSize int) {
	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		err := os.MkdirAll(keyDir, 0755)
		CheckError(err)

	}
	base := filepath.Join(keyDir, keyName)
	pubPath := base + ".pub"
	priPath := base + ".pem"
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	CheckError(err)

	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	if _, err := os.Stat(priPath); os.IsNotExist(err) {
		pemFile, err := os.Create(priPath)
		pemFile.Chmod(0400)
		CheckError(err)
		defer pemFile.Close()
		err = pem.Encode(pemFile, privateKey)
		CheckError(err)
	}

	if _, err := os.Stat(pubPath); os.IsNotExist(err) {
		pubFile, err := os.Create(pubPath)
		pubFile.Chmod(0644)
		CheckError(err)
		defer pubFile.Close()
		pubKey := key.PublicKey
		//asn1Bytes, err := asn1.Marshal(pubKey)
		defPkix := x509.MarshalPKCS1PublicKey(&pubKey)
		//CheckError(err)
		var publicKey = &pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: defPkix,
		}
		err = pem.Encode(pubFile, publicKey)
		CheckError(err)
	}

}

func RSAEncrypt(publicKey []byte, origData string) (string, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return "", errors.New("public key error")
	}
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	//pub := pubInterface.(*rsa.PublicKey)
	label := []byte("")
	sha256hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(sha256hash, rand.Reader, pub, []byte(origData), label)
	decodedtext := base64.StdEncoding.EncodeToString(ciphertext)
	return decodedtext, err
	//return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)

}

func RSADecrypt(privateKey []byte, ciphertext string) (string, error) {
	block, _ := pem.Decode(privateKey)
	decodedtext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed, error=%s\n", err.Error())
	}

	if block == nil {
		return "", errors.New("private key error")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	sha256hash := sha256.New()
	decryptedtext, err := rsa.DecryptOAEP(sha256hash, rand.Reader, priv, decodedtext, nil)
	if err != nil {
		return "", fmt.Errorf("RSA decrypt failed, error=%s\n", err.Error())
	}
	return string(decryptedtext), err
	//return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)

}

func GetPublicKey(pkiDir string, name string) []byte {
	tgt := filepath.Join(pkiDir, name+".pub")
	content, err := ioutil.ReadFile(tgt)
	CheckError(err)
	return content
}

func GetPrivateKey(pkiDir string, name string) []byte {
	tgt := filepath.Join(pkiDir, name+".pem")
	content, err := ioutil.ReadFile(tgt)
	CheckError(err)
	return content
}

func GenAESKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	CheckError(err)
	return key
}

func AESEncrypt(data []byte) ([]byte, int64, error) {
	aesKey, version := GetAESKey()
	block, err := aes.NewCipher(aesKey)
	CheckError(err)
	blockSize := block.BlockSize()
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	rawData := append(data, padText...)
	blockMode := cipher.NewCBCEncrypter(block, aesKey[:blockSize])
	result := make([]byte, len(rawData))
	blockMode.CryptBlocks(result, rawData)
	return result, version, nil
}

func AESDecrypt(data []byte, version int64) ([]byte, error) {
	aesKey, aesKeyVersion := GetAESKey()
	log.Debug(aesKeyVersion)
	log.Debug(version)
	if aesKeyVersion != version {
		return nil, DecryptDataFailure
	}
	block, err := aes.NewCipher(aesKey)
	CheckError(err)
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, aesKey[:blockSize])
	result := make([]byte, len(data))
	blockMode.CryptBlocks(result, data)
	length := len(result)
	unpadding := int(result[length-1])
	return result[:(length - unpadding)], nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenToken(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[mrand.Intn(len(letterBytes))]
	}
	encoded := base64.StdEncoding.EncodeToString(b)
	return []byte(encoded[:n])
}

func Test() {
	//key := GetAESKey()
	//fmt.Println(key)
	//d, _ := AESEncrypt([]byte("122"))
	//r, _ := AESDecrypt(d)
	//fmt.Printf("%s", r)
	//GenRSAKeyPairs("/tmp", "go_test", 2048)
	log.Info(GenToken(64))
}
