#!/bin/sh

export STORAGE_DIR=/media/james/usb_ssd
export AWS_PROFILE=thingco-dev
export UPLOAD_PATH_PREFIX=james
export MAX_FILE_SIZE=10000 # Bytes
export UPLOAD_INTERVAL=2 # Minutes

exec ./refurbination