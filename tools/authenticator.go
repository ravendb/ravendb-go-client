package tools

import (
	"net/http"
	"github.com/GoKillers/libsodium-go/cryptobox"
	"github.com/GoKillers/libsodium-go/randombytes"
	"errors"
	b64 "encoding/base64"
	"encoding/hex"
)

type AuthError struct {//todo
	Url string
	ApiKey string
	Err error
}

type Authenticator struct{
	publicKeys map[string]string
}

//func NewAuthenticator() *Authenticator{
//
//}

func (authenticator Authenticator) Authenticate(url string, apiKey string, headers []http.Header){
	if serverPk, ok := authenticator.publicKeys[url]; !ok{
		serverPk = authenticator.GetServerPK(url)
		authenticator.publicKeys[url] = serverPk
	}
	//chunks := strings.Split(apiKey, "/")
	//name, secret := chunks[0], chunks[1]


}

func (authenticator Authenticator) GetServerPK(url string) string{

	return ""
}

func (authenticator Authenticator) GenerateKeyPair() ([]byte, []byte, error){
	pk, sk, exitCode := cryptobox.CryptoBoxKeyPair()
	if exitCode != 0{
		return nil, nil, errors.New("tools: Error generating keypairs")
	}
	return pk, sk, nil
}

func (authenticator Authenticator) BuildServerRequest(pk []byte, sk []byte, secret string, serverPk []byte) map[string][]byte{
	nonce := randombytes.RandomBytes(cryptobox.CryptoBoxNonceBytes())

	dataBytes := []byte(secret)
	dataPadded := append(dataBytes, randombytes.RandomBytes((64 - (len(dataBytes) % 64)))...)
	encryptedSecret, exitCode := cryptobox.CryptoBoxEasy(dataPadded, nonce, serverPk, sk)
	if exitCode != 0{
		errors.New("tools: Error encrypting secret")
	}
	data := make(map[string][]byte)
	data["Secret"] = make([]byte, hex.EncodedLen(len(encryptedSecret)))
	b64.RawStdEncoding.Encode(data["Secret"], encryptedSecret)
	data["PublicKey"] = make([]byte, hex.EncodedLen(len(pk)))
	b64.RawStdEncoding.Encode(data["PublicKey"], pk)
	data["Nonce"] = make([]byte, hex.EncodedLen(len(nonce)))
	b64.RawStdEncoding.Encode(data["Nonce"], nonce)
	data["ServerKey"] = make([]byte, hex.EncodedLen(len(serverPk)))
	b64.RawStdEncoding.Encode(data["ServerKey"], serverPk)
	return data
}