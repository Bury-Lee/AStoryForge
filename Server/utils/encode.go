package utils

import (
	"encoding/base64"
	"os"
)

// EncodeFile encodes a local file into a Base64 string.
func EncodeFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}
