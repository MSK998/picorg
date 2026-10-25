package fswrapper

import (
	"io/fs"
	"os"
	"path/filepath"
)

type FS interface {
	Stat(name string) (os.FileInfo, error)
	MkDirAll(path string, perm os.FileMode) error
	Rename(oldPath string, newPath string) error
	WalkDir(root string, fn fs.WalkDirFunc) error
}

type OsFilesystem struct{}

func (of *OsFilesystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (of *OsFilesystem) MkDirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (of *OsFilesystem) Rename(oldPath string, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (of *OsFilesystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}
