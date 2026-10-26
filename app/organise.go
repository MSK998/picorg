package app

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"

	"github.com/MSK998/picorg/internal/fswrapper"
)

type Options struct {
	IsDryRun bool
	Dir string
	Regex *regexp.Regexp
}

func (a *App) Organise(picorg *Options) error {	
	if _, err := a.Fs.Stat(picorg.Dir); err != nil {
		return err
	}

	plannedFileMoves := getPlannedFileMoves(a.Fs, picorg.Dir, picorg.Regex)

	newDirs, err := getAllNewDirs(plannedFileMoves)
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

	err = makeAllDirs(a.Fs, newDirs)
	if err != nil {
		return err
	}

	err = moveFiles(a.Fs, plannedFileMoves)
	if err != nil {
		return err
	}
	return nil
}

// A sequential, single threaded implementation
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

func worker(jobs <-chan string, movements map[string]string, lock *sync.Mutex, wg *sync.WaitGroup, regex *regexp.Regexp){
	defer wg.Done()
	for job := range jobs {
		pp := getPlannedPath(job, regex)
		lock.Lock()
		movements[job] = pp
		lock.Unlock()
	}
}

// A channel based concurrent implementation ~45% slower than the sequential (based on benchmarks)
func getPlannedFileMovesConcurrent(filesystem fswrapper.FS, root string, regex *regexp.Regexp) map[string]string {
	jobs := make(chan string)
	fileMap := make(map[string]string)
	var lock sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker(jobs, fileMap, &lock, &wg, regex)
	}
	
	filesystem.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && path != root {
			return filepath.SkipDir
		}
		if path == root {
			return nil
		}
		jobs <- path
		
		return nil
	})

	close(jobs)
	wg.Wait()

	return fileMap
}

// A batched concurrent implementation, splits the work evenly between the set number of goroutines
// this approach avoids races, or locks by only giving the goroutines access to chunks of the result.
func getPlannedFileMovesBatchConcurrent(filesystem fswrapper.FS, root string, regex *regexp.Regexp) map[string]string {
	files := make([]string, 0)
	var wg sync.WaitGroup
	
	filesystem.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && path != root {
			return filepath.SkipDir
		}
		if path == root {
			return nil
		}

		files = append(files, path)
		return nil
	})

	planned := make([]string, len(files))
	numProcs := runtime.GOMAXPROCS(0)
	chunkSize := (len(files) + numProcs - 1) / numProcs
	for start := 0; start < len(files); start += chunkSize {
		end := start + chunkSize
		if end > len(files) {
			end = len(files)
		}
		wg.Add(1)
		go func(start int, end int){
			defer wg.Done()
			for i := start; i < end; i++ {
				planned[i] = getPlannedPath(files[i], regex)
			}
		}(start, end)
	}

	wg.Wait()

	plannedMoves := make(map[string]string, len(planned))
	for i, p := range files {
		plannedMoves[p] = planned[i]
	}
	
	return plannedMoves
}

func getPlannedPath(oldPath string, r *regexp.Regexp) string {
	filename := filepath.Base(oldPath)
	yM := r.FindStringSubmatch(filename)
	if len(yM) < 3 {
		return oldPath
	}
	return path.Join(filepath.Dir(oldPath), yM[1], yM[2], filename)
}

func getAllNewDirs(plannedMoves map[string]string) (map[string]int, error) {
	newDirs := make(map[string]int, 0)
	for _, n := range plannedMoves {
		base := filepath.Dir(n)
		newDirs[base] += 1
	}

	return newDirs, nil
}

func makeAllDirs(filesystem fswrapper.FS, dirs map[string]int) error {
	for v := range dirs {
		if _, err := filesystem.Stat(v); err == nil {
			continue
		}

		err := filesystem.MkDirAll(v, 0750)
		if err != nil {
			return err
		}
	}
	return nil
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