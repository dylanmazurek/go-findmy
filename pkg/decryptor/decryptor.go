package decryptor

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/rs/zerolog/log"
)

type Decryptor struct {
	OwnerKey string
}

func NewDecryptor(ownerKey *string) (*Decryptor, error) {
	if ownerKey == nil {
		return nil, fmt.Errorf("owner key is nil")
	}

	newDecryptor := &Decryptor{
		OwnerKey: *ownerKey,
	}

	return newDecryptor, nil
}

func decryptEik(ctx context.Context, ownerKey []byte, encryptedEik []byte) ([]byte, error) {
	log := log.Ctx(ctx)
	_ = log

	eikLen := len(encryptedEik)
	if eikLen == 48 {
		return decryptAesNoPadding(ctx, ownerKey, encryptedEik)
	}

	if eikLen == 60 {
		log.Info().Msgf("owner key: %x", ownerKey)
		log.Info().Msgf("encrypted Eik: %x", encryptedEik)

		return decryptAesGcm(ctx, ownerKey, encryptedEik)
	}

	return nil, fmt.Errorf("the encrypted Eik has invalid length: %d", eikLen)
}

func decryptAesNoPadding(ctx context.Context, ownerKey []byte, encryptedData []byte) ([]byte, error) {
	log := log.Ctx(ctx)
	_ = log

	validKey := prepareAESKey(ownerKey)

	if len(encryptedData) <= aes.BlockSize {
		return nil, fmt.Errorf("invalid data length: %d", len(encryptedData))
	}

	iv := encryptedData[:aes.BlockSize]
	ciphertext := encryptedData[aes.BlockSize:]

	block, err := aes.NewCipher(validKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create aes cipher: %w", err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decryptedData := make([]byte, len(ciphertext))
	mode.CryptBlocks(decryptedData, ciphertext)

	return decryptedData, nil
}

func decryptAesGcm(ctx context.Context, key, encryptedData []byte) ([]byte, error) {
	log := log.Ctx(ctx)
	_ = log

	iv := encryptedData[:12]
	ciphertext := encryptedData[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := aesgcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}
