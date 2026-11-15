// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
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

// Logger configuration fields are for construction or exclusive use. Use the
// synchronized methods during concurrent use. A Logger must not be copied
// after first use.
type Logger struct {
	ttyMu        sync.Mutex
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
	// DefaultLogger is retained for source compatibility with exclusive access.
	// Use SwapDefaultLogger to replace it during concurrent logging.
	DefaultLogger *Logger

	// Only enabled calls enter this critical section. Each publication has its
	// own group, which no new call can join after the default changes.
	defaultLoggerMu    sync.Mutex
	defaultLoggerUsers = new(sync.WaitGroup)
)

func loadDefaultLogger() *Logger {
	return (*Logger)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&DefaultLogger))))
}

func storeDefaultLogger(logger *Logger) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&DefaultLogger)), unsafe.Pointer(logger))
}

// LogItInit - Initialize the logger
// logFile - name of the logfile
// logLevel - specify current loglevel
// toTTY - specify if logs need to be redirected to TTY as well or not
func LogItInit(logFile string, logLevel LogLevelT, toTTY bool) *Logger {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := NewLogger(file, logLevel, toTTY)
	if err != nil {
		file.Close()
		log.Fatal(err)
	}
	defaultLoggerMu.Lock()
	if loadDefaultLogger() == nil {
		storeDefaultLogger(logger)
	}
	defaultLoggerMu.Unlock()
	return logger
}

// NewLogger constructs a complete logger without publishing it. The caller owns
// output and must keep it open until all uses (including default uses) finish.
func NewLogger(output io.Writer, logLevel LogLevelT, toTTY bool) (*Logger, error) {
	if output == nil || logLevel < LogEmerg || logLevel > LogTrace {
		return nil, errors.New("invalid logger output or level")
	}
	logger := new(Logger)

	logger.CurrLogLevel = logLevel
	logger.LogTTY = toTTY
	logger.LogItEmer = log.New(output, "EMER: ", log.Ldate|log.Ltime)
	logger.LogItAlert = log.New(output, "ALRT: ", log.Ldate|log.Ltime)
	logger.LogItCrit = log.New(output, "CRIT: ", log.Ldate|log.Ltime)
	logger.LogItErr = log.New(output, "ERR:  ", log.Ldate|log.Ltime)
	logger.LogItWarn = log.New(output, "WARN: ", log.Ldate|log.Ltime)
	logger.LogItNotice = log.New(output, "NOTI: ", log.Ldate|log.Ltime)
	logger.LogItInfo = log.New(output, "INFO: ", log.Ldate|log.Ltime)
	logger.LogItDebug = log.New(output, "DBG:  ", log.Ldate|log.Ltime)
	logger.LogItTrace = log.New(output, "TRACE:  ", log.Ldate|log.Ltime)

	return logger, nil
}

// SwapDefaultLogger publishes a fully configured logger (nil disables logging).
// It waits for LogIt calls through the old publication to finish. The caller
// may close that output after the swap unless it is also in use elsewhere.
// A log writer or formatter must not call SwapDefaultLogger from within a log
// call, since that call would have to wait for itself to finish.
func SwapDefaultLogger(logger *Logger) *Logger {
	defaultLoggerMu.Lock()
	previous := loadDefaultLogger()
	previousUsers := defaultLoggerUsers
	defaultLoggerUsers = new(sync.WaitGroup)
	storeDefaultLogger(logger)
	defaultLoggerMu.Unlock()
	previousUsers.Wait()
	return previous
}

// Log uses Printf format internally
// Arguments are considered in-line with fmt.Printf.
func (logger *Logger) Log(l LogLevelT, format string, v ...interface{}) {
	if !logger.enabled(l) {
		return
	}
	logger.log(l, format, v...)
}

func (logger *Logger) enabled(level LogLevelT) bool {
	return level >= LogEmerg && level <= LogTrace && level <= logger.LogItGetLevel()
}

func (logger *Logger) log(l LogLevelT, format string, v ...interface{}) {
	logger.ttyMu.Lock()
	tty := logger.LogTTY
	logger.ttyMu.Unlock()
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

	if tty {
		fmt.Printf("%s ", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf(format, v...)
	}
}

// LogIt uses Printf format internally
func LogIt(l LogLevelT, format string, v ...interface{}) {
	logger := loadDefaultLogger()
	if logger == nil || !logger.enabled(l) {
		return
	}

	defaultLoggerMu.Lock()
	logger = loadDefaultLogger()
	if logger == nil || !logger.enabled(l) {
		defaultLoggerMu.Unlock()
		return
	}
	users := defaultLoggerUsers
	users.Add(1)
	defaultLoggerMu.Unlock()
	defer users.Done()
	logger.log(l, format, v...)
}

// LogItSetLevel - Set level of the logger
func (logger *Logger) LogItSetLevel(logLevel LogLevelT) error {

	if logLevel < LogEmerg || logLevel > LogTrace {
		return errors.New("LogLevel is out of bounds")
	}

	atomic.StoreUintptr((*uintptr)(unsafe.Pointer(&logger.CurrLogLevel)), uintptr(logLevel))
	return nil
}

// LogItGetLevel reads the configured level while other goroutines log or update it.
func (logger *Logger) LogItGetLevel() LogLevelT {
	// LogLevelT has the width of int. A uintptr has the same width on Go's
	// supported architectures, including 32-bit targets.
	return LogLevelT(atomic.LoadUintptr((*uintptr)(unsafe.Pointer(&logger.CurrLogLevel))))
}

// LogItSetTTY changes terminal mirroring and returns its previous setting.
func (logger *Logger) LogItSetTTY(enabled bool) bool {
	logger.ttyMu.Lock()
	defer logger.ttyMu.Unlock()
	previous := logger.LogTTY
	logger.LogTTY = enabled
	return previous
}
