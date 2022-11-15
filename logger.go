// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"fmt"
	"log"
	"os"
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
)

// variables used
var (
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
)

// LogItInit - Initialize the logger
// logFile - name of the logfile
// logLevel - specify current loglevel
// toTTY - specify if logs need to be redirected to TTY as well or not
func LogItInit(logFile string, logLevel LogLevelT, toTTY bool) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	if logLevel < LogEmerg || logLevel > LogDebug {
		log.Fatal(err)
	}

	CurrLogLevel = logLevel
	LogTTY = toTTY
	LogItEmer = log.New(file, "EMER: ", log.Ldate|log.Ltime)
	LogItAlert = log.New(file, "ALRT: ", log.Ldate|log.Ltime)
	LogItCrit = log.New(file, "CRIT: ", log.Ldate|log.Ltime)
	LogItErr = log.New(file, "ERR:  ", log.Ldate|log.Ltime)
	LogItWarn = log.New(file, "WARN: ", log.Ldate|log.Ltime)
	LogItNotice = log.New(file, "NOTI: ", log.Ldate|log.Ltime)
	LogItInfo = log.New(file, "INFO: ", log.Ldate|log.Ltime)
	LogItDebug = log.New(file, "DBG:  ", log.Ldate|log.Ltime)
}

// LogIt uses Printf format internally
// Arguments are considered in-line with fmt.Printf.
func LogIt(l LogLevelT, format string, v ...interface{}) {
	if l < 0 || l > CurrLogLevel {
		return
	}
	switch l {
	case LogEmerg:
		LogItEmer.Printf(format, v...)
		break
	case LogAlert:
		LogItAlert.Printf(format, v...)
		break
	case LogCritical:
		LogItCrit.Printf(format, v...)
		break
	case LogError:
		LogItErr.Printf(format, v...)
		break
	case LogWarning:
		LogItWarn.Printf(format, v...)
		break
	case LogNotice:
		LogItNotice.Printf(format, v...)
		break
	case LogInfo:
		LogItInfo.Printf(format, v...)
		break
	case LogDebug:
		LogItDebug.Printf(format, v...)
		break
	default:
		break
	}

	if LogTTY == true {
		fmt.Printf(format, v...)
	}
}
