package decryptor

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/dylanmazurek/google-findmy/pkg/nova/models/protos/bindings"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type Decryptor struct {
	OwnerKey string
}

func NewDecryptor(ownerKey string) *Decryptor {
	newDecryptor := &Decryptor{
		OwnerKey: ownerKey,
	}

	return newDecryptor
}

func (d *Decryptor) DecryptLocations(ctx context.Context, deviceUpdate *bindings.DeviceUpdate) {
	log := log.Ctx(ctx)

	deviceInformation := deviceUpdate.GetDeviceMetadata().GetInformation()
	//isCustomTracker := deviceInformation.GetDeviceRegistration().GetFastPairModelId() == "003200"

	encryptedIdentityKey := deviceInformation.GetDeviceRegistration().GetEncryptedUserSecrets().GetEncryptedIdentityKey()
	//log.Info().Msgf("encrypted identity key: %s", hex.EncodeToString(encryptedIdentityKey))

	//encryptedIdentityKey := flipBits(encryptedIdentityKey, isCustomTracker)
	//ownerKey := flipBits([]byte(d.OwnerKey), isCustomTracker)

	ownerKey := []byte(d.OwnerKey)
	identityKey, err := decryptEIK(ctx, ownerKey, encryptedIdentityKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decrypt EIK")
		return
	}

	locationsProto := deviceInformation.GetLocationInformation().GetReports().GetRecentLocationAndNetworkLocations()

	// At All Areas Reports or Own Reports
	recentLocation := locationsProto.GetRecentLocation()
	recentLocationTime := locationsProto.GetRecentLocationTimestamp()

	// High Traffic Reports
	networkLocations := locationsProto.GetNetworkLocations()
	networkLocationsTime := locationsProto.GetNetworkLocationTimestamps()

	if locationsProto.GetRecentLocation() != nil {
		networkLocations = append(networkLocations, recentLocation)
		networkLocationsTime = append(networkLocationsTime, recentLocationTime)
	}

	fmt.Println("----------------------------------------")
	fmt.Println("[DecryptLocations] Decrypted Locations:")

	if len(networkLocations) == 0 {
		fmt.Println("No locations found.")
		return
	}

	for i, loc := range networkLocations {
		time := networkLocationsTime[i]

		if loc.GetStatus() == bindings.Status_SEMANTIC {
			fmt.Println("Semantic Location Report")
			fmt.Printf("Semantic Location: %s\n", loc.GetSemanticLocation().GetLocationName())
		} else {
			encryptedLocation := loc.GetGeoLocation().GetEncryptedReport().GetEncryptedLocation()
			publicKeyRandom := loc.GetGeoLocation().GetEncryptedReport().GetPublicKeyRandom()

			var decryptedLocation []byte
			if len(publicKeyRandom) == 0 {
				identityKeyHash := sha256.Sum256(identityKey)
				decryptedLocation, err = decryptAESGCM(ctx, identityKeyHash[:], encryptedLocation)
				if err != nil {
					log.Error().Err(err).Msg("Failed to decrypt AES GCM")
					continue
				}
			} else {
				log.Error().Msg("Public Key Random is not empty, not implemented")
			}

			protoLoc := &bindings.Location{}
			err = proto.Unmarshal(decryptedLocation, protoLoc)
			if err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal location proto")
				continue
			}

			latitude := float64(protoLoc.GetLatitude())   // 1e7
			longitude := float64(protoLoc.GetLongitude()) // 1e7
			altitude := protoLoc.GetAltitude()

			fmt.Printf("Latitude: %f\n", latitude)
			fmt.Printf("Longitude: %f\n", longitude)
			fmt.Printf("Altitude: %d\n", altitude)
		}

		fmt.Printf("Time: %s\n", time.String())
		fmt.Printf("Status: %s\n", loc.GetStatus())
		fmt.Printf("Is Own Report: %t\n", loc.GetGeoLocation().GetEncryptedReport().GetIsOwnReport())
		fmt.Println("----------------------------------------")
	}
}

func flipBits(data []byte, enabled bool) []byte {
	if enabled {
		flipped := make([]byte, len(data))
		for i, b := range data {
			flipped[i] = b ^ 0xFF
		}
		return flipped
	}
	return data
}

func decryptEIK(ctx context.Context, ownerKey []byte, encryptedEIK []byte) ([]byte, error) {
	if len(encryptedEIK) == 48 {
		return decryptAESCBCNoPadding(ctx, ownerKey, encryptedEIK)
	}

	if len(encryptedEIK) == 60 {
		return decryptAESGCM(ctx, ownerKey, encryptedEIK)
	}

	return nil, fmt.Errorf("the encrypted EIK has invalid length: %d", len(encryptedEIK))
}

func prepareAESKey(key []byte) []byte {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		sum := sha256.Sum256(key)
		return sum[:]
	}

	return key
}

func decryptAESCBCNoPadding(ctx context.Context, key, encryptedData []byte) ([]byte, error) {
	log := log.Ctx(ctx)

	log.Info().Msg("Decrypting AES CBC No Padding")

	validKey := prepareAESKey(key)

	log.Info().Msgf("key: %s", hex.EncodeToString(validKey))

	if len(encryptedData) <= aes.BlockSize {
		return nil, fmt.Errorf("invalid data length: %d", len(encryptedData))
	}

	iv := encryptedData[:aes.BlockSize]
	ciphertext := encryptedData[aes.BlockSize:]

	log.Info().Msgf("iv: %s", hex.EncodeToString(iv))
	log.Info().Msgf("ciphertext: %s", hex.EncodeToString(ciphertext))

	block, err := aes.NewCipher(validKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decryptedData := make([]byte, len(ciphertext))
	mode.CryptBlocks(decryptedData, ciphertext)

	return decryptedData, nil
}

func decryptAESGCM(ctx context.Context, key, encryptedData []byte) ([]byte, error) {
	log := log.Ctx(ctx)

	log.Info().Msg("Decrypting AES GCM")

	iv := encryptedData[:12]
	ciphertext := encryptedData[12:]
	keyStr := key[:32]

	log.Info().Msgf("iv: \t\t%s", hex.EncodeToString(iv))
	log.Info().Msgf("ciphertext: \t%s", hex.EncodeToString(ciphertext))

	log.Info().Msgf("key: \t\t%s", key)

	block, err := aes.NewCipher(keyStr)
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
