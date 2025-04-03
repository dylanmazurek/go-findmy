package decryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

type Decryptor struct {
	OwnerKey string
}

func NewDecryptor(ownerKey *string) (*Decryptor, error) {
	if ownerKey == nil {
		err := fmt.Errorf("owner key is nil")
		return nil, err
	}

	newDecryptor := &Decryptor{
		OwnerKey: *ownerKey,
	}

	return newDecryptor, nil
}

func decryptEik(ownerKey []byte, encryptedEik []byte) ([]byte, error) {
	eikLen := len(encryptedEik)

	var err error
	var validKey []byte
	switch eikLen {
	case 48:
		validKey, err = decryptAesNoPadding(ownerKey, encryptedEik)
	case 60:
		validKey, err = decryptAesGcm(ownerKey, encryptedEik)
	default:
		err = fmt.Errorf("invalid eik length: %d", eikLen)
	}

	return validKey, err
}

func decryptAesNoPadding(ownerKey []byte, encryptedData []byte) ([]byte, error) {
	validKey := prepareAESKey(ownerKey)

	if len(encryptedData) <= aes.BlockSize {
		err := fmt.Errorf("invalid data length: %d", len(encryptedData))
		return nil, err
	}

	iv := encryptedData[:aes.BlockSize]
	ciphertext := encryptedData[aes.BlockSize:]

	block, err := aes.NewCipher(validKey)
	if err != nil {
		err := fmt.Errorf("failed to create aes cipher: %w", err)
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decryptedData := make([]byte, len(ciphertext))
	mode.CryptBlocks(decryptedData, ciphertext)

	return decryptedData, nil
}

func decryptAesGcm(key, encryptedData []byte) ([]byte, error) {
	if len(encryptedData) < 12 {
		err := fmt.Errorf("invalid data length: %d", len(encryptedData))
		return nil, err
	}

	iv := encryptedData[:12]
	ciphertext := encryptedData[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		err := fmt.Errorf("failed to create AES cipher: %w", err)
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		err := fmt.Errorf("failed to create GCM: %w", err)
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		err := fmt.Errorf("failed to decrypt: %w", err)
		return nil, err
	}

	return plaintext, nil
}
