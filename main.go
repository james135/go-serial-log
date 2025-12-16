package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/tarm/serial"
)

const (
	BAUD_RATE     = 1000
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
			writer.Write(buffer[:n])
			writer.Flush() // Flush immediately after writing
		}

		if time.Since(lastSync) >= SYNC_INTERVAL {
			writer.Flush()
			lastSync = time.Now().UTC()
		}
	}

	exit <- port
}

func main() {

	// argsWithoutProg := os.Args[1:]

	// if len(argsWithoutProg) > 1 {
	// 	path = argsWithoutProg[1]
	// }

	// fmt.Println(argsWithoutProg[0], path)

	path := "serial_log.txt"

	exitChan := make(chan string)

	// /dev/cu.usbserial-FT51P4YT
	// /dev/ttyUSB0
	go logSerialData("/dev/cu.usbserial-21240", path, exitChan)

	for i := range exitChan {
		fmt.Printf("%s has exited\n", i)
	}
}
