package token

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate_ReturnsNonEmpty(t *testing.T) {
	raw, hash, err := Generate()
	require.NoError(t, err)
	assert.NotEmpty(t, raw)
	assert.NotEmpty(t, hash)
}

func TestGenerate_ReturnsDifferentValuesEachCall(t *testing.T) {
	raw1, hash1, err1 := Generate()
	require.NoError(t, err1)

	raw2, hash2, err2 := Generate()
	require.NoError(t, err2)

	assert.NotEqual(t, raw1, raw2)
	assert.NotEqual(t, hash1, hash2)
}

func TestHash_ConsistentWithGenerate(t *testing.T) {
	raw, expectedHash, err := Generate()
	require.NoError(t, err)

	computedHash, err := Hash(raw)
	require.NoError(t, err)
	assert.Equal(t, expectedHash, computedHash)
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name  string
		match bool
		setup func(t *testing.T) (raw string, hash []byte)
	}{
		{
			name:  "returns true for matching token",
			match: true,
			setup: func(t *testing.T) (string, []byte) {
				raw, hash, err := Generate()
				require.NoError(t, err)
				return raw, hash
			},
		},
		{
			name:  "returns false for wrong token",
			match: false,
			setup: func(t *testing.T) (string, []byte) {
				raw1, _, err := Generate()
				require.NoError(t, err)
				_, hash2, err := Generate()
				require.NoError(t, err)
				return raw1, hash2
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, hash := tt.setup(t)
			assert.Equal(t, tt.match, Compare(raw, hash))
		})
	}
}

func TestGenerate_RawIsHexEncoded256Bit(t *testing.T) {
	raw, _, err := Generate()
	require.NoError(t, err)

	// 32 bytes hex-encoded = 64 hex characters
	assert.Len(t, raw, 64)

	// Verify it is valid hex
	decoded, err := hex.DecodeString(raw)
	require.NoError(t, err)
	assert.Len(t, decoded, 32)
}

func TestGenerate_HashIs32Bytes(t *testing.T) {
	_, hash, err := Generate()
	require.NoError(t, err)
	assert.Len(t, hash, 32)
}
