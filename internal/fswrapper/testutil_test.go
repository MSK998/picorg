package fswrapper

import (
	"io/fs"
	"strings"
	"testing"
)

func TestNewTestMemFsWithFiles(t *testing.T) {
	mfs := NewTestMemFsWithFiles("photos", 1000)

	info, err := mfs.Stat("photos")
	if err != nil {
		t.Fatalf("stat root: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected root to be a directory")
	}

	var fileCount int
	err = mfs.WalkDir("photos", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			fileCount++
			if !strings.HasSuffix(d.Name(), ".jpg") {
				t.Errorf("expected .jpg suffix, got %s", d.Name())
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if fileCount != 1000 {
		t.Fatalf("expected 1000 files, got %d", fileCount)
	}
}

func BenchmarkNewTestMemFsWithFiles(b *testing.B) {
	b.Run("create", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewTestMemFsWithFiles("photos", 10000)
		}
	})

	b.Run("walk", func(b *testing.B) {
		mfs := NewTestMemFsWithFiles("photos", 10000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var count int
			_ = mfs.WalkDir("photos", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() {
					count++
				}
				return nil
			})
			_ = count
		}
	})
}
