package main

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

type TreeProvider interface {
	RootDir() string
	GetTree() (*Tree, error)
}

type Tree struct {
	Root  *Node
	Index map[string]*Node
}

type Node struct {
	ID       string
	Title    string
	Path     string
	Children []*Node
	Parent   *Node
}

func (n *Node) Walk(fn func(node *Node, depth int) error) error {
	return n.walkRec(0, fn)
}

func (n *Node) walkRec(depth int, fn func(node *Node, depth int) error) error {
	if n != nil {
		err := fn(n, depth)
		if err != nil {
			return err
		}

		for _, child := range n.Children {
			err = child.walkRec(depth+1, fn)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *Node) Read() (string, error) {
	bs, err := ioutil.ReadFile(n.Path)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

func LoadTree(dir string) (*Tree, error) {
	fsys := os.DirFS(dir)

	nodes, err := LoadTreeNodesFromDir(fsys, dir, ".", nil)
	if err != nil {
		return nil, err
	}

	if len(nodes) > 0 {
		root := nodes[0]

		tree := &Tree{
			Root:  root,
			Index: make(map[string]*Node),
		}
		root.Walk(func(node *Node, _ int) error {
			node.ID = strings.TrimPrefix(node.Path, dir)
			node.ID = strings.ReplaceAll(node.ID, "\\", "/")
			tree.Index[node.ID] = node
			return nil
		})

		return tree, nil
	}

	return nil, fmt.Errorf("no *.html files in \"%s\"", dir)
}

func LoadTreeNodesFromDir(fsys fs.FS, parentDir, dir string, parentNode *Node) ([]*Node, error) {
	parentDir = filepath.Clean(filepath.Join(parentDir, dir))
	fsys, err := fs.Sub(fsys, dir)
	if err != nil {
		return nil, err
	}

	pathes, err := fs.Glob(fsys, "*.html")
	if err != nil {
		return nil, err
	}

	var nodes []*Node
	for _, path := range pathes {
		node, err := LoadTreeNode(fsys, parentDir, path, parentNode)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func LoadTreeNode(fsys fs.FS, parentDir, path string, parentNode *Node) (*Node, error) {
	node, err := LoadTreeNodeInfo(fsys, parentDir, path)
	if err != nil {
		return nil, err
	}

	node.Parent = parentNode

	_, filename := filepath.Split(path)
	ext := filepath.Ext(filename)
	if ext != "" {
		filename = strings.TrimSuffix(filename, ext)
	}

	_, err = fs.Stat(fsys, filename)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	} else {
		node.Children, err = LoadTreeNodesFromDir(fsys, parentDir, filename, node)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

func LoadTreeNodeInfo(fsys fs.FS, parentDir, path string) (*Node, error) {
	_, filename := filepath.Split(path)
	filename = strings.TrimSuffix(filename, ".html")
	parts := strings.Split(filename, " ")
	id := parts[len(parts)-1]

	f, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	markup, err := html.Parse(f)
	if err != nil {
		return nil, err
	}

	article := FindHTML(markup, "article")
	if article == nil {
		return nil, fmt.Errorf("no \"article\" node in \"%s\"", path)
	}

	for _, attr := range article.Attr {
		if attr.Key == "data-title" {
			node := &Node{
				ID:    id,
				Path:  filepath.Join(parentDir, path),
				Title: attr.Val,
			}
			return node, nil
		}
	}

	return nil, fmt.Errorf("no \"data-title\" attribute in \"%s\"", path)
}
