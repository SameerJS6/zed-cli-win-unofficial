package utils

import (
	"fmt"
)

// DebugMode controls whether debug messages are printed
// Set to false for production builds
const DebugMode bool = false

// Debug prints debug messages only when DebugMode is true
func Debug(format string, args ...interface{}) {
	if DebugMode {
		fmt.Printf("[DEBUG] "+format, args...)
	}
}

// Debugln prints debug messages with newline only when DebugMode is true
func Debugln(message string) {
	if DebugMode {
		fmt.Println("[DEBUG] " + message)
	}
}

// Info prints important user-facing messages (always shown)
func Info(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Infoln prints important user-facing messages with newline (always shown)
func Infoln(message string) {
	fmt.Println(message)
}

// Success prints success messages (always shown)
func Success(message string) {
	fmt.Println("✅ " + message)
}

// Warning prints warning messages (always shown)
func Warning(message string) {
	fmt.Println("⚠️ " + message)
}

// Error prints error messages (always shown)
func Error(message string) {
	fmt.Println("❌ " + message)
}
