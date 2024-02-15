package underflow

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/onflow/cadence"
)

// go string to cadence string panic if error
func cadenceString(input string) cadence.String {
	value, err := cadence.NewString(input)
	if err != nil {
		panic(err)
	}
	return value
}

func getAndUnquoteString(value cadence.Value) string {
	result, err := strconv.Unquote(value.String())
	if err != nil {
		result = value.String()
		if strings.Contains(result, "\\u") || strings.Contains(result, "\\U") {
			result = value.ToGoValue().(string)
		}
	}

	return strings.Replace(result, "\x00", "", -1)
}

// HexToAddress converts a hex string to an Address.
func hexToAddress(h string) (*cadence.Address, error) {
	trimmed := strings.TrimPrefix(h, "0x")
	if len(trimmed)%2 == 1 {
		trimmed = "0" + trimmed
	}
	b, err := hex.DecodeString(trimmed)
	if err != nil {
		return nil, err
	}
	address := cadence.BytesToAddress(b)
	return &address, nil
}
