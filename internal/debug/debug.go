package debug

import "log"

var debug bool

func SetDebug(enableDebug bool) {
	debug = enableDebug
}

func Debugf(format string, args ...interface{}) {
	if debug {
		log.Printf(format, args...)
	}
}
