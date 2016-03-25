package logger

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
)

const FLAGS = 0

var GOPATH = os.Getenv("GOPATH")
var debugreplacer = regexp.MustCompile(`\+0x[0-9a-f]+$`)

var logDebug *log.Logger
var logWarn = log.New(os.Stderr, "warning: ", 0)
var logError = log.New(os.Stderr, "error: ", 0)

func init() {
	DEBUG_OUTPUT, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	isDebug := os.Getenv("DEBUG")
	if len(isDebug) > 0 {
		DEBUG_OUTPUT = os.Stderr
	}
	logDebug = log.New(DEBUG_OUTPUT, "debug: ", FLAGS)
}

func getCaller() string {
	buf := make([]byte, 1<<16)
	stackLen := runtime.Stack(buf, false)
	if stackLen == 0 {
		return ""
	}
	// log.Println(string(buf))
	stack := bytes.Split(buf, []byte{'\n'})
	if len(stack) < 7 {
		return ""
	}
	return formatCaller(string(stack[6]))
}

func formatCaller(str string) string {
	str = strings.TrimSpace(str)
	str = strings.Replace(str, GOPATH+"/src/", "", -1)
	return debugreplacer.ReplaceAllString(str, "")
}

func Debug(args ...interface{}) {
	args = append([]interface{}{getCaller()}, args...)
	logDebug.Println(args...)
}

func Debugf(format string, args ...interface{}) {
	format = getCaller() + " " + format
	logDebug.Printf(format, args...)
}

func Warn(args ...interface{}) {
	args = append([]interface{}{getCaller()}, args...)
	logWarn.Println(args...)
}

func Warnf(format string, args ...interface{}) {
	format = getCaller() + " " + format
	logWarn.Printf(format, args...)
}

func Error(args ...interface{}) {
	args = append([]interface{}{getCaller()}, args...)
	logError.Println(args...)
}

func Errorf(format string, args ...interface{}) {
	format = getCaller() + " " + format
	logError.Printf(format, args...)
}

func Fatal(args ...interface{}) {
	logError.Println(args...)
	os.Exit(1)
}

func Fatalf(format string, args ...interface{}) {
	logError.Printf(format, args...)
	os.Exit(1)
}

func Log(args ...interface{}) {
	fmt.Println(args...)
}

func Logf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func EnableDebug() {
	logDebug = log.New(os.Stderr, "debug: ", FLAGS)
}
