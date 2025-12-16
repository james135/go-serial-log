package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"go-serial/aws"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type FileWatcher struct {
	mu sync.Mutex
	v  map[string]bool
}

func (c *FileWatcher) WatchFile(key string) {
	c.mu.Lock()
	c.v[key] = true
	c.mu.Unlock()
}

func (c *FileWatcher) RemoveFile(key string) {
	c.mu.Lock()
	delete(c.v, key)
	c.mu.Unlock()
}

func (c *FileWatcher) CheckFileIsOpen(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.v[key]
	return ok
}

func UploadFiles() error {

	fmt.Printf("Starting upload process (%s)\n", time.Now().UTC().Format(time.RFC3339))

	entries, err := os.ReadDir(STORAGE_DIR)
	if err != nil {
		return err
	}

	for _, entry := range entries {

		fileName := entry.Name()

		fmt.Printf("--- %s ---\n", fileName)

		if entry.IsDir() {
			continue
		}

		path := fmt.Sprintf("%s/%s", STORAGE_DIR, fileName)

		fileIsOpen := fw.CheckFileIsOpen(path)

		if fileIsOpen {
			fmt.Printf("warning - %s is still open\n", fileName)
			continue
		}

		fi, err := entry.Info()
		if err != nil {
			fmt.Printf("warning - Unable to get file info %s (%s)\n", fileName, err)
			continue
		}

		if fi.Size() == 0 {
			if err := os.Remove(path); err != nil {
				fmt.Printf("warning - Unable to delete file from local file system (%s)\n", err)
			}
			continue
		}

		file, err := os.Open(path)
		if err != nil {
			fmt.Printf("warning - Unable to open file %s (%s)\n", fileName, err)
			continue
		}
		defer file.Close()

		fb, err := fileToBytes(file)
		if err != nil {
			fmt.Printf("warning - Unable to convert file %s (%s)\n", fileName, err)
			continue
		}

		if !strings.HasSuffix(fileName, ".gz") {
			fb, err = gZipData(fb)
			if err != nil {
				fmt.Printf("warning - Unable to compress file %s (%s)\n", fileName, err)
				continue
			}
			fileName += ".gz"
		}

		if err := aws.UploadToS3(UPLOAD_BUCKET, fmt.Sprintf("%s/%s", UPLOAD_PATH, fileName), fb); err != nil {
			fmt.Printf("warning - Unable to upload file %s (%s)\n", path, err)
			continue
		}

		if err := os.Remove(path); err != nil {
			fmt.Printf("warning - Unable to delete file from local file system (%s)\n", err)
			continue
		}
	}

	return nil
}

func gZipData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err := gz.Write(data)
	if err != nil {
		return nil, err
	}
	if err = gz.Flush(); err != nil {
		return nil, err
	}
	if err = gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func compressFile(filePath string) error {

	f, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	fGz, err := gZipData(f)
	if err != nil {
		return err
	}

	if err := os.WriteFile(fmt.Sprintf("%s.gz", filePath), fGz, 0644); err != nil {
		return err
	}

	if err := os.Remove(filePath); err != nil {
		return err
	}

	return nil
}

func fileToBytes(file *os.File) ([]byte, error) {

	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(file)
}
