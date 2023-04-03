// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package tlog

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"
)

// LogStates map of log id states
type LogStates map[string]bool
type TlogOption func()

const (
	devPrefixPath = "/dev/pts"
)

// TTYLog - Log tty information
type TTYLog struct {
	inited bool        // initialized flag
	tty    string      // tty path
	fd     *os.File    // file descriptor
	out    chan string // output channel
	quit   chan bool   // quit channel
	done   chan bool   // done channel
	states LogStates   // log states
}

var tlog *TTYLog = new(TTYLog) // global TTYLog instance

const (
	FatalLog string = "FatalLog" // FatalLog for fatal error log message
	ErrorLog string = "ErrorLog" // ErrorLog for error log messages
	WarnLog  string = "WarnLog"  // WarnLog for warning log messages
	InfoLog  string = "InfoLog"  // InfoLog for normal information
	DebugLog string = "DebugLog" // DebugLog for normal information
	PrintLog string = "PrintLog" // DoPrintLog for normal information
)

func tlogInit() {
	tlog.states = make(LogStates)

	tlog.states[FatalLog] = true
	tlog.states[ErrorLog] = true
	tlog.states[WarnLog] = true
	tlog.states[InfoLog] = true
	tlog.states[DebugLog] = false
	tlog.states[PrintLog] = true

	tlog.out = make(chan string, 32)
	tlog.quit = make(chan bool)
	tlog.done = make(chan bool)
}

func Open(options ...TlogOption) error {

	tlogInit()

	for _, option := range options {
		option()
	}

	if tlog.tty != "" {
		fd, err := os.OpenFile(tlog.tty, os.O_WRONLY, 0755)
		if err != nil {
			return err
		}
		tlog.fd = fd

		tlog.inited = true
		go logger()

		tlog.out <- fmt.Sprintf("\n*** TLOG STARTED AT %s ***\n", time.Now().Format(time.DateTime))
	}
	return nil
}

func WithLogID(ttyID int) TlogOption {
	return func() {
		tlog.tty = ""
		if ttyID > 0 {
			tlog.tty = fmt.Sprintf("%s/%d", devPrefixPath, ttyID)
			fmt.Printf("Using TTY: %s\n", tlog.tty)
		}
	}
}

// Close - close the channel and fd
func Close() {

	// If logger is not initialized, return immediately.
	// This is a safeguard against segfaults
	if tlog.inited {
		if tlog.quit != nil {
			tlog.quit <- true // Signal logger to quit
			<-tlog.done       // Wait for logger to finish
			close(tlog.quit)  // Close the done channel
			tlog.quit = nil
		}

		if tlog.out != nil {
			close(tlog.out)
			tlog.out = nil
		}
		if tlog.fd != nil {
			tlog.fd.Close()
			tlog.fd = nil
		}

		tlogInit() // Reset the log states and inited flag
	}
}

func logger() {

ForLoop:
	for {
		select {
		case <-tlog.quit: // Quit
			fmt.Fprintf(tlog.fd, "\n*** TLOG STOPPING AT %s ***\n", time.Now().Format(time.DateTime))
			break ForLoop
		case str := <-tlog.out:
			fmt.Fprintf(tlog.fd, "%s", str)
		}
	}
	tlog.done <- true
}

func isInited() bool {
	return tlog != nil && tlog.inited
}

// Register is a function to register new logging type strings
func Register(id string, state ...bool) {

	if tlog == nil || !tlog.inited {
		tlogInit()
	}
	flg := false
	if state != nil {
		flg = state[0]
	}

	tlog.states[id] = flg
}

// Delete a log id
func Delete(id string) error {

	_, ok := tlog.states[id]
	if ok {
		delete(tlog.states, id)
		return nil
	}

	return fmt.Errorf("log id not registered")
}

// State is a function to return the current logid state
func State(id string) (bool, error) {

	state, ok := tlog.states[id]
	if !ok {
		return false, fmt.Errorf("unknown logid %s", id)
	}
	return state, nil
}

// SetState on a logid
func SetState(id string, state bool) {
	tlog.states[id] = state
}

// IsActive - return true if log type id is true else false
func IsActive(id string) bool {

	state, ok := tlog.states[id]
	if !ok {
		return false
	}

	return state
}

// GetList returns the list of states and log ids
func GetList() LogStates {

	return tlog.states
}

// Log - output using printf like routine
func Log(id string, format string, a ...interface{}) int {
	s := fmt.Sprintf(getCallerFunc()+format, a...)
	if isInited() {
		if IsActive(id) {
			tlog.out <- s
		} else {
			s = ""
		}
	} else {
		fmt.Printf(s)
	}
	if id == FatalLog {
		os.Exit(1)
	}
	return len(s)
}

// FatalPrintf to print out fatal error messages
func FatalPrintf(format string, a ...interface{}) {
	Log(FatalLog, fmt.Sprintf(format, a...))
}

// ErrorPrintf to print out error messages
func ErrorPrintf(format string, a ...interface{}) {
	Log(ErrorLog, fmt.Sprintf(format, a...))
}

// WarnPrintf to print out warning messages
func WarnPrintf(format string, a ...interface{}) {
	Log(WarnLog, fmt.Sprintf(format, a...))
}

// InfoPrintf to print out informational messages
func InfoPrintf(format string, a ...interface{}) {
	Log(InfoLog, fmt.Sprintf(format, a...))
}

// DebugPrintf to print out informational messages
func DebugPrintf(format string, a ...interface{}) {
	Log(DebugLog, fmt.Sprintf(format, a...))
}

// Printf - output using printf like format without leading text and checks
func Printf(format string, a ...interface{}) int {
	return Log(PrintLog, fmt.Sprintf(format, a...))
}

// HexDump the data to the tlog
func HexDump(msg string, b []byte, n int) {
	if len(msg) > 0 {
		Printf("*** %s ***\n", msg)
	}
}

func WriteTo(name string) {
	if isInited() {
		if name == "" {
			name = "goroutine"
		}
		pprof.Lookup(name).WriteTo(tlog.fd, 2)
	}
}
