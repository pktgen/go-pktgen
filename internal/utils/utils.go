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

// Get File, function and line number formatted in a string
func GetFFL() string {
	if pc, file, line, ok := runtime.Caller(2); ok {
		file = filepath.Base(file)
		function := filepath.Base(runtime.FuncForPC(pc).Name())
		function = function[strings.Index(function, ".")+1:]
	return fmt.Sprintf("[%s:%s():%d] ", file, function, line)
	}
	return "Unknown caller: "
}
