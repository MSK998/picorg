package fswrapper

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type MemFs struct {
	files map[string]*MemFile
}

func NewMemFs() *MemFs{
	return &MemFs{
		files: make(map[string]*MemFile),
	}
}

func (mfs *MemFs) AddFile(path string, content []byte) {
	fp := filepath.Clean(path)
	mfs.files[fp] = &MemFile{
		name:    filepath.Base(fp),
		content: content,
		mode:    0644,
		modTime: time.Now(),
	}
	mfs.ensureDirs(filepath.Dir(fp))
}

func (mfs *MemFs) ensureDirs(path string) {
	dir := filepath.Clean(path)
	if dir == "." || dir == "/" {
		return
	}
	if _, ok := mfs.files[dir]; ok {
		return
	}

	mfs.files[dir] = &MemFile{
		name:    filepath.Base(dir),
		mode:    0750 | fs.ModeDir,
		modTime: time.Now(),
	}
	mfs.ensureDirs(filepath.Dir(dir))
}

func (mfs *MemFs) walk(path string, fn fs.WalkDirFunc) error {
	rootMemFile, ok := mfs.files[path]
	if !ok {
		return nil
	}

	rootDirEntry := fs.FileInfoToDirEntry(rootMemFile)
	if err := fn(path, rootDirEntry, nil); err != nil {
		if errors.Is(err, fs.SkipDir) {
			return nil
		}
		return err
	}

	if !rootMemFile.IsDir(){
		return nil
	}

	prefix := path + string(filepath.Separator)
	seen := make(map[string]struct{})
	var children []string

	for memFile := range mfs.files {
		if strings.HasPrefix(memFile, prefix){
			rel := strings.TrimPrefix(memFile, prefix)
			extraParts := strings.Split(rel, string(filepath.Separator))
			child := extraParts[0]
			childPath := filepath.Join(path, child)

			if _, ok := seen[childPath]; !ok {
				seen[childPath] = struct{}{}
				children = append(children, childPath)
			}
		}
	}

	sort.Strings(children)

	for _, child := range children {
		if err := mfs.walk(child, fn); err != nil {
			return err
		}
	}
	return nil
}

func (mfs *MemFs) Stat(name string) (os.FileInfo, error) {
	fp := filepath.Clean(name)
	file, ok := mfs.files[fp]
	if !ok {
		return nil, &os.PathError{Op: "stat", Path: name, Err: errors.New("file not found")}
	}

	return file, nil
}

func (mfs *MemFs) MkDirAll(path string, perm os.FileMode) error {
	fp := filepath.Clean(path)
	mfs.ensureDirs(path)
	if dir := mfs.files[fp]; dir != nil {
		dir.mode = perm | fs.ModeDir
	}
	return nil
}

func (mfs *MemFs) Rename(oldPath string, newPath string) error {
	ofp := filepath.Clean(oldPath)
	nfp := filepath.Clean(newPath)

	file, ok := mfs.files[ofp]

	if !ok {
		return &fs.PathError{Op: "rename", Path: ofp, Err: errors.New("file not found")}
	}

	mfs.ensureDirs(filepath.Dir(nfp))

	file.name = filepath.Base(nfp)
	mfs.files[nfp] = file
	delete(mfs.files, ofp)

	if file.IsDir() {
		for pathToFile := range mfs.files {
			if pathToFile == ofp {
				continue
			}
			if strings.HasPrefix(pathToFile, ofp+string(filepath.Separator)) {
				rel := strings.TrimPrefix(pathToFile, oldPath+string(filepath.Separator))
				newChild := filepath.Join(nfp, rel)
				mfs.files[newChild] = mfs.files[pathToFile]
				delete(mfs.files, pathToFile)
			}
		}
	}
	return nil
}

func (mfs *MemFs) WalkDir(root string, fn fs.WalkDirFunc) error {
	root = filepath.Clean(root)

	if _, ok := mfs.files[root]; !ok {
		return &os.PathError{Op: "walk", Path: root, Err: errors.New("root not found")}
	}

	return mfs.walk(root, fn)
}

type MemFile struct {
	name    string
	content []byte
	mode    fs.FileMode
	modTime time.Time
}

func (mf *MemFile) Name() string {
	return mf.name
}

func (mf *MemFile) Size() int64 {
	return int64(len(mf.content))
}

func (mf *MemFile) Mode() os.FileMode {
	return mf.mode
}

func (mf *MemFile) IsDir() bool {
	return mf.mode.IsDir()
}

func (mf *MemFile) ModTime() time.Time {
	return mf.modTime
}

func (mf *MemFile) Sys() any {
	return nil
}
