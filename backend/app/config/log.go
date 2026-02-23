package config

import "log"

var LoggingEnabled = true

func Logf(format string, v ...any) {
	if LoggingEnabled {
		log.Printf("[tracewaybackend] "+format, v...)
	}
}

func Logln(v ...any) {
	if LoggingEnabled {
		args := append([]any{"[tracewaybackend]"}, v...)
		log.Println(args...)
	}
}
