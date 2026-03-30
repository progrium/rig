package tree

import (
	"bytes"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
)

var startTime = time.Now()

type FileReader interface {
	ReadFile() ([]byte, error)
}

type statCache struct {
	fileInfo
	lastTime time.Time
}

type FS struct {
	statCache map[string]statCache

	manifold.Component
}

func (fsys *FS) resolvePath(name string) (manifold.Node, bool) {
	if name == "." {
		return fsys.Object(), true
	}
	parts := strings.Split(strings.TrimLeft(name, "/."), "/")
	obj := fsys.Object()
	for _, name := range parts {
		n, ok := obj.Objects().FindByName(name)
		if !ok {
			return nil, false
		}
		obj = n
	}
	return obj, true
}

func (fsys *FS) OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error) {
	return fsys.Open(name)
}

func (fsys *FS) Open(name string) (fs.File, error) {
	// log.Println("open:", name)
	n, ok := fsys.resolvePath(name)
	if !ok {
		return nil, fs.ErrNotExist
	}
	var buf bytes.Buffer
	fr := node.Get[FileReader](n)
	if fr != nil {
		b, _ := fr.ReadFile()
		buf = *bytes.NewBuffer(b)
	}
	return &fileHandle{
		ReadCloser: io.NopCloser(&buf),
		Name:       name,
		FS:         fsys,
	}, nil
}

func (fsys *FS) stat(name string) (*fileInfo, error) {
	// log.Println("stat:", name)
	if fsys.statCache == nil {
		fsys.statCache = make(map[string]statCache)
	}
	fi, found := fsys.statCache[name]
	if found && time.Since(fi.lastTime).Milliseconds() < 1000 {
		return &fi.fileInfo, nil
	}

	n, ok := fsys.resolvePath(name)
	if !ok {
		return nil, fs.ErrNotExist
	}

	size := n.Objects().Count()

	fr := node.Get[FileReader](n)
	if fr != nil {
		data, _ := fr.ReadFile()
		size = len(data)
	}

	isDir := n.Objects().Count() > 0
	i := &fileInfo{
		name:    n.Name(),
		size:    int64(size),
		mode:    0644,
		modTime: startTime.Unix(),
		isDir:   isDir,
	}
	if isDir {
		i.mode = 0755
	}

	fsys.statCache[name] = statCache{
		fileInfo: *i,
		lastTime: time.Now(),
	}
	return i, nil
}

func (fsys *FS) Stat(name string) (fs.FileInfo, error) {
	return fsys.stat(name)
}

func (fsys *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	// log.Println("readdir:", name)
	n, ok := fsys.resolvePath(name)
	if !ok {
		return nil, fs.ErrNotExist
	}

	var out []fs.DirEntry
	for _, child := range n.Objects().Nodes() {
		info, err := fsys.stat(filepath.Join(name, child.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, info)
	}
	return out, nil
}

type fileHandle struct {
	io.ReadCloser
	Name string
	FS   *FS
}

func (f *fileHandle) Stat() (fs.FileInfo, error) {
	return f.FS.Stat(f.Name)
}

func (f *fileHandle) ReadDir(n int) ([]fs.DirEntry, error) {
	return f.FS.ReadDir(f.Name)
}

type fileInfo struct {
	name    string
	size    int64
	mode    uint
	modTime int64
	isDir   bool
}

func (i *fileInfo) Name() string       { return i.name }
func (i *fileInfo) Size() int64        { return i.size }
func (i *fileInfo) ModTime() time.Time { return time.Unix(i.modTime, 0) }
func (i *fileInfo) IsDir() bool        { return i.isDir }
func (i *fileInfo) Sys() any           { return nil }
func (i *fileInfo) Mode() fs.FileMode {
	if i.IsDir() {
		return fs.FileMode(i.mode) | fs.ModeDir
	}
	return fs.FileMode(i.mode)
}

// these allow it to act as DirInfo as well
func (i *fileInfo) Info() (fs.FileInfo, error) {
	return i, nil
}
func (i *fileInfo) Type() fs.FileMode {
	return i.Mode()
}
