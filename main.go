package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go-serial/aws"
	"os"
	"time"
	"unicode/utf8"

	"github.com/tarm/serial"
)

var fw FileWatcher = FileWatcher{
	v: map[string]bool{},
}

func createFileName(fileName string) string {
	return fmt.Sprintf("%s/%s_%s", STORAGE_DIR, fileName, time.Now().UTC().Format("2006-01-02_15-04-05"))
}

func logSerialData(port string, logfileName string) {
	// Open serial port
	config := &serial.Config{Name: port, Baud: BAUD_RATE, ReadTimeout: time.Millisecond * 100}
	ser, err := serial.OpenPort(config)
	if err != nil {
		fmt.Printf("%s Serial error: %s", port, err)
		return
	}
	defer ser.Close()

	filePath := createFileName(logfileName)
	fw.WatchFile(filePath)

	// Open log file
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("%s File error: %s", port, err)
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
						toWrite += fmt.Sprintf("%s: %s", time.Now().UTC().Format(time.RFC3339), rows[i])
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

	if err := os.MkdirAll(STORAGE_DIR, 0755); err != nil {
		fmt.Printf("error - Unable to create storage directory\n")
		os.Exit(1)
	}

	aws.Configure(AWS_PROFILE)

	// config := map[string]string{
	// 	"/dev/cu.usbserial-21240": "port_D",
	// 	"/dev/cu.usbserial-21260": "port_F",
	// 	"/dev/cu.usbserial-21270": "port_G",
	// 	"/dev/cu.usbserial-21250": "port_E",
	// }

	config := map[string]string{
		"/dev/ttyUSB0": "0",
		"/dev/ttyUSB1": "1",
		"/dev/ttyUSB2": "2",
		"/dev/ttyUSB3": "3",
		"/dev/ttyUSB4": "4",
		"/dev/ttyUSB5": "5",
		"/dev/ttyUSB6": "6",
	}

	for port, logName := range config {
		go logSerialData(port, logName)
	}

	for {
		time.Sleep(UPLOAD_INTERVAL)
		if err := UploadFiles(); err != nil {
			fmt.Printf("warning - File upload error: %s\n", err)
		}
	}
}
