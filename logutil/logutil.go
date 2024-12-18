package logutil

import "log"

func Warn(message string) {
	log.Printf("[%s]: %s\n", "WARN", message)
}

func Info(message string) {
	log.Printf("[%s]: %s\n", "INFO", message)
}

func Fatal(message string) {
	log.Fatalf("[%s]: %s\n", "FATAL", message)
}
