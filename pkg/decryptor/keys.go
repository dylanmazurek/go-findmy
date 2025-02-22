package decryptor

import "crypto/sha256"

func prepareAESKey(key []byte) []byte {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		sum := sha256.Sum256(key)
		return sum[:]
	}

	return key
}
