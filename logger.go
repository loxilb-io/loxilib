// SPDX-License-Identifier: Apache 2.0
// Copyright Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"fmt"
	"log"
	"os"
)

type LogLevelT int

const (
	LOG_EMERG LogLevelT = iota
	LOG_ALERT
	LOG_CRITICAL
	LOG_ERROR
	LOG_WARNING
	LOG_NOTICE
	LOG_INFO
	LOG_DEBUG
)

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

func LogItInit(logFile string, logLevel LogLevelT, toTTY bool) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	if logLevel < LOG_EMERG || logLevel > LOG_DEBUG {
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
	case LOG_EMERG:
		LogItEmer.Printf(format, v...)
		break
	case LOG_ALERT:
		LogItAlert.Printf(format, v...)
		break
	case LOG_CRITICAL:
		LogItCrit.Printf(format, v...)
		break
	case LOG_ERROR:
		LogItErr.Printf(format, v...)
		break
	case LOG_WARNING:
		LogItWarn.Printf(format, v...)
		break
	case LOG_NOTICE:
		LogItNotice.Printf(format, v...)
		break
	case LOG_INFO:
		LogItInfo.Printf(format, v...)
		break
	case LOG_DEBUG:
		LogItDebug.Printf(format, v...)
		break
	default:
		break
	}

	if LogTTY == true {
		fmt.Printf(format, v...)
	}
}
