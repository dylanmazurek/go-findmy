package decryptor

import (
	"crypto/aes"
	"crypto/sha256"
	"fmt"
	"math/big"

	"golang.org/x/crypto/hkdf"

	"github.com/ProtonMail/go-crypto/eax"
	"github.com/deatil/go-cryptobin/elliptic/secp"
	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	"github.com/dylanmazurek/go-findmy/pkg/shared/models"
	"google.golang.org/protobuf/proto"
)

func decryptSemantic(loc *bindings.LocationReport) (*models.LocationReport, error) {
	semanticLocation := loc.GetSemanticLocation()
	if semanticLocation == nil {
		return nil, fmt.Errorf("semantic location is nil")
	}

	semanticLocationName := semanticLocation.GetLocationName()

	newLocation := &models.LocationReport{
		ReportType:   models.ReportTypeSemantic,
		SemanticName: &semanticLocationName,
	}

	return newLocation, nil
}

func decryptReport(loc *bindings.LocationReport, identityKey []byte) (*models.LocationReport, error) {
	encryptedLocation := loc.GetGeoLocation().GetEncryptedReport().GetEncryptedLocation()
	publicKeyRandom := loc.GetGeoLocation().GetEncryptedReport().GetPublicKeyRandom()
	hasPublicKey := len(publicKeyRandom) != 0

	var decryptedLocation []byte
	var err error
	if !hasPublicKey {
		location, err := decryptLocation(encryptedLocation, identityKey)
		if err != nil {
			return nil, err
		}

		decryptedLocation = location
	} else {
		beaconTimerCounter := loc.GetGeoLocation().DeviceTimeOffset
		decryptedLocation, err = decryptLocationWithPublicKey(encryptedLocation, publicKeyRandom, beaconTimerCounter, identityKey)
		if err != nil {
			return nil, err
		}
	}

	protoLoc := &bindings.Location{}
	err = proto.Unmarshal(decryptedLocation, protoLoc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted location: %w", err)
	}

	newLocation := &models.LocationReport{
		ReportType: models.ReportTypeLocation,
		Latitude:   float64(protoLoc.GetLatitude()) / 1e7,
		Longitude:  float64(protoLoc.GetLongitude()) / 1e7,
		Altitude:   float64(protoLoc.GetAltitude()),
	}

	return newLocation, nil
}

func decryptLocation(encryptedLocation []byte, identityKey []byte) ([]byte, error) {
	identityKeyHash := sha256.Sum256(identityKey)
	decryptedLocation, err := decryptAesGcm(identityKeyHash[:], encryptedLocation)
	if err != nil {
		return nil, err
	}

	return decryptedLocation, nil
}

func decryptLocationWithPublicKey(encryptedAndTag []byte, sxBytes []byte, beaconTime uint32, identityKey []byte) ([]byte, error) {
	var curve = secp.P160r1()

	encryptedMessage := encryptedAndTag[:len(encryptedAndTag)-16]
	tag := encryptedAndTag[len(encryptedAndTag)-16:]

	rxInt, err := calculateR(identityKey, int(beaconTime))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate R: %w", err)
	}

	sxInt := new(big.Int).SetBytes(sxBytes)
	syInt, err := rxToRy(*sxInt, curve.Params())
	if err != nil {
		return nil, fmt.Errorf("failed to calculate Ry: %w", err)
	}

	sxCoord, _ := curve.ScalarMult(sxInt, syInt, rxInt.Bytes())

	hkdf := hkdf.New(sha256.New, sxCoord.Bytes(), nil, nil)
	k := make([]byte, 32)
	_, err = hkdf.Read(k)
	if err != nil {
		return nil, fmt.Errorf("failed to read from hkdf: %w", err)
	}

	R_x, _ := curve.ScalarBaseMult(rxInt.Bytes())

	lRx := R_x.Bytes()[12:]
	lSx := sxInt.Bytes()[12:]

	nonce := append(lRx, lSx...)

	decrypted, err := decryptAes(encryptedMessage, tag, nonce, k)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return decrypted, nil
}

func decryptAes(data []byte, tag []byte, nonce []byte, key []byte) ([]byte, error) {
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	ciphertextWithTag := append(data, tag...)

	eaxInstance, err := eax.NewEAX(aesCipher)
	if err != nil {
		return nil, fmt.Errorf("failed to create EAX: %w", err)
	}

	decrypted, err := eaxInstance.Open(nil, nonce, ciphertextWithTag, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return decrypted, nil
}
