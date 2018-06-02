package fs

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Dir struct {
	dir   string
	files map[string]bool
}

func New(dir string) (*Dir, error) {
	return &Dir{dir: dir, files: map[string]bool{}}, nil
}

func (t *Dir) AddReader(name string, r io.Reader) error {
	t.files[name] = true
	path := filepath.Join(t.dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create dir for file: %s", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("open file for write: %s", err)
	}
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("copy data to file: %s", err)
	}
	return f.Close()
}

func (t *Dir) AddBytes(name string, data []byte) error {
	return t.AddReader(name, bytes.NewReader(data))
}

func (t *Dir) AddFile(name, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	return t.AddReader(name, file)
}

func (t *Dir) AddURL(name, url string) error {
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Error while downloading %s - %s", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Non 200 status code: %d: %s", response.StatusCode, url)
	}
	return t.AddReader(name, response.Body)
}

func (t *Dir) Exists(name string) bool {
	t.files[name] = true
	_, err := os.Stat(filepath.Join(t.dir, name))
	return err == nil
}

func (t *Dir) Writer(name string) (io.WriteCloser, error) {
	t.files[name] = true
	path := filepath.Join(t.dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create dir for file: %s", err)
	}
	return os.Create(path)
}

func (t *Dir) DeleteOld() error {
	return filepath.Walk(t.dir, func(path string, info os.FileInfo, err error) error {
		name, err := filepath.Rel(t.dir, path)
		if err != nil {
			return err
		}
		if name == "." || t.files[name] || info.IsDir() {
			return nil
		}
		return os.Remove(path)
	})
}
