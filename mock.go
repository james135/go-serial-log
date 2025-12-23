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

	t := time.Now().Minute()

	if t%2 == 0 && ms.ExecutionCount < 1000 {

		ms.ExecutionCount++

		msg := []byte(fmt.Sprintf("hello there: %d\n", ms.ExecutionCount))

		if ms.ExecutionCount%2 == 0 {
			msg = []byte(fmt.Sprintf("only partial (%d) ", ms.ExecutionCount))
		}

		copy(b, msg)
		return len(msg), nil
	}

	return 0, io.EOF
}
