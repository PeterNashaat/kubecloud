package internal

import (
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
	keyPair, err := sr25519.Scheme{}.FromPhrase(mnemonic, "")
	if err != nil {
		return nil, "", err
	}
	address, err := keyPair.SS58Address(SS58AddressFormat)
	if err != nil {
		return nil, "", err
	}
	return keyPair, address, nil
}
