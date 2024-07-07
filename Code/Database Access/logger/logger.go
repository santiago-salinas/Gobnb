package logger

import (
	"log"
	"os"
	"path/filepath"
)

var (
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
)

// Initialize initializes the logger with output to the specified file
func Initialize(logFile string) {
	// Ensure the log folder exists
	logDir := "log"
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// Open the log file in the log folder
	logFilePath := filepath.Join(logDir, logFile)
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	debugLogger = log.New(file, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime)
	warnLogger = log.New(file, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	fatalLogger = log.New(file, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Debug logs a debug message
func Debug(v ...interface{}) {
	debugLogger.Println(v...)
}

// Info logs an info message
func Info(v ...interface{}) {
	infoLogger.Println(v...)
}

// Warn logs a warning message
func Warn(v ...interface{}) {
	warnLogger.Println(v...)
}

// Error logs an error message
func Error(v ...interface{}) {
	errorLogger.Println(v...)
}

// Fatal logs a fatal message and exits
func Fatal(v ...interface{}) {
	fatalLogger.Fatalln(v...)
}
