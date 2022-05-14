package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var sourceDir, destDir, httpEndpoint string

	flag.StringVar(&sourceDir, "i", "", "path to input directory")
	flag.StringVar(&destDir, "c", "", "path to cache directory")
	flag.StringVar(&httpEndpoint, "p", "0.0.0.0:8000", "http endpoint")

	flag.Parse()

	if sourceDir == "" {
		panic("missing input directory")
	}

	if destDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		destDir = filepath.Join(wd, "cache")
	}

	log.Printf("source directory: %s", sourceDir)
	log.Printf("cache directory: %s", destDir)
	log.Printf("http endpoint: %s", httpEndpoint)

	w, err := NewWatcher(sourceDir, destDir)
	if err != nil {
		panic(err)
	}

	err = ServeHTML(httpEndpoint, w)
	if err != nil {
		panic(err)
	}
}
