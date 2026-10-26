package app

import (
	"regexp"
	"testing"

	"github.com/MSK998/picorg/internal/fswrapper"
)

func TestGetPlannedFileMoves(t *testing.T){
	fs := fswrapper.NewMemFs()
	fs.AddFile("/toor/01-01-test.jpg", make([]byte, 0))

	moves := getPlannedFileMoves(fs, "/toor", regexp.MustCompile(`^(\d{2})-(\d{2})`))

	if len(moves) != 1 {
		t.Errorf("Expected 1 move got: %d", len(moves))
	}
}

func TestGetPlannedFileMovesConcurrent(t *testing.T){
	fs := fswrapper.NewMemFs()
	fs.AddFile("/toor/01-01-test.jpg", make([]byte, 0))

	moves := getPlannedFileMovesConcurrent(fs, "/toor", regexp.MustCompile(`^(\d{2})-(\d{2})`))

	if len(moves) != 1 {
		t.Errorf("Expected 1 move got: %d", len(moves))
	}
}

func TestGetPlannedFileMovesBatchConcurrent(t *testing.T){
	fs := fswrapper.NewMemFs()
	fs.AddFile("/toor/01-01-test.jpg", make([]byte, 0))

	moves := getPlannedFileMovesBatchConcurrent(fs, "/toor", regexp.MustCompile(`^(\d{2})-(\d{2})`))

	if len(moves) != 1 {
		t.Errorf("Expected 1 move got: %d", len(moves))
	}
}

func BenchmarkGetPlannedFileMoves(b *testing.B) {
	fs := fswrapper.NewTestMemFsWithFiles("/toor", 10000)
	regex := regexp.MustCompile(`^(\d{2})-(\d{2})`)

	b.Run("sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = getPlannedFileMoves(fs, "/toor", regex)
		}
	})

	b.Run("concurrent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = getPlannedFileMovesConcurrent(fs, "/toor", regex)
		}
	})

	b.Run("concurrent_batch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = getPlannedFileMovesBatchConcurrent(fs, "/toor", regex)
		}
	})
}

