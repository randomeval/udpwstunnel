package mlog

import (
    "fmt"
    "log"
    "sync"
    "gopkg.in/natefinch/lumberjack.v2"
)

const (
    _           = iota
    LEVEL_FATAL
    LEVEL_ERROR
    LEVEL_INFO
    LEVEL_DEBUG 
)

// 
type Logger struct {
    mu          sync.Mutex
    haveInit    bool
    haveStdout  bool
    level       int
    debugLog   *log.Logger
    infoLog    *log.Logger
    errorLog   *log.Logger
    fatalLog   *log.Logger
}

// 
var gMlogLogger Logger

func LogInit(filename string, filesize int, backups int, haveStdout bool, level int) {
    if gMlogLogger.haveInit {
        gMlogLogger.haveStdout = haveStdout
        gMlogLogger.level      = level 
        return 
    }

    gMlogLogger.haveInit   = true
    gMlogLogger = LogNew(filename, filesize, backups)
    gMlogLogger.haveStdout = haveStdout
    gMlogLogger.level      = level 
}

func DebugLog(format string, v ...interface{}) {
    gMlogLogger.mu.Lock()
    defer gMlogLogger.mu.Unlock()

    if gMlogLogger.level >= LEVEL_DEBUG {
        if gMlogLogger.haveStdout {
            fmt.Printf(format, v...)
        }

        gMlogLogger.debugLog.Printf(format, v...)
    }
}

func InfoLog(format string, v ...interface{}) {
    gMlogLogger.mu.Lock()
    defer gMlogLogger.mu.Unlock()

    if gMlogLogger.level >= LEVEL_INFO {
        if gMlogLogger.haveStdout {
            fmt.Printf(format, v...)
        }
        gMlogLogger.infoLog.Printf(format, v...)
    }
}

func ErrorLog(format string, v ...interface{}) {
    gMlogLogger.mu.Lock()
    defer gMlogLogger.mu.Unlock()

    if gMlogLogger.level >= LEVEL_ERROR {
        if gMlogLogger.haveStdout {
            fmt.Printf(format, v...)
        }

        gMlogLogger.errorLog.Printf(format, v...)
    }
}

func FatalLog(format string, v ...interface{}) {
    gMlogLogger.mu.Lock()
    defer gMlogLogger.mu.Unlock()

    if gMlogLogger.haveStdout {
        fmt.Printf(format, v...)
    }
    gMlogLogger.fatalLog.Fatalf(format, v...)
}

func LogNew(filename string, filesize int, backups int) Logger {
    var ret Logger

    l := &lumberjack.Logger {
        Filename: filename,
        MaxSize: filesize,
        MaxBackups: backups,
    }
    
    ret.debugLog = log.New(l, "[DEBUG]", log.Ldate | log.Ltime | log.Lshortfile) 
    ret.infoLog  = log.New(l, "[INFO]", log.Ldate | log.Ltime | log.Lshortfile) 
    ret.errorLog = log.New(l, "[ERROR]", log.Ldate | log.Ltime | log.Lshortfile) 
    ret.fatalLog = log.New(l, "[FATAL]", log.Ldate | log.Ltime | log.Lshortfile) 
    return ret
}

