#!/bin/sh

# export STORAGE_DIR=/media/james/usb_ssd
export STORAGE_DIR=./data
export AWS_PROFILE=thingco-dev
export UPLOAD_PATH_PREFIX=james
export MAX_FILE_SIZE=10000 # Bytes
export UPLOAD_INTERVAL=2 # Minutes

export P0=/dev/cu.usbserial-21240
# export P1=/dev/cu.usbserial-21350
# export P2=/dev/cu.usbserial-21360
# export P3=/dev/cu.usbserial-21370
export P1=
export P2=
export P3=
export P4=
export P5=
export P6=

# export P0=/dev/ttyUSB0
# export P1=/dev/ttyUSB1
# export P2=/dev/ttyUSB2
# export P3=/dev/ttyUSB3
# export P4=/dev/ttyUSB4
# export P5=/dev/ttyUSB5
# export P6=/dev/ttyUSB6