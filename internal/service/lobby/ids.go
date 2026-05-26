package lobby

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

func newGameID() (string, error) {
	// 16 bytes => 32 hex chars
	return randomHex(16)
}

func newPlayerID() (string, error) {
	// 16 bytes => 32 hex chars
	return randomHex(16)
}

func newJoinCode() (string, error) {
	// 6 chars from [A-Z0-9]
	const length = 6
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := randInt(len(alphabet))
		if err != nil {
			return "", err
		}
		b[i] = alphabet[n]
	}
	return string(b), nil
}

func randomHex(nBytes int) (string, error) {
	buf := make([]byte, nBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func randInt(max int) (int, error) {
	// crypto/rand-safe random int in [0, max)
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(nBig.Int64()), nil
}
