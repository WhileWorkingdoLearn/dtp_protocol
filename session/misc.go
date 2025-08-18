package session

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateSessionId(min, max int64) (int, error) {
	if min > max {
		return 0, fmt.Errorf("invalid range: %d > %d", min, max)
	}
	span := max - min + 1
	nBig, err := rand.Int(rand.Reader, big.NewInt(span))
	if err != nil {
		return 0, err
	}
	return int(nBig.Int64() + min), nil
}
