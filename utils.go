package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"go-serial/aws"
	"io"
	"os"
	"sort"
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

var fw FileWatcher = FileWatcher{
	v: map[string]bool{},
}

func UploadFiles() error {

	fmt.Printf("Starting upload process (%s)\n", time.Now().UTC().Format(time.RFC3339))

	entries, err := os.ReadDir(STORAGE_DIR)
	if err != nil {
		return err
	}

	type DataHolder struct {
		Date string
		Data []byte
		Path string
	}

	data := map[string][]DataHolder{}

	for _, entry := range entries {

		fileName := entry.Name()

		if entry.IsDir() {
			continue
		}

		path := fmt.Sprintf("%s/%s", STORAGE_DIR, fileName)

		fileIsOpen := fw.CheckFileIsOpen(path)

		if fileIsOpen {
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

		group := strings.Split(fileName, "_")[0]
		date := strings.SplitAfterN(fileName, "_", 2)[1]

		if strings.HasSuffix(fileName, ".gz") {
			date = strings.Split(date, ".")[0]
		} else {
			fb, err = gZipData(fb)
			if err != nil {
				fmt.Printf("warning - Unable to compress file %s (%s)\n", fileName, err)
				continue
			}
		}

		data[group] = append(data[group], DataHolder{Date: date, Path: path, Data: fb})
	}

	for group, dhList := range data {

		if len(dhList) == 0 {
			continue
		}

		sort.Slice(dhList, func(i, j int) bool {
			return dhList[i].Date < dhList[j].Date
		})

		file := []byte{}

		fmt.Printf("Concatenating %d files\n", len(dhList))

		for _, dh := range dhList {
			fmt.Printf("- %s (%s) [%s]\n", dh.Path, dh.Date, group)
			file = append(file, dh.Data...)
		}

		fileName := fmt.Sprintf("%s_%s.log.gz", group, dhList[0].Date)

		if err := aws.UploadToS3(UPLOAD_BUCKET, fmt.Sprintf("%s/%s/%s", UPLOAD_PATH, UPLOAD_PATH_PREFIX, fileName), file); err != nil {
			fmt.Printf("warning - Unable to upload file %s (%s)\n", fileName, err)
			continue
		}

		for _, dh := range dhList {
			if err := os.Remove(dh.Path); err != nil {
				fmt.Printf("warning - Unable to delete file from local file system (%s)\n", err)
				continue
			}
			fmt.Printf("%s removed\n", dh.Path)
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

func createFileName(fileName string) string {
	return fmt.Sprintf("%s/%s_%s", STORAGE_DIR, fileName, time.Now().UTC().Format("2006-01-02_15-04-05"))
}
