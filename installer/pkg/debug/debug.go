package debug

import (
	"log"
	"os"
)

var (
	Enabled bool = false
)

func Printf(format string, args ...interface{}) {
	if !isEnabled() {
		return
	}
	log.Printf(format, args...)
}

func isEnabled() bool {
	if Enabled {
		return true
	}
	if _, ok := os.LookupEnv("VIRTSHIFT_DEBUG"); ok {
		return true
	}
	return false
}
