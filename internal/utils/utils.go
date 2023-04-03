package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"log"
)

func printDebug(format string, args ...interface{}) {

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

			if file != "utils.go" {
				fmt.Printf("[%s:%s():%d] ", file, function, line)
                break
            }
		}
	}
	fmt.Printf(format, args...)
}

func Printf(format string, args ...interface{}) {

	printDebug(format, args...)
}

func PrintErr(format string, args ...interface{}) {

	printDebug("Error: " + format, args...)
}

// Dump out the spec of the given object
func DisplayStruct(msg, name string, spec interface{}) {

	if pc, path, line, ok := runtime.Caller(1); ok {
		file := filepath.Base(path)
		function := filepath.Base(runtime.FuncForPC(pc).Name())
		function = function[strings.Index(function, ".")+1:]

		fmt.Printf("[%s:%s():%d] ", file, function, line)
	}

	fmt.Printf("%v: %v.Spec -> %v\n", msg, name, MarshalIndent(spec))
}

// Marshal indented JSON string of the given object
func MarshalIndent(v interface{}) string {

	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err.Error())
	}
	return string(b)
}

func Indent(s string) string {
	var dst bytes.Buffer

	if err := json.Indent(&dst, []byte(s), "", "  "); err != nil {
		return s
	} else {
		return dst.String()
	}
}
