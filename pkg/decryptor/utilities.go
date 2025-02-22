package decryptor

import (
	"crypto/aes"
	"crypto/elliptic"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/deatil/go-cryptobin/elliptic/secp"
)

func genData(tsBytes []byte, k byte) []byte {
	data := make([]byte, 32)
	for i := 0; i < 11; i++ {
		data[i] = 0xFF
	}

	data[11] = k
	copy(data[12:16], tsBytes)
	copy(data[16:27], make([]byte, 11))
	data[27] = k
	copy(data[28:32], tsBytes)

	return data
}

func maskTimestamp(timestamp int, K int) []byte {
	mask := ^((1 << K) - 1)
	timestamp &= mask

	tsBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(tsBytes, uint32(timestamp))

	return tsBytes
}

func calculateR(identityKey []byte, timestamp int) (*big.Int, error) {
	const k = 10
	tsBytes := maskTimestamp(timestamp, k)

	data := genData(tsBytes, byte(k))

	cipherBlock, err := aes.NewCipher(identityKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	rDash := make([]byte, len(data))
	blockSize := cipherBlock.BlockSize()

	for i := 0; i < len(data); i += blockSize {
		cipherBlock.Encrypt(rDash[i:i+blockSize], data[i:i+blockSize])
	}

	rDashInt := new(big.Int).SetBytes(rDash)

	n := secp.P160r1().Params().N
	r := new(big.Int).Mod(rDashInt, n)

	return r, nil
}

func rxToRy(rx big.Int, curve *elliptic.CurveParams) (*big.Int, error) {
	a := new(big.Int).Sub(big.NewInt(0), big.NewInt(3))

	var ryy *big.Int = new(big.Int)
	ryy.Exp(&rx, big.NewInt(3), curve.P)
	ryy.Add(ryy, new(big.Int).Mul(a, &rx))
	ryy.Add(ryy, curve.B)
	ryy.Mod(ryy, curve.P)

	var ry *big.Int = new(big.Int)
	ry.Add(curve.P, big.NewInt(1))
	ry.Div(ry, big.NewInt(4))
	ry.Exp(ryy, ry, curve.P)

	if new(big.Int).Exp(ry, big.NewInt(2), curve.P).Cmp(ryy) != 0 {
		return nil, fmt.Errorf("invalid y coordinate")
	}

	if new(big.Int).Mod(ry, big.NewInt(2)).Cmp(big.NewInt(0)) != 0 {
		ry.Sub(curve.P, ry)
	}

	return ry, nil
}
