package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go-serial/aws"
	"os"
	"strconv"
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

	filePath := createFileName(logfileName)
	fw.WatchFile(filePath)

	// Open log file
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("%s File error: %s\n", port, err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	fmt.Printf("Logging serial data from %s at %d baud to %s\n", port, BAUD_RATE, filePath)

	lastSync := time.Now().UTC()
	buffer := make([]byte, 256)

	defer func() {
		writer.Flush()
		file.Sync()
	}()

	runRun := []byte{}
	bytesWritten := 0

	for {
		n, err := ser.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				continue // Ignore EOF and wait for new data
			}
			fmt.Printf("%s Read error: %s\n", port, err)
			break
		}

		if n > 0 {
			data := buffer[:n]

			// Ensure valid UTF-8
			if !utf8.Valid(data) {
				data = bytes.ToValidUTF8(data, []byte("�"))
			}

			runRun = append(runRun, data...)

			lastNewLineIndex := bytes.LastIndex(runRun, []byte("\n"))

			if lastNewLineIndex >= 0 {

				rows := bytes.Split(runRun[:lastNewLineIndex+1], []byte("\n"))

				toWrite := ""

				for i := range rows {
					if len(rows[i]) > 0 {
						toWrite += fmt.Sprintf("%s: %s", time.Now().UTC().Format("2006-01-02 15:04:05"), rows[i])
					}
				}

				bw, err := writer.Write([]byte(toWrite))
				if err != nil {
					fmt.Printf("[warning] %s Write error: %s\n", port, err)
				}

				runRun = runRun[lastNewLineIndex+1:]

				bytesWritten += bw
			}
		}

		if time.Since(lastSync) >= SYNC_INTERVAL {

			writer.Flush()
			lastSync = time.Now().UTC()

			if bytesWritten > MAX_FILE_SIZE {

				file.Close()

				if err := compressFile(filePath); err != nil {
					fmt.Printf("[warning] %s File compression error: %s", port, err)
				}

				fw.RemoveFile(logfileName)

				filePath = createFileName(logfileName)
				fw.WatchFile(filePath)

				// Open log file
				file, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Printf("%s File error: %s", port, err)
					return
				}
				defer file.Close()

				writer = bufio.NewWriter(file)

				bytesWritten = 0
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
