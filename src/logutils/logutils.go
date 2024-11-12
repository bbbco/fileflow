package logutils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type LogWriter struct {
	logFile     *os.File
	LogFilePath string
	logger      *log.Logger
}

func NewLogWriter(appName string) (*LogWriter, error) {

	logFilePath, err := getLogFilePath(appName)
	if err != nil {
		return nil, fmt.Errorf("could not build log file path for app %s due to error: %v", appName, err)
	}
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not open log file: %v", err)
	}

	logger := log.New(logFile, "", log.LstdFlags) // LstdFlags Adds timestamps

	return &LogWriter{
		logFile:     logFile,
		logger:      logger,
		LogFilePath: logFilePath,
	}, nil
}

func getLogFilePath(appName string) (string, error) {
	// Use the OS-specific cache directory
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine cache directory: %v", err)
	}

	// Create an application-specific subdirectory
	logDir := filepath.Join(dir, appName)
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create log directory: %v", err)
	}

	// Return the path to the log file within this directory
	return filepath.Join(logDir, "application.log"), nil
}

func (lw *LogWriter) Write(message string) {
	lw.logger.Println(message)
}

func (lw *LogWriter) Fatal(message string, err error) {
	lw.logger.Fatal(message, err)
}
