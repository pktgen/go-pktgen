// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2019-2023 Intel Corporation

package tlog

import (
	"fmt"
	"os"
	"runtime/pprof"
	"strings"
	"time"
)

// LogStates map of log id states
type LogStates map[string]bool

const (
	devPrefixPath = "/dev/pts"
)

// TTYLog - Log tty information
type TTYLog struct {
	inited bool
	tty    string
	fd     *os.File
	out    chan string
	done   chan bool
	states LogStates
}

var tlog *TTYLog

const (
	// FatalLog for fatal error log message
	FatalLog string = "FatalLog"
	// ErrorLog for error log messages
	ErrorLog string = "ErrorLog"
	// WarnLog for warning log messages
	WarnLog string = "WarnLog"
	// InfoLog for normal information
	InfoLog string = "InfoLog"
	// DebugLog for normal information
	DebugLog string = "DebugLog"
)

func tlogInit() {
	tlog = new(TTYLog)
	tlog.states = make(LogStates)

	tlog.states[FatalLog] = true
	tlog.states[ErrorLog] = true
	tlog.states[WarnLog] = true
	tlog.states[InfoLog] = true
	tlog.states[DebugLog] = false

	tlog.out = make(chan string)
	tlog.done = make(chan bool)
}

func Open(w interface{}) error {

	tlogInit()

	switch v := w.(type) {
	case string:
		// Check if the provided argument is a device path. If not, prepend '/dev/pts/'
		s := w.(string)
		if !strings.HasPrefix(v, devPrefixPath) {
			tlog.tty = devPrefixPath + "/" + s
		} else {
			tlog.tty = s
		}
	case int, uint, int64, uint8, uint16, uint32:
		d := w.(int)
		tlog.tty = fmt.Sprintf("%s/%d", devPrefixPath, d)
	default:
		return fmt.Errorf("invalid argument type for Open()")
	}

	fd, err := os.OpenFile(tlog.tty, os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	tlog.fd = fd

	tlog.inited = true
	go logger()

	return nil
}

// Close - close the channel and fd
func Close() {

	// If logger is not initialized, return immediately.  This is a safeguard against segfaults
	if tlog.inited {
		if tlog.done != nil {
			tlog.done <- true       // Signal logger to quit
			time.Sleep(time.Second) // Wait for logger to finish
			close(tlog.done)		// Close the done channel
			tlog.done = nil
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

// Register is a function to register new logging type strings
func logger() {

ForLoop:
	for {
		select {
		case <-tlog.done: // Quit
			break ForLoop
		case str := <-tlog.out:
			fmt.Fprintf(tlog.fd, "%s", str)
		}
	}
}

func isInited() bool {
	return tlog != nil && tlog.inited
}

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

// FatalPrintf to print out fatal error messages
func FatalPrintf(format string, a ...interface{}) {
	if isInited() {
		s := fmt.Sprintf("Fatal: "+format, a...)
		tlog.out <- s
		os.Exit(1)
	}
}

// ErrorPrintf to print out error messages
func ErrorPrintf(format string, a ...interface{}) {
	if isInited() {
		if IsActive(ErrorLog) {
			s := fmt.Sprintf("Error: "+format, a...)
			tlog.out <- s
		}
	}
}

// WarnPrintf to print out warning messages
func WarnPrintf(format string, a ...interface{}) {
	if isInited() {
		if IsActive(WarnLog) {
			s := fmt.Sprintf("Warning: "+format, a...)
			tlog.out <- s
		}
	}
}

// InfoPrintf to print out informational messages
func InfoPrintf(format string, a ...interface{}) {
	if isInited() {
		if IsActive(InfoLog) {
			s := fmt.Sprintf("Info: "+format, a...)
			tlog.out <- s
		}
	}
}

// DebugPrintf to print out informational messages
func DebugPrintf(format string, a ...interface{}) {
	if isInited() {
		if IsActive(DebugLog) {
			s := fmt.Sprintf("Debug: "+format, a...)
			tlog.out <- s
		}
	}
}

// Log - output using printf like routine
func Log(id string, format string, a ...interface{}) int {
	if isInited() {
		if IsActive(id) {
			s := fmt.Sprintf(format, a...)
			tlog.out <- s
			if id == FatalLog {
				os.Exit(1)
			}
			return len(s)
		}
	}
	return 0
}

// Print a fmt.Print like function for verbose output
func Print(id string, a ...interface{}) int {
	return Log(id, fmt.Sprint(a...))
}

// Println a fmt.Println like function for verbose output
func Println(id string, format string, a ...interface{}) int {
	return Log(id, fmt.Sprintf(format, a...)+"\n")
}

// Printf a fmt.Print like function for verbose output
func Printf(id string, format string, a ...interface{}) int {
	return Log(id, fmt.Sprintf(format, a...))
}

// DoPrintf - output using printf like format without leading text and checks
func DoPrintf(format string, a ...interface{}) int {
	if isInited() {
		s := fmt.Sprintf(format, a...)
		tlog.out <- s
		return len(s)
	}
	return 0
}

// HexDump the data to the tlog
func HexDump(msg string, b []byte, n int) {
	if len(msg) > 0 {
		DoPrintf("*** %s ***\n", msg)
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
