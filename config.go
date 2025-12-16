package main

import "time"

const (
	// MAX_FILE_SIZE = 1000 * 1000 * 1000 * 1 // 1 Mb
	MAX_FILE_SIZE   = 1000 * 10
	UPLOAD_INTERVAL = 120 * time.Second
)

const (
	STORAGE_DIR = "/media/james/usb_ssd"
	// STORAGE_DIR   = "./data"
)

const (
	AWS_PROFILE   = "thingco-dev"
	UPLOAD_BUCKET = "platformsupport"
	UPLOAD_PATH   = "r13"
)

const (
	BAUD_RATE     = 115200
	SYNC_INTERVAL = 10 * time.Second
)
