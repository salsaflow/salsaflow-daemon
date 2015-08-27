package log

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
)

var defaultLogger = NewLogger()

func Info(req *http.Request, format string, v ...interface{}) {
	defaultLogger.Info(req, format, v...)
}

func Warn(req *http.Request, format string, v ...interface{}) {
	defaultLogger.Warn(req, format, v...)
}

func Error(req *http.Request, err error) {
	defaultLogger.Error(req, err)
}

type Logger struct {
	skipCallers int
}

func NewLogger() *Logger {
	return &Logger{4}
}

func (logger *Logger) IncreaseSkippedCallers() {
	logger.skipCallers++
}

func (logger *Logger) DecreaseSkippedCallers() {
	logger.skipCallers--
}

func (logger *Logger) Info(req *http.Request, format string, v ...interface{}) {
	logger.printRecord("INFO", req, format, v...)
}

func (logger *Logger) Warn(req *http.Request, format string, v ...interface{}) {
	logger.printRecord("WARNING", req, format, v...)
}

func (logger *Logger) Error(req *http.Request, err error) {
	logger.printRecord("ERROR", req, err.Error())
}

func (logger *Logger) printRecord(
	kind string,
	req *http.Request,
	format string,
	v ...interface{},
) {
	log.Printf("%v [request = %v %v] [position = %v]: %v\n",
		kind, req.Method, req.URL.Path, logger.trace(), fmt.Sprintf(format, v...))
}

func (logger *Logger) trace() string {
	pc := make([]uintptr, 1)
	runtime.Callers(logger.skipCallers, pc)
	fn := runtime.FuncForPC(pc[0])
	file, line := fn.FileLine(pc[0])
	return fmt.Sprintf("%s:%d %s", file, line, fn.Name())
}
