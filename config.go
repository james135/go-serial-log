package main

import "time"

// Default environment variables (overwrite in env.sh)
var (
	STORAGE_DIR        = "./data"
	MAX_FILE_SIZE      = 1000 * 1000 * 1000 * 1 // 1 Mb
	UPLOAD_INTERVAL    = 60 * 12                // 12 Hours
	NO_DATA_READ_COUNT = 10 * 60 * 15           // ~15 Minutes (approx. 10 reads per second)
	AWS_PROFILE        = "thingco-dev"
	UPLOAD_PATH_PREFIX = "default"
)

const (
	UPLOAD_BUCKET = "platformsupport"
	UPLOAD_PATH   = "r13"
)

const (
	BAUD_RATE     = 115200
	SYNC_INTERVAL = 15 * time.Second
)
