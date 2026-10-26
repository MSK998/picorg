package fswrapper

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func newTestMemFs() *MemFs {
	return &MemFs{files: make(map[string]*MemFile)}
}

func TestMemFs_Stat(t *testing.T) {
	mfs := newTestMemFs()
	mfs.AddFile("photos/01-02-beach.jpg", []byte("data"))

	t.Run("existing file", func(t *testing.T) {
		info, err := mfs.Stat("photos/01-02-beach.jpg")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := info.Name(); got != "01-02-beach.jpg" {
			t.Errorf("want name 01-02-beach.jpg, got %s", got)
		}
		if info.IsDir() {
			t.Error("expected file, got directory")
		}
	})

	t.Run("existing directory", func(t *testing.T) {
		info, err := mfs.Stat("photos")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected directory, got file")
		}
	})

	t.Run("missing path", func(t *testing.T) {
		_, err := mfs.Stat("missing.txt")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func TestMemFs_MkDirAll(t *testing.T) {
	mfs := newTestMemFs()

	t.Run("creates single directory", func(t *testing.T) {
		if err := mfs.MkDirAll("newdir", 0750); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		info, err := mfs.Stat("newdir")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected directory")
		}
	})

	t.Run("creates nested directories", func(t *testing.T) {
		if err := mfs.MkDirAll("a/b/c", 0750); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, path := range []string{"a", "a/b", "a/b/c"} {
			info, err := mfs.Stat(path)
			if err != nil {
				t.Fatalf("stat %s: unexpected error: %v", path, err)
			}
			if !info.IsDir() {
				t.Errorf("expected %s to be a directory", path)
			}
		}
	})
}

func TestMemFs_Rename(t *testing.T) {
	t.Run("rename file", func(t *testing.T) {
		mfs := newTestMemFs()
		mfs.AddFile("old.txt", []byte("data"))

		if err := mfs.Rename("old.txt", "new.txt"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := mfs.Stat("old.txt"); err == nil {
			t.Error("expected old.txt to not exist")
		}
		info, err := mfs.Stat("new.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name() != "new.txt" {
			t.Errorf("want name new.txt, got %s", info.Name())
		}
	})

	t.Run("rename directory with children", func(t *testing.T) {
		mfs := newTestMemFs()
		mfs.AddFile("old/a.txt", []byte("1"))
		mfs.AddFile("old/sub/b.txt", []byte("2"))

		if err := mfs.Rename("old", "new"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := mfs.Stat("old"); err == nil {
			t.Error("expected old to not exist")
		}
		if _, err := mfs.Stat("old/a.txt"); err == nil {
			t.Error("expected old/a.txt to not exist")
		}
		for _, path := range []string{"new", "new/a.txt", "new/sub", "new/sub/b.txt"} {
			if _, err := mfs.Stat(path); err != nil {
				t.Errorf("expected %s to exist: %v", path, err)
			}
		}
	})

	t.Run("missing source", func(t *testing.T) {
		mfs := newTestMemFs()
		if err := mfs.Rename("missing.txt", "new.txt"); err == nil {
			t.Error("expected error for missing source")
		}
	})
}

func TestMemFs_WalkDir(t *testing.T) {
	t.Run("walks in sorted order", func(t *testing.T) {
		mfs := newTestMemFs()
		mfs.AddFile("root/b.txt", []byte("2"))
		mfs.AddFile("root/a.txt", []byte("1"))
		mfs.AddFile("root/sub/c.txt", []byte("3"))

		var paths []string
		err := mfs.WalkDir("root", func(path string, d fs.DirEntry, err error) error {
			paths = append(paths, path)
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []string{"root", "root/a.txt", "root/b.txt", "root/sub", "root/sub/c.txt"}
		if !reflect.DeepEqual(paths, want) {
			t.Errorf("walk order mismatch\ngot:  %v\nwant: %v", paths, want)
		}
	})

	t.Run("skip dir", func(t *testing.T) {
		mfs := newTestMemFs()
		mfs.AddFile("root/sub/c.txt", []byte("3"))
		mfs.AddFile("root/a.txt", []byte("1"))

		var paths []string
		err := mfs.WalkDir("root", func(path string, d fs.DirEntry, err error) error {
			paths = append(paths, path)
			if d.IsDir() && d.Name() == "sub" {
				return fs.SkipDir
			}
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []string{"root", "root/a.txt", "root/sub"}
		if !reflect.DeepEqual(paths, want) {
			t.Errorf("walk order mismatch\ngot:  %v\nwant: %v", paths, want)
		}
	})

	t.Run("missing root", func(t *testing.T) {
		mfs := newTestMemFs()
		err := mfs.WalkDir("missing", func(path string, d fs.DirEntry, err error) error {
			return nil
		})
		if err == nil {
			t.Error("expected error for missing root")
		}
	})
}

func TestOsFilesystem(t *testing.T) {
	tmpDir := t.TempDir()
	osFS := &OsFilesystem{}

	t.Run("MkDirAll and Stat", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "nested", "dir")
		if err := osFS.MkDirAll(dir, 0750); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		info, err := osFS.Stat(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected directory")
		}
	})

	t.Run("WalkDir", func(t *testing.T) {
		if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("1"), 0644); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var found bool
		err := osFS.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
			if d.Name() == "a.txt" {
				found = true
			}
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !found {
			t.Error("expected to find a.txt")
		}
	})
}
