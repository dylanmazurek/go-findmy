package decryptor

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
)

func TestGenData(t *testing.T) {
	tsBytes := []byte{0x00, 0x01, 0x02, 0x03}
	k := byte(0xAA)
	result := genData(tsBytes, k)

	if len(result) != 32 {
		t.Fatalf("genData: expected length 32, got %d", len(result))
	}

	// First 11 bytes should be 0xFF.
	for i := 0; i < 11; i++ {
		if result[i] != 0xFF {
			t.Errorf("genData: index %d expected 0xFF, got %X", i, result[i])
		}
	}

	// Index 11 should equal k.
	if result[11] != k {
		t.Errorf("genData: index 11 expected %X, got %X", k, result[11])
	}

	// Positions 12 to 15 should equal tsBytes.
	if !reflect.DeepEqual(result[12:16], tsBytes) {
		t.Errorf("genData: indices 12-15 expected %v, got %v", tsBytes, result[12:16])
	}

	// Positions 16 to 26 should be zeros.
	for i := 16; i < 27; i++ {
		if result[i] != 0 {
			t.Errorf("genData: index %d expected 0, got %d", i, result[i])
		}
	}

	// Index 27 should equal k.
	if result[27] != k {
		t.Errorf("genData: index 27 expected %X, got %X", k, result[27])
	}

	// Positions 28 to 31 should equal tsBytes.
	if !reflect.DeepEqual(result[28:32], tsBytes) {
		t.Errorf("genData: indices 28-31 expected %v, got %v", tsBytes, result[28:32])
	}

	fmt.Printf("genData: %v\n", hex.EncodeToString(result))
}

func TestMaskTimestamp(t *testing.T) {
	timestamp := 0x12345678
	K := 8
	tsBytes := maskTimestamp(timestamp, K)
	expectedTimestamp := timestamp & ^((1 << K) - 1)
	gotTimestamp := binary.BigEndian.Uint32(tsBytes)
	if gotTimestamp != uint32(expectedTimestamp) {
		t.Errorf("maskTimestamp: expected %X, got %X", expectedTimestamp, gotTimestamp)
	}

	K = 0
	tsBytes = maskTimestamp(timestamp, K)
	expectedTimestamp = timestamp & ^((1 << K) - 1)
	gotTimestamp = binary.BigEndian.Uint32(tsBytes)
	if gotTimestamp != uint32(expectedTimestamp) {
		t.Errorf("maskTimestamp with K=0: expected %X, got %X", expectedTimestamp, gotTimestamp)
	}
}
