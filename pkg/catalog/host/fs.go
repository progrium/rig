package host

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/progrium/rig/pkg/entity"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
)

func FilepathNode(name, path string) *node.Raw {
	fi, err := fs.Stat(os.DirFS(path), ".")
	if err != nil {
		log.Println(err)
		return nil
	}
	if fi.IsDir() {
		return dirNode(name, path)
	}
	return fileNode(name, path)
}

type File struct {
	Path string
}

type Directory struct {
	Path string
}

func dirNode(name, path string) *node.Raw {
	return node.New(name, node.Attrs{
		"desc": " ",
		"view": "host.Directory",
	}, &Directory{Path: path})
}

func fileNode(name, path string) *node.Raw {
	return node.New(name, node.Attrs{
		"icon": "symbol-file",
	}, &File{Path: path})
}

func (d *Directory) Nodes(com manifold.Node) (nodes entity.Nodes) {
	entries, err := fs.ReadDir(os.DirFS(d.Path), ".")
	if err != nil {
		log.Println(err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			nodes = append(nodes, dirNode(entry.Name(), filepath.Join(d.Path, entry.Name())))
		}
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			nodes = append(nodes, fileNode(entry.Name(), filepath.Join(d.Path, entry.Name())))
		}
	}
	return
}
