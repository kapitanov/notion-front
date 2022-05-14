package main

import (
	"bytes"

	"golang.org/x/net/html"
)

func FindHTML(node *html.Node, tag string) *html.Node {
	root:=node

	node = root
	for node != nil {
		if node.Type == html.ElementNode && node.Data == tag {
			return node
		}

		node = node.NextSibling
	}

	node = root
	for node != nil {
		result := FindHTML(node.FirstChild, tag)
		if result != nil {
			return result
		}

		node = node.NextSibling
	}

	return nil
}

func RenderHTML(node *html.Node) ([]byte, error) {
	var w bytes.Buffer
	err := html.Render(&w, node)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func TraverseHTML(node *html.Node, visit func(node *html.Node) error) error {
	for node != nil {
		err := visit(node)
		if err != nil {
			return err
		}

		err = TraverseHTML(node.FirstChild, visit)
		if err != nil {
			return err
		}

		node = node.NextSibling
	}

	return nil
}
