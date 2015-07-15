package log

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
)

func Info(req *http.Request, format string, v ...interface{}) {
	printRecord("INFO", req, format, v...)
}

func Warn(req *http.Request, format string, v ...interface{}) {
	printRecord("WARNING", req, format, v...)
}

func Error(req *http.Request, err error) {
	printRecord("ERROR", req, err.Error())
}

func printRecord(kind string, req *http.Request, format string, v ...interface{}) {
	log.Printf("%v [request = %v %v] [position = %v]: %v\n",
		kind, req.Method, req.URL.Path, trace(), fmt.Sprintf(format, v...))
}

func trace() string {
	pc := make([]uintptr, 1)
	runtime.Callers(4, pc)
	fn := runtime.FuncForPC(pc[0])
	file, line := fn.FileLine(pc[0])
	return fmt.Sprintf("%s:%d %s", file, line, fn.Name())
}
