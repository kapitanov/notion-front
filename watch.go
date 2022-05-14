package main

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	sourceDir, destDir string
	mutex              sync.Mutex
	tree               *Tree
	ticker             *time.Ticker
	pendingUpdate      *atomic.Value
}

func NewWatcher(sourceDir, destDir string) (*Watcher, error) {
	w := &Watcher{
		sourceDir:     sourceDir,
		destDir:       destDir,
		ticker:        time.NewTicker(5 * time.Second),
		pendingUpdate: &atomic.Value{},
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fsw.Add(sourceDir)

	var (
		pending interface{} = "P"
		ready   interface{} = "R"
	)

	go func() {
		for range fsw.Events {
			w.pendingUpdate.Swap(pending)
		}
	}()

	go func() {
		for range w.ticker.C {
			val := w.pendingUpdate.Swap(ready)
			if val == pending {
				log.Printf("Changes detected")
				_ = w.Update()
			}
		}
	}()

	err = w.Update()
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Watcher) RootDir() string {
	return w.destDir
}

func (w *Watcher) GetTree() (*Tree, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.tree, nil
}

func (w *Watcher) Update() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	err := TransformTree(w.sourceDir, w.destDir)
	if err != nil {
		return err
	}

	tree, err := LoadTree(w.destDir)
	if err != nil {
		return err
	}

	w.tree = tree
	return nil
}
