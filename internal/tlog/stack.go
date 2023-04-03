package tlog

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"log"
)

func getCallerFunc() string {

	var pc uintptr
	var file, path, function string
	var line int

	var pcs = make([]uintptr, 15)

	n := runtime.Callers(1, pcs)
	if n > 0 {
		frames := runtime.CallersFrames(pcs[:n])
		for {
			frame, _ := frames.Next()
			pc, path, line = frame.PC, frame.File, frame.Line

			file = filepath.Base(path)
			function = filepath.Base(runtime.FuncForPC(pc).Name())
			function = function[strings.Index(function, ".")+1:]

			if file == "stack.go" || file == "tlog.go" {
				continue
			} else {
				return fmt.Sprintf("[%-32s:%4d] ", fmt.Sprintf("%s:%s()", file, function), line)
            }
		}
	}

	return ""
}

// Dump out the spec of the given object
func DisplayStruct(msg, name string, spec interface{}) {

	if pc, path, line, ok := runtime.Caller(1); ok {
		file := filepath.Base(path)
		function := filepath.Base(runtime.FuncForPC(pc).Name())
		function = function[strings.Index(function, ".")+1:]

		fmt.Printf("[%s:%s():%d] ", file, function, line)
	}

	var v interface{}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("%v: %v.Spec -> %v\n", msg, name, string(b))
}
