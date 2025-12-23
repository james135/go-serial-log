package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go-serial/aws"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tarm/serial"
)

func logSerialData(port string, logfileName string) {

	// Open serial port

	config := &serial.Config{Name: port, Baud: BAUD_RATE, ReadTimeout: time.Millisecond * 100}

	ser, err := serial.OpenPort(config)
	if err != nil {
		fmt.Printf("%s Serial error: %s\n", port, err)
		return
	}
	defer ser.Close()

	lastSync := time.Now().UTC()
	buffer := make([]byte, 256)

	var writer *bufio.Writer
	var file *os.File

	bytesWritten := 0
	logFile := []byte{}
	filePath := ""

	// Attempts to close any existing file and create a new one
	createWritableFile := func() error {

		if writer != nil {

			if err := writer.Flush(); err != nil {
				fmt.Printf("warning - Could not flush writer: %s\n", err)
			}

			writer = nil
		}

		if file != nil {

			// TODO - Move Sync() and Close() to the goroutine safely somehow

			if err := file.Sync(); err != nil {
				fmt.Printf("warning - Could not sync file: %s\n", err)
			}
			if err := file.Close(); err != nil {
				fmt.Printf("warning - Could not close file: %s\n", err)
			}

			file = nil

			go func(path string) {

				// Attempt to compress the most recently written to file
				if err := compressFile(path); err != nil {
					fmt.Printf("[warning] %s File compression error: %s\n", port, err)
				}

				// No more operations to be performed on active file by the serial logger - release for upload
				fw.RemoveFile(path)

			}(filePath)
		}

		filePath = createFileName(logfileName)

		// Open log file
		file, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		writer = bufio.NewWriter(file)

		fmt.Printf("Logging serial data from %s at %d baud to %s\n", port, BAUD_RATE, filePath)

		bytesWritten = 0

		fw.WatchFile(filePath)

		return nil
	}

	if err := createWritableFile(); err != nil {
		fmt.Printf("%s File error: %s\n", port, err)
		return
	}

	defer func() {
		if writer != nil {
			if err := writer.Flush(); err != nil {
				fmt.Printf("warning - Could not flush writer: %s\n", err)
			}
		}
		if file != nil {
			if err := file.Sync(); err != nil {
				fmt.Printf("warning - Could not sync file: %s\n", err)
			}
			if err := file.Close(); err != nil {
				fmt.Printf("warning - Could not close file: %s\n", err)
			}
		}
	}()

	ch := make(chan []byte, 100)

	// Serial reader - read from the serial port and push into buffered channel for processing

	go func() {
		// ms := MockSerial{}
		for {
			n, err := ser.Read(buffer)
			// n, err := ms.Read(buffer)
			if err != nil && err != io.EOF {

				// Transient error - retry to read from serial
				if errors.Is(err, os.ErrDeadlineExceeded) {
					continue
				}

				// Likely fatal error - exit
				fmt.Printf("[error] %s Read error: %s\n", port, err)
				break
			}

			ch <- buffer[:n]
		}
		close(ch)
	}()

	// Serial Processor - read from the data channel, process the data and write to files

	readsWithoutData := 0

	for data := range ch {

		// fmt.Printf("-------\n[%d] %s\n-------\n", len(data), data)

		if len(data) == 0 {

			// Check if we have stopped receiving data

			// fmt.Printf("Ticks without data: %+v (%s)\n", readsWithoutData, time.Now().UTC().Format("2006-01-02 15:04:05"))

			readsWithoutData++

			if readsWithoutData > 6000 {

				fmt.Printf("\n%s\n10 minutes since data on port %s - checking file size\n", time.Now().Format(time.RFC3339), port)

				if writer != nil {
					if err := writer.Flush(); err != nil {
						fmt.Printf("[warning] - %s could not Flush the writer (%s)\n", port, err)
					}
				}

				if file != nil {
					info, err := file.Stat()
					if err != nil {
						fmt.Printf("[warning] - %s could not read file stats (%s)\n", port, err)
					} else {

						fileSize := info.Size()
						fmt.Printf("Active file '%s' is currently %d bytes\n", filePath, fileSize)

						if fileSize > 0 {

							fmt.Printf("Dangling file '%s' found - rotating so it will become available for upload\n", filePath)

							if err := createWritableFile(); err != nil {
								fmt.Printf("%s File error: %s\n", port, err)
								return
							}
						}
					}
				}

				readsWithoutData = 0
			}

			continue
		}

		readsWithoutData = 0

		// Ensure valid UTF-8
		if !utf8.Valid(data) {
			data = bytes.ToValidUTF8(data, []byte("�"))
		}

		logFile = append(logFile, data...)

		lastNewLineIndex := bytes.LastIndex(logFile, []byte("\n"))

		if lastNewLineIndex >= 0 {

			rows := bytes.Split(logFile[:lastNewLineIndex+1], []byte("\n"))

			var sb strings.Builder

			for _, row := range rows {
				if len(row) > 0 {
					sb.WriteString(time.Now().UTC().Format("2006-01-02 15:04:05"))
					sb.WriteString(": ")
					sb.Write(row)
					sb.WriteByte('\n')
				}
			}

			// fmt.Printf("===============\n%s\n=================\n", sb.String())

			bw, err := writer.WriteString(sb.String())
			if err != nil {
				fmt.Printf("[warning] %s Write error: %s\n", port, err)
			}

			logFile = logFile[lastNewLineIndex+1:]

			bytesWritten += bw
		}

		if time.Since(lastSync) >= SYNC_INTERVAL {

			writer.Flush()
			lastSync = time.Now().UTC()

			if bytesWritten > MAX_FILE_SIZE {

				if err := createWritableFile(); err != nil {
					fmt.Printf("%s File error: %s\n", port, err)
					return
				}
			}
		}
	}

}

