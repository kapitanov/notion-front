package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

func TransformTree(sourceDir, destDir string) error {
	_, err := os.Stat(destDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else {
		err = os.RemoveAll(destDir)
		if err != nil {
			return err
		}
	}

	return TransformTreeDir(sourceDir, destDir, 0)
}

func TransformTreeDir(sourceDir, destDir string, depth int) error {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			err = TransformTreeDir(
				filepath.Join(sourceDir, entry.Name()),
				filepath.Join(destDir, TransformFileName(entry.Name())),
				depth+1)
			if err != nil {
				return err
			}
		} else {
			err = TransformTreeFile(filepath.Join(sourceDir, entry.Name()), destDir, depth)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func TransformTreeFile(sourcePath, destDir string, depth int) error {
	destPath := GenerateTreeFilePath(sourcePath, destDir, depth)

	ext := filepath.Ext(sourcePath)
	if ext != ".html" {
		return CopyFile(sourcePath, destPath, nil)
	}

	return CopyFile(sourcePath, destPath, TransformHTMLContent)
}

func GenerateTreeFilePath(sourcePath, destDir string, depth int) string {
	filename := filepath.Base(sourcePath)
	ext := filepath.Ext(sourcePath)
	filename = strings.TrimSuffix(filename, ext)

	// if ext == ".html" {
	// 	parentDirName := filepath.Base(filepath.Dir(sourcePath))
	// 	if parentDirName == filename {
	// 		filename = "index.html"
	// 	} else {
	// 		filename = TransformFileName(filename) + ext
	// 	}
	// } else {
	// 	filename = TransformFileName(filename) + ext
	// }

	filename = TransformFileName(filename) + ext

	return filepath.Join(destDir, filename)
}

func TransformFileName(filename string) string {
	parts := strings.Split(filename, " ")
	return parts[len(parts)-1]
}

func CopyFile(sourcePath, destPath string, transform func(input []byte) ([]byte, error)) error {
	err := os.MkdirAll(filepath.Dir(destPath), os.ModeDir)
	if err != nil {
		return err
	}

	sourceFile, err := os.OpenFile(sourcePath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer func() {
		_ = sourceFile.Close()
	}()

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer func() {
		_ = destFile.Close()
	}()

	if transform != nil {
		buffer, err := ioutil.ReadAll(sourceFile)
		if err != nil {
			return err
		}

		buffer, err = transform(buffer)
		if err != nil {
			return err
		}

		_, err = destFile.Write(buffer)
		if err != nil {
			return err
		}
	} else {
		_, err = io.Copy(destFile, sourceFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func TransformHTMLContent(input []byte) ([]byte, error) {
	markup, err := html.Parse(bytes.NewBuffer(input))
	if err != nil {
		return nil, err
	}

	err = TraverseHTML(markup, func(node *html.Node) error {
		TransformHTMLNode(node)
		return nil
	})
	if err != nil {
		return nil, err
	}

	markup = FindHTML(markup, "html")
	head := FindHTML(markup, "head")
	title := FindHTML(head, "title")
	if title == nil {
		return nil, fmt.Errorf("no \"title\" node found")
	}

	body := FindHTML(markup, "body")
	article := FindHTML(body, "article")
	if article == nil {
		return nil, fmt.Errorf("no \"article\" node found")
	}

	article.Attr = append(markup.Attr, html.Attribute{
		Key: "data-title",
		Val: title.FirstChild.Data,
	})

	return RenderHTML(article)
}

func TransformHTMLNode(node *html.Node) {
	if node.Type != html.ElementNode {
		return
	}

	for i := range node.Attr {
		if node.Attr[i].Key == "href" || node.Attr[i].Key == "src" {
			node.Attr[i].Val = TransformHref(node.Attr[i].Val)
		}
	}
}

func TransformHref(href string) string {
	u, err := url.Parse(href)
	if err != nil || u.Scheme != "" {
		return href
	}

	segments := strings.Split(href, "/")
	for i := range segments {
		segment := segments[i]
		segment, _ = url.PathUnescape(segment)

		ext := filepath.Ext(segment)
		filename := strings.TrimSuffix(segment, ext)
		parts := strings.Split(filename, " ")
		segment = parts[len(parts)-1] + ext

		segments[i] = segment
	}

	href = strings.Join(segments, "/")
	return href
}
