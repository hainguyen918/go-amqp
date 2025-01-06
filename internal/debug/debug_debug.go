//go:build debug
// +build debug

package debug

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

var (
	debugLevel = 4
	logger     = log.New(os.Stderr, "", log.Lmicroseconds)
)

func init() {
	fmt.Println("\n\n\n INIT FUNCTION IS CALLED \n\n\n")
	level, err := strconv.Atoi(os.Getenv("DEBUG_LEVEL"))
	if err != nil {
		return
	}

	debugLevel = level
}

// Log writes the formatted string to stderr.
// Level indicates the verbosity of the messages to log.
// The greater the value, the more verbose messages will be logged.
func Log(level int, format string, v ...any) {
	// print all the debug log (for testing purpose only)
	logger.Printf(format, v...)
	fmt.Println("\n\n\n Log FUNCTION IS CALLED \n\n\n")
	fmt.Printf(format, v...)

	if level <= debugLevel {
		logger.Printf(format, v...)
	}
}

// Assert panics if the specified condition is false.
func Assert(condition bool) {
	if !condition {
		panic("assertion failed!")
	}
}

// Assert panics with the provided message if the specified condition is false.
func Assertf(condition bool, msg string, v ...any) {
	if !condition {
		panic(fmt.Sprintf(msg, v...))
	}
}
