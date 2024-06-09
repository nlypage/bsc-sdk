package bsc_sdk

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/qushedo/bsc-sdk/types"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"math/big"
	"strconv"
	"strings"
)

type Wallet struct {
	Address       string
	AddressBase58 string
	PrivateKey    string
	PublicKey     string
	ethClient     *ethclient.Client
}

func MnemonicToWallet(networkUrl types.NetworkUrl, mnemonic, accountPath string) (*Wallet, error) {
	seed := bip39.NewSeed(mnemonic, "")
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to create master bip32Key: %w", err)
	}

	// Split the path and parse each component
	segments := strings.Split(accountPath, "/")
	var bip32Key = masterKey
	for _, segment := range segments[1:] { // skipping the 'm' part
		var hardened bool
		if strings.HasSuffix(segment, "'") {
			hardened = true
			segment = segment[:len(segment)-1]
		}

		index, err := strconv.Atoi(segment)
		if err != nil {
			return nil, fmt.Errorf("invalid path segment '%s': %w", segment, err)
		}

		if hardened {
			bip32Key, err = bip32Key.NewChildKey(uint32(index) + bip32.FirstHardenedChild)
		} else {
			bip32Key, err = bip32Key.NewChildKey(uint32(index))
		}
		if err != nil {
			return nil, fmt.Errorf("failed to derive bip32Key at %s: %w", segment, err)
		}
	}

	privateKey, _ := crypto.HexToECDSA(hex.EncodeToString(bip32Key.Key))
	publicKeyHex := convertPublicKeyToHex(privateKey.Public().(*ecdsa.PublicKey))
	address := getAddressFromPublicKey(privateKey.Public().(*ecdsa.PublicKey))

	client, errConnect := ethclient.Dial(string(networkUrl))
	if errConnect != nil {
		return nil, errConnect
	}

	return &Wallet{
		Address:    address,
		PrivateKey: hex.EncodeToString(bip32Key.Key),
		PublicKey:  publicKeyHex,
		ethClient:  client,
	}, nil

}

func (w *Wallet) Balance() (*big.Int, error) {
	address := common.HexToAddress(w.Address)
	balance, err := w.ethClient.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to get balance: %v", err))
	}

	return balance, nil
}

func (w *Wallet) CalculateTransactionFee() (*big.Int, error) {
	gasPrice, err := w.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	fee := new(big.Int).Mul(gasPrice, big.NewInt(0).SetUint64(21000))

	return fee, nil
}

func (w *Wallet) Transfer(toAddress string, amount int64) (string, error) {
	nonce, err := w.ethClient.PendingNonceAt(context.Background(), common.HexToAddress(w.Address))
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %v", err)
	}

	gasPrice, err := w.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %v", err)
	}
	gasLimit := uint64(21000)

	toAddressCommon := common.HexToAddress(toAddress)
	tx := ethTypes.NewTx(&ethTypes.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddressCommon,
		Value:    big.NewInt(amount),
		Data:     []byte{},
	})

	chainID, err := w.ethClient.NetworkID(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get chain ID: %v", err)
	}

	ecdsaPrivateKey, err := crypto.HexToECDSA(w.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to get ecdsa private key from hex: %v", err)
	}
	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(chainID), ecdsaPrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %v", err)
	}

	err = w.ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %v", err)
	}

	return signedTx.Hash().Hex(), nil
}

func convertPublicKeyToHex(publicKey *ecdsa.PublicKey) string {

	privateKeyBytes := crypto.FromECDSAPub(publicKey)

	return hexutil.Encode(privateKeyBytes)[2:]
}

func getAddressFromPublicKey(publicKey *ecdsa.PublicKey) string {

	address := crypto.PubkeyToAddress(*publicKey).Hex()

	address = "0x" + address[2:]

	return address
}
