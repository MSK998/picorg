package app

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/MSK998/picorg/internal/fswrapper"
)

type Picorg struct {
	IsDryRun bool
	Dir string
	Regex *regexp.Regexp
	plannedFileMoves map[string]string
}

func (a *App) Picorg(picorg *Picorg) error {
	var err error

	if _, err = a.Fs.Stat(picorg.Dir); err != nil {
		return err
	}

	picorg.plannedFileMoves = getPlannedFileMoves(a.Fs, picorg.Dir, picorg.Regex)

	newDirs, err := getAllNewDirs(picorg.plannedFileMoves, picorg.Regex)
	if err != nil {
		return err
	}

	if picorg.IsDryRun {
		fmt.Fprintln(a.Out, "The following directories will be created:")
		for d, c := range newDirs {
			_, err := fmt.Fprintf(a.Out, "\t%s | %d file(s)\n", d, c)
			if err != nil {
				return err
			}
		}
		return nil
	}


	err = makeAllDirs(newDirs)
	if err != nil {
		return err
	}

	err = moveFiles(a.Fs, picorg.plannedFileMoves)
	if err != nil {
		return err
	}
	return nil
}

func getAllNewDirs(plannedMoves map[string]string, regex *regexp.Regexp) (map[string]int, error) {
	newDirs := make(map[string]int, 0)
	for _, n := range plannedMoves {
		base := filepath.Dir(n)
		newDirs[base] += 1
	}

	return newDirs, nil
}

func makeAllDirs(dirs map[string]int) error {
	for v := range dirs {
		if _, err := os.Stat(v); err == nil {
			continue
		}

		err := os.MkdirAll(v, 0750)
		if err != nil {
			return err
		}
	}
	return nil
}

func getPlannedFileMoves(filesystem fswrapper.FS, root string, regex *regexp.Regexp) map[string]string {
	fileMap := make(map[string]string)
	filesystem.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && path != root {
			return filepath.SkipDir
		}
		if path == root {
			return nil
		}
		pp := getPlannedPath(path, regex)
		fileMap[path] = pp
		return nil
	})
	return fileMap
}

func getPlannedPath(oldPath string, r *regexp.Regexp) string {
	filename := filepath.Base(oldPath)
	yM := r.FindStringSubmatch(filename)
	if len(yM) < 3 {
		return oldPath
	}
	return path.Join(filepath.Dir(oldPath), yM[1], yM[2], filename)
}

func moveFiles(filesystem fswrapper.FS, files map[string]string) error {
	for o, n := range files {
		err := filesystem.Rename(o, n)
		if err != nil {
			return err
		}
	}

	return nil
}