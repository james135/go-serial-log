package main

import (
	"fmt"
	"io"
	"time"
)

type MockSerial struct {
	ExecutionCount int
}

func (ms *MockSerial) Read(b []byte) (n int, err error) {

	time.Sleep(50 * time.Millisecond)

	if time.Now().Minute()%2 == 0 {

		ms.ExecutionCount++

		msg := []byte(fmt.Sprintf("hello there: %d\n", ms.ExecutionCount))
		copy(b, msg)
		return len(msg), nil
	}

	return 0, io.EOF
}
