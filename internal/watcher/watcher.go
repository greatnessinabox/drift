package watcher

import (
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Event struct {
	Path      string
	Timestamp time.Time
}

type Watcher struct {
	inner      *fsnotify.Watcher
	root       string
	exclude    []string
	extensions []string
	Events     chan Event
	Errors     chan error
	done       chan struct{}
}

func New(root string, exclude []string, extensions []string) (*Watcher, error) {
	inner, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		inner:      inner,
		root:       root,
		exclude:    exclude,
		extensions: extensions,
		Events:     make(chan Event, 100),
		Errors:     make(chan error, 10),
		done:       make(chan struct{}),
	}

	if err := w.addDirs(root); err != nil {
		inner.Close()
		return nil, err
	}

	go w.loop()

	return w, nil
}

func (w *Watcher) Close() {
	close(w.done)
	w.inner.Close()
}

func (w *Watcher) loop() {
	debounce := make(map[string]*time.Timer)

	for {
		select {
		case <-w.done:
			return
		case event, ok := <-w.inner.Events:
			if !ok {
				return
			}

			if !w.matchesExtension(event.Name) {
				continue
			}

			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			path := event.Name

			if timer, exists := debounce[path]; exists {
				timer.Stop()
			}

			debounce[path] = time.AfterFunc(200*time.Millisecond, func() {
				w.Events <- Event{
					Path:      path,
					Timestamp: time.Now(),
				}
				delete(debounce, path)
			})

		case err, ok := <-w.inner.Errors:
			if !ok {
				return
			}
			w.Errors <- err
		}
	}
}

func (w *Watcher) addDirs(root string) error {
	return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if w.isExcluded(info.Name()) {
				return filepath.SkipDir
			}
			return w.inner.Add(path)
		}
		return nil
	})
}

func (w *Watcher) isExcluded(name string) bool {
	for _, ex := range w.exclude {
		if name == ex {
			return true
		}
	}
	return false
}

func (w *Watcher) matchesExtension(path string) bool {
	for _, ext := range w.extensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
