/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2022-2024 Intel Corporation.
 */

package dbg

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
)

const (
	DefaultStackFrame = 2 // default stack frame index to dump skipping dbg routines
)

type DbgInfo struct {
	printAll    bool
	enableColor bool
	fileMap     map[string]bool
	funcMap     map[string]bool
}

type StackFrame struct {
	dbgInfo  *DbgInfo
	pc       uintptr
	file     string
	function string
	line     int
}

func (dbg *DbgInfo) SetPrintState(v bool) {
	dbg.printAll = v
}

func (dbg *DbgInfo) SetColorState(v bool) {
	dbg.enableColor = v
}

func (dbg *DbgInfo) AddFile(file string) {
	if _, ok := dbg.fileMap[file]; ok {
		return
	} else {
		dbg.fileMap[file] = true
	}
}

func (dbg *DbgInfo) AddFiles(funcNames []string) {
	for _, fn := range funcNames {
		dbg.AddFile(fn)
	}
}

func (dbg *DbgInfo) AddFunction(str string) {
	if _, ok := dbg.funcMap[str]; ok {
		return
	} else {
		dbg.funcMap[str] = true
	}
}

func (dbg *DbgInfo) AddFunctions(funcNames []string) {
	for _, fn := range funcNames {
		dbg.AddFunction(fn)
	}
}

func (dbg *DbgInfo) allowPrint(f *StackFrame) bool {
	if dbg.printAll {
		return true
	}
	if _, ok := dbg.fileMap[f.file]; ok {
		return true
	}
	for k := range dbg.funcMap {
		if strings.Contains(f.function, k) {
			return true
		}
	}
	return false
}

func (f *StackFrame) String() string {

	dbg := f.dbgInfo
	if dbg != nil && f != nil {
		fileFn := func(s string, a ...interface{}) string { return fmt.Sprintf(s, a...) }
		funcFn := func(s string, a ...interface{}) string { return fmt.Sprintf(s, a...) }
		lineFn := func(s string, a ...interface{}) string { return fmt.Sprintf(s, a...) }
		if dbg.enableColor {
			fileFn = color.New(color.FgHiCyan).SprintfFunc()
			funcFn = color.New(color.FgHiMagenta).SprintfFunc()
			lineFn = color.New(color.FgBlue).SprintfFunc()
		}
		return fmt.Sprintf("%s:%s:%s", fileFn(f.file), funcFn(f.function), lineFn("%d", f.line))
	}
	return ""
}

func New() *DbgInfo {
	return &DbgInfo{
		printAll:    false,
		enableColor: true,
		fileMap:     make(map[string]bool),
		funcMap:     make(map[string]bool),
	}
}

func (dbg *DbgInfo) StackFrame(stackFrame int) *StackFrame {

	if pc, path, line, ok := runtime.Caller(stackFrame); ok {
		file := filepath.Base(path)
		function := filepath.Base(runtime.FuncForPC(pc).Name())
		function = function[strings.Index(function, ".")+1:]
		f := &StackFrame{
			dbgInfo:  dbg,
			pc:       pc,
			file:     file,
			function: function,
			line:     line,
		}
		return f
	} else {
		return &StackFrame{
			dbgInfo:  dbg,
			pc:       0,
			file:     "",
			function: "",
			line:     0,
		}
	}
}

func (dbg *DbgInfo) caller(stackFrame int) string {

	return fmt.Sprintf("%v", dbg.StackFrame(DefaultStackFrame))
}

func (dbg *DbgInfo) backTrace(begin, traceCount int) []string {
	stacks := []string{}

	for i := begin; i < begin+traceCount; i++ {
		s := dbg.caller(i)
		if s == "" {
			break
		}
		stacks = append(stacks, s)
		if traceCount > 0 && i > traceCount {
			break
		}
	}
	return stacks
}

func (dbg *DbgInfo) Backtrace(traceCount int) string {

	funcFn := color.New(color.FgBlue).SprintfFunc()

	str := fmt.Sprintf("%s %s:\n", dbg.caller(4), funcFn("Backtrace"))
	for _, s := range dbg.backTrace(DefaultStackFrame, traceCount) {
		str += fmt.Sprintf("    %s\n", s)
	}
	return str[:len(str)-1] // remove last newline
}

func (dbg *DbgInfo) PrintBacktrace(traceCount int) {
	if dbg.printAll {
		fmt.Printf(":%s\n", dbg.Backtrace(traceCount))
	}
}

func (f *StackFrame) Sprintf(format string, a ...interface{}) string {
	return f.String() + " " + fmt.Sprintf(format, a...)
}

func (dbg *DbgInfo) Printf(format string, args ...interface{}) {

	f := dbg.StackFrame(DefaultStackFrame)
	if dbg.allowPrint(f) {
		fmt.Printf("%s", f.Sprintf(format, args...))
	}
}

func (dbg *DbgInfo) DoPrintf(format string, args ...interface{}) {

	f := dbg.StackFrame(DefaultStackFrame)
	if dbg.allowPrint(f) {
		fmt.Printf("%s", fmt.Sprintf(format, args...))
	}
}

func (dbg *DbgInfo) Errorf(format string, args ...interface{}) error {

	f := dbg.StackFrame(DefaultStackFrame)
	s := f.Sprintf(format, args...)

	if s[len(s)-1] == '\n' { // Remove trailing newline, if any
		s = s[:len(s)-1]
	}

	return fmt.Errorf(s)
}
