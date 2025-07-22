package internal

import (
	"errors"
	"strings"

	"github.com/vedhavyas/go-subkey"
	"github.com/vedhavyas/go-subkey/sr25519"
)

// Blockchain-related constants
const (
	// SS58AddressFormat is the format used for Substrate-based chain addresses
	SS58AddressFormat = 42
)

// KeyPairFromMnemonic derives a keypair and SS58 address from a mnemonic
func KeyPairFromMnemonic(mnemonic string) (subkey.KeyPair, string, error) {
	mnemonic = strings.TrimSpace(mnemonic)
	if mnemonic == "" {
		return nil, "", errors.New("mnemonic cannot be empty")
	}
	words := strings.Fields(mnemonic)
	if len(words) < 12 {
		return nil, "", errors.New("mnemonic must be at least 12 words")
	}
	keyPair, err := sr25519.Scheme{}.FromPhrase(mnemonic, "")
	if err != nil {
		return nil, "", err
	}
	address, err := AccountAddressFromKeypair(keyPair)
	if err != nil {
		return nil, "", err
	}
	return keyPair, address, nil
}

// AccountAddressFromKeypair returns the SS58 address for a given keypair
func AccountAddressFromKeypair(keyPair subkey.KeyPair) (string, error) {
	return keyPair.SS58Address(SS58AddressFormat)
}

// AccountFromMnemonic returns the SS58 address from a mnemonic
func AccountFromMnemonic(mnemonic string) (string, error) {
	keyPair, _, err := KeyPairFromMnemonic(mnemonic)
	if err != nil {
		return "", err
	}
	return AccountAddressFromKeypair(keyPair)
}
