package node

import (
	"crypto/ecdsa"

	"github.com/mosaicnetworks/babble/src/crypto/keys"
)
//Validator struct holds information about the validator for a node
type Validator struct {
	Key     *ecdsa.PrivateKey
	Moniker string

	id       uint32
	pubBytes []byte
	pubHex   string
}
