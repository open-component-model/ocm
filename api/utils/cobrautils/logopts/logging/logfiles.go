package logging

import (
	"os"
	"sync"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"slices"
)

type LogFile struct {
	count int
	path  string
	file  vfs.File
	fs    vfs.FileSystem
}

func (l *LogFile) File() vfs.File {
	return l.file
}

func (l *LogFile) Close() error {
	lock.Lock()
	defer lock.Unlock()

	l.count--
	if l.count == 0 {
		i := slices.Index(logFiles, l)
		if i >= 0 {
			logFiles.DeleteIndex(i)
		}
		return l.file.Close()
	}
	return nil
}

var (
	lock     sync.Mutex
	logFiles sliceutils.Slice[*LogFile]
)

func CloseLogFiles() {
	lock.Lock()
	defer lock.Unlock()

	for _, f := range logFiles {
		f.file.Close()
	}
	logFiles = nil
}

func LogFileFor(path string, fs vfs.FileSystem) (*LogFile, error) {
	lock.Lock()
	defer lock.Unlock()

	path, f := getLogFileFor(path, fs)
	if f != nil {
		f.count++
		return f, nil
	}
	lf, err := fs.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	f = &LogFile{path: path, file: lf, fs: fs, count: 1}
	logFiles.Add(f)
	return f, nil
}

func getLogFileFor(path string, fs vfs.FileSystem) (string, *LogFile) {
	path, err := vfs.Canonical(fs, path, false)
	if err != nil {
		return path, nil
	}

	for _, f := range logFiles {
		if f.path == path && f.fs == fs {
			return path, f
		}
	}
	return path, nil
}

func GetLogFileFor(path string, fs vfs.FileSystem) *LogFile {
	lock.Lock()
	defer lock.Unlock()

	_, l := getLogFileFor(path, fs)
	return l
}
