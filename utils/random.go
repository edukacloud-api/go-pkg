package utils

import (
	"github.com/oklog/ulid"
	"math/rand"
	"time"
)

func GenerateThreadID() string {
	t := time.Now()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	uniqueID := ulid.MustNew(ulid.Timestamp(t), entropy)
	return uniqueID.String()
}
