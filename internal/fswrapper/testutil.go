package fswrapper

import (
	"fmt"
	"math/rand"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// NewTestMemFsWithFiles creates a new MemFs populated with count files under
// root. Files are named using the convention YY-MM-img-<random>.jpg where YY is
// a year from 24-26, MM is a month from 01-12 and <random> is a short random
// suffix to keep names unique.
//
// It is intended for tests and benchmarks only.
func NewTestMemFsWithFiles(root string, count int) *MemFs {
	mfs := NewMemFs()
	mfs.MkDirAll(root, 0750)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < count; i++ {
		year := 24 + rng.Intn(3)
		month := 1 + rng.Intn(12)
		suffix := randomSuffix(rng, 8)

		name := fmt.Sprintf("%02d-%02d-img-%s.jpg", year, month, suffix)
		path := root + "/" + name
		mfs.AddFile(path, []byte("dummy"))
	}

	return mfs
}

func randomSuffix(rng *rand.Rand, length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = alphabet[rng.Intn(len(alphabet))]
	}
	return string(b)
}
