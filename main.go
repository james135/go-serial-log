package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"time"
	"unicode/utf8"

	"github.com/tarm/serial"
)

const (
	BAUD_RATE     = 115200
	SYNC_INTERVAL = 120 * time.Second
)

func logSerialData(port string, logfile string, exit chan string) {
	// Open serial port
	config := &serial.Config{Name: port, Baud: BAUD_RATE, ReadTimeout: time.Millisecond * 100}
	ser, err := serial.OpenPort(config)
	if err != nil {
		fmt.Println("Serial error:", err)
		return
	}
	defer ser.Close()

	// Open log file
	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("File error:", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	fmt.Printf("Logging serial data from %s at %d baud to %s\n", port, BAUD_RATE, logfile)

	lastSync := time.Now().UTC()
	buffer := make([]byte, 256)

	defer func() {
		writer.Flush()
		file.Sync() // optional but safer
	}()

	for {
		n, err := ser.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				continue // Ignore EOF and wait for new data
			}
			fmt.Println("Read error:", err)
			break
		}

		if n > 0 {
			data := buffer[:n]

			// Ensure valid UTF-8
			if !utf8.Valid(data) {
				data = bytes.ToValidUTF8(data, []byte("�"))
			}

			_, _ = writer.Write(data)
			// writer.Flush()
		}

		if time.Since(lastSync) >= SYNC_INTERVAL {
			writer.Flush()
			lastSync = time.Now().UTC()
		}
	}

	exit <- port
}

func main() {

	config := map[string]string{
		"/dev/cu.usbserial-21240": "port_D",
		"/dev/cu.usbserial-21260": "port_F",
		"/dev/cu.usbserial-21270": "port_G",
		"/dev/cu.usbserial-21250": "port_E",
	}

	exitChan := make(chan string)

	for port, logName := range config {
		go logSerialData(port, logName, exitChan)
	}

	// /dev/cu.usbserial-FT51P4YT
	// /dev/ttyUSB0
	// go logSerialData("/dev/cu.usbserial-21240", "serial_log.txt", exitChan)

	for i := range exitChan {
		fmt.Printf("%s has exited\n", i)
	}
}
