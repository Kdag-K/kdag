package common

import (
	"encoding/hex"
	"fmt"

	"github.com/palantir/stacktrace"
)

// EncodeToString returns the UPPERCASE string representation of hexBytes with
// the 0X prefix.
func EncodeToString(hexBytes []byte) string {
	return fmt.Sprintf("0X%X", hexBytes)
}

// DecodeFromString converts a hex string with 0X prefix to a byte slice.
func DecodeFromString(hexString string) ([]byte, error) {
	hash, err := hex.DecodeString(hexString[2:])
	if err != nil {
		return nil, stacktrace.NewError("DecodeFromString failed, the error is %v", err) //nolint:wrapcheck
	}

	return hash, nil
}
