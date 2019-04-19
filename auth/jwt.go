package auth

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

type AuthToken struct {
	TokenType string `json:"token_type"`
	Token     string `json:"access_token"`
	ExpiresIn int64  `json:"expires_in"`
}

type AuthTokenClaim struct {
	*jwt.StandardClaims
	PolicyId    int8     `json:"pid"`
	UserId      string   `json:"uid"`
	MapLayers   []string `json:"ml"`
	MapDatabase string   `json:"md"`
}

type Keys struct {
	Keys  []Key  `json:"keys"`
	MaxId string `json:"maxid"`
}

type Key struct {
	KeyType string `json:"kty"`
	Hex     string `json:"hex"`
	Algo    string `json:"alg"`
	KeyId   string `json:"kid"`
}

func IssueToken(kid string, msg types.Message) (*AuthToken, error) {
	expiresAt := time.Now().Add(time.Minute * 1).Unix()

	// TODO: claims to be read from smart contract
	token := jwt.NewWithClaims(jwt.SigningMethodES256, &AuthTokenClaim{
		&jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
		0,
		msg.From().Hex(),
		[]string{"beaches", "barbeque"},
		"hkopendata",
	})
	token.Header["kid"] = kid

	privKeyData, err := readData("./keys/private.json")
	if err != nil {
		log.Print(err)
		return nil, err
	}

	var signkey Key
	for _, v := range privKeyData.Keys {
		if v.KeyId == kid {
			signkey = v
			break
		}
	}
	if signkey.Algo == "" {
		log.Print("key unavailable")
		return nil, errors.New("key unavailable")
	}

	key, err := crypto.HexToECDSA(signkey.Hex)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	tokenString, err := token.SignedString(key)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return &AuthToken{
		Token:     tokenString,
		TokenType: "Bearer",
		ExpiresIn: expiresAt,
	}, nil
}

func FindTransaction(txStr string) (types.Message, error) {
	ctx := context.Background()

	// TODO: CONFIG - eth dial network
	conn, _ := ethclient.Dial("https://ropsten.infura.io/")
	tx, _, err := conn.TransactionByHash(ctx, common.HexToHash(txStr))
	if err != nil {
		return types.Message{}, err
	}
	return tx.AsMessage(types.NewEIP155Signer(tx.ChainId()))
}

func NewKeySet(keyfolder string) error {
	// TODO: change key folder and privkey permission to 0600 and pubkey files to 0644
	// TODO: CONFIG - set below hard coded values in config
	if _, err := os.Stat(keyfolder); os.IsNotExist(err) {
		os.MkdirAll(keyfolder, 0755)
	}
	pubKeyData, err := readData(keyfolder + "public.json")
	privKeyData, err := readData(keyfolder + "private.json")

	kid := "0001"
	algo := "ES256"
	kty := "EC"

	privateKey, err := crypto.GenerateKey()
	privHex := hex.EncodeToString(crypto.FromECDSA(privateKey))
	privK := &Key{
		Algo:    algo,
		KeyType: kty,
		Hex:     privHex,
		KeyId:   kid,
	}
	if err != nil {
		log.Fatal(err)
	}

	pubKey := privateKey.Public().(*ecdsa.PublicKey)
	pubHex := hex.EncodeToString(crypto.FromECDSAPub(pubKey))
	pubK := &Key{
		Algo:    algo,
		KeyType: kty,
		Hex:     pubHex,
		KeyId:   kid,
	}
	if privKeyData == nil {
		pubKeyData = &Keys{[]Key{*pubK}, kid}
		privKeyData = &Keys{[]Key{*privK}, kid}
	} else {
		maxid, err := strconv.Atoi(privKeyData.MaxId)
		if err != nil {
			log.Fatal(err)
		}
		pubK.KeyId, privK.KeyId, pubKeyData.MaxId, privKeyData.MaxId = fmt.Sprintf("%04d", maxid+1), fmt.Sprintf("%04d", maxid+1), fmt.Sprintf("%04d", maxid+1), fmt.Sprintf("%04d", maxid+1)
		pubKeyData.Keys = append(pubKeyData.Keys, *pubK)
		privKeyData.Keys = append(privKeyData.Keys, *privK)
	}
	err = writeData(keyfolder+"public.json", pubKeyData)
	if err != nil {
		log.Fatal(err)
	}
	err = writeData(keyfolder+"private.json", privKeyData)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func readData(file string) (*Keys, error) {
	var structKeys Keys

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	byteValue, _ := ioutil.ReadAll(f)
	json.Unmarshal([]byte(byteValue), &structKeys)

	return &structKeys, nil
}

func writeData(file string, data *Keys) error {
	jsonKeys, _ := json.MarshalIndent(data, "", " ")
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return ioutil.WriteFile(file, jsonKeys, 0644)
}
