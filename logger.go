// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevelT - Current log-level
type LogLevelT int

// Log levels
const (
	LogEmerg LogLevelT = iota
	LogAlert
	LogCritical
	LogError
	LogWarning
	LogNotice
	LogInfo
	LogDebug
	LogTrace
)

type Logger struct {
	LogTTY       bool
	CurrLogLevel LogLevelT
	LogItEmer    *log.Logger
	LogItAlert   *log.Logger
	LogItCrit    *log.Logger
	LogItErr     *log.Logger
	LogItWarn    *log.Logger
	LogItNotice  *log.Logger
	LogItInfo    *log.Logger
	LogItDebug   *log.Logger
	LogItTrace   *log.Logger
}

var (
	DefaultLogger *Logger
)

// LogItInit - Initialize the logger
// logFile - name of the logfile
// logLevel - specify current loglevel
// toTTY - specify if logs need to be redirected to TTY as well or not
func LogItInit(logFile string, logLevel LogLevelT, toTTY bool) *Logger {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	if logLevel < LogEmerg || logLevel > LogTrace {
		log.Fatal(err)
	}

	logger := new(Logger)

	logger.CurrLogLevel = logLevel
	logger.LogTTY = toTTY
	logger.LogItEmer = log.New(file, "EMER: ", log.Ldate|log.Ltime)
	logger.LogItAlert = log.New(file, "ALRT: ", log.Ldate|log.Ltime)
	logger.LogItCrit = log.New(file, "CRIT: ", log.Ldate|log.Ltime)
	logger.LogItErr = log.New(file, "ERR:  ", log.Ldate|log.Ltime)
	logger.LogItWarn = log.New(file, "WARN: ", log.Ldate|log.Ltime)
	logger.LogItNotice = log.New(file, "NOTI: ", log.Ldate|log.Ltime)
	logger.LogItInfo = log.New(file, "INFO: ", log.Ldate|log.Ltime)
	logger.LogItDebug = log.New(file, "DBG:  ", log.Ldate|log.Ltime)
	logger.LogItTrace = log.New(file, "TRACE:  ", log.Ldate|log.Ltime)

	if DefaultLogger == nil {
		DefaultLogger = logger
	}

	return logger
}

// Log uses Printf format internally
// Arguments are considered in-line with fmt.Printf.
func (logger *Logger) Log(l LogLevelT, format string, v ...interface{}) {
	if l < 0 || l > logger.CurrLogLevel {
		return
	}
	switch l {
	case LogEmerg:
		logger.LogItEmer.Printf(format, v...)
	case LogAlert:
		logger.LogItAlert.Printf(format, v...)
	case LogCritical:
		logger.LogItCrit.Printf(format, v...)
	case LogError:
		logger.LogItErr.Printf(format, v...)
	case LogWarning:
		logger.LogItWarn.Printf(format, v...)
	case LogNotice:
		logger.LogItNotice.Printf(format, v...)
	case LogInfo:
		logger.LogItInfo.Printf(format, v...)
	case LogDebug:
		logger.LogItDebug.Printf(format, v...)
	case LogTrace:
		logger.LogItTrace.Printf(format, v...)
	default:
		break
	}

	if logger.LogTTY {
		fmt.Printf("%s ", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf(format, v...)
	}
}

// LogIt uses Printf format internally
func LogIt(l LogLevelT, format string, v ...interface{}) {
	if DefaultLogger == nil {
		return
	}

	DefaultLogger.Log(l, format, v...)
}

// LogItSetLevel - Set level of the logger
func (logger *Logger) LogItSetLevel(logLevel LogLevelT) error {

	if logLevel < LogEmerg || logLevel > LogTrace {
		return errors.New("LogLevel is out of bounds")
	}

	logger.CurrLogLevel = logLevel
	return nil
}
