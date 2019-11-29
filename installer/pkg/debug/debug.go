package debug

import (
	"log"
	"os"
)

var (
	debugLogs bool = false
)

func Printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func init() {
	if _, ok := os.LookupEnv("VIRTSHIFT_DEBUG"); ok {
		debugLogs = true
	}
}
