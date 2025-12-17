package main

import "time"

// var config = map[string]string{
// 	"/dev/cu.usbserial-21240": "port_D",
// 	"/dev/cu.usbserial-21260": "port_F",
// 	"/dev/cu.usbserial-21270": "port_G",
// 	"/dev/cu.usbserial-21250": "port_E",
// }

var config = map[string]string{
	"/dev/ttyUSB0": "0",
	"/dev/ttyUSB1": "1",
	"/dev/ttyUSB2": "2",
	"/dev/ttyUSB3": "3",
	"/dev/ttyUSB4": "4",
	"/dev/ttyUSB5": "5",
	"/dev/ttyUSB6": "6",
}

var (
	STORAGE_DIR        = "./data"
	MAX_FILE_SIZE      = 1000 * 1000 * 1000 * 1 // 1 Mb
	UPLOAD_INTERVAL    = 60 * 12                // 12 Hours
	AWS_PROFILE        = "thingco-dev"
	UPLOAD_PATH_PREFIX = "default"
)

const (
	UPLOAD_BUCKET = "platformsupport"
	UPLOAD_PATH   = "r13"
)

const (
	BAUD_RATE     = 115200
	SYNC_INTERVAL = 30 * time.Second
)
