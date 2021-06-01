package mobile

import (
	"fmt"
	"os"

	"github.com/Kdag-K/kdag/src/crypto/keys"
)

// GetPrivPublKeys generates a new public key pair and returns it in the
// following formatted string <public key hex>=!@#@!=<private key hex>.
func GetPrivPublKeys() string {
	key, err := keys.GenerateECDSAKey()
	if err != nil {
		fmt.Println("Error generating new key")
		os.Exit(2)
	}

	priv := keys.PrivateKeyHex(key)
	pub := keys.PublicKeyHex(&key.PublicKey)

	return pub + "=!@#@!=" + priv
}