func main() {

	// Override defaults with env variables

	if os.Getenv("STORAGE_DIR") != "" {
		STORAGE_DIR = os.Getenv("STORAGE_DIR")
	}

	if os.Getenv("AWS_PROFILE") != "" {
		AWS_PROFILE = os.Getenv("AWS_PROFILE")
	}

	if os.Getenv("UPLOAD_PATH_PREFIX") != "" {
		UPLOAD_PATH_PREFIX = os.Getenv("UPLOAD_PATH_PREFIX")
	}

	if os.Getenv("UPLOAD_INTERVAL") != "" {
		i, err := strconv.Atoi(os.Getenv("UPLOAD_INTERVAL"))
		if err != nil {
			fmt.Printf("error - Unable to read UPLOAD_INTERVAL env variable\n")
			os.Exit(1)
		}
		UPLOAD_INTERVAL = i
	}

	if os.Getenv("MAX_FILE_SIZE") != "" {
		i, err := strconv.Atoi(os.Getenv("MAX_FILE_SIZE"))
		if err != nil {
			fmt.Printf("error - Unable to read MAX_FILE_SIZE env variable\n")
			os.Exit(1)
		}
		MAX_FILE_SIZE = i
	}

	// Create storage directory if it doesn't exist already

	if err := os.MkdirAll(STORAGE_DIR, 0755); err != nil {
		fmt.Printf("error - Unable to create storage directory '%s'\n", STORAGE_DIR)
		os.Exit(1)
	}

	// Create AWS Session

	aws.Configure(AWS_PROFILE)

	// Start serial reader/writer goroutines

	serialPorts := map[string]string{
		"0": os.Getenv("P0"),
		"1": os.Getenv("P1"),
		"2": os.Getenv("P2"),
		"3": os.Getenv("P3"),
		"4": os.Getenv("P4"),
		"5": os.Getenv("P5"),
		"6": os.Getenv("P6"),
	}

	for portName, port := range serialPorts {
		if port == "" {
			continue
		}
		go logSerialData(port, portName)
	}

	// Keep main process alive and attempt to upload any stored files to s3 every interval

	for {

		time.Sleep(time.Duration(UPLOAD_INTERVAL) * time.Minute)

		if err := UploadFiles(); err != nil {
			fmt.Printf("warning - File upload error: %s\n", err)
		}
	}
}
