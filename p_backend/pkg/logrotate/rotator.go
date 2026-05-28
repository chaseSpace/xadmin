package logrotate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Rotator is an io.WriteCloser that writes to a file and rotates it
// based on max size. Archived files are named as name_yyyymmdd.log.
type Rotator struct {
	Filename   string // current log file path
	MaxSize    int    // max size in MB before rotation
	MaxBackups int    // max number of archived files to keep

	mu   sync.Mutex
	file *os.File
	size int64
}

func (r *Rotator) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file == nil {
		if err = r.open(); err != nil {
			return 0, err
		}
	}

	if r.size+int64(len(p)) > int64(r.MaxSize)*1024*1024 {
		if err = r.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = r.file.Write(p)
	r.size += int64(n)
	return n, err
}

func (r *Rotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

func (r *Rotator) open() error {
	if err := os.MkdirAll(filepath.Dir(r.Filename), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(r.Filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	r.file = f
	r.size = info.Size()
	return nil
}

func (r *Rotator) rotate() error {
	if r.file != nil {
		r.file.Close()
		r.file = nil
	}

	// archive name: name_yyyymmdd.log (with sequence if exists)
	dir := filepath.Dir(r.Filename)
	ext := filepath.Ext(r.Filename)
	base := strings.TrimSuffix(filepath.Base(r.Filename), ext)
	date := time.Now().Format("20060102")

	archiveName := filepath.Join(dir, fmt.Sprintf("%s_%s%s", base, date, ext))
	// if archive already exists, add sequence
	if _, err := os.Stat(archiveName); err == nil {
		for i := 1; ; i++ {
			candidate := filepath.Join(dir, fmt.Sprintf("%s_%s_%d%s", base, date, i, ext))
			if _, err := os.Stat(candidate); os.IsNotExist(err) {
				archiveName = candidate
				break
			}
		}
	}

	os.Rename(r.Filename, archiveName)
	r.cleanup(dir, base, ext)
	return r.open()
}

func (r *Rotator) cleanup(dir, base, ext string) {
	if r.MaxBackups <= 0 {
		return
	}
	pattern := filepath.Join(dir, base+"_*"+ext)
	matches, _ := filepath.Glob(pattern)
	if len(matches) <= r.MaxBackups {
		return
	}
	sort.Sort(sort.Reverse(sort.StringSlice(matches)))
	for _, f := range matches[r.MaxBackups:] {
		os.Remove(f)
	}
}
