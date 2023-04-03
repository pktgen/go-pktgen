// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package helpers

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"sync"

	"github.com/pktgen/go-pktgen/internal/constants"
	"github.com/shirou/gopsutil/cpu"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
)

var (
	numCPUs        int
	VersionStr     string = "24.10.0"
	ApplicationStr string = "Go-pktgen Traffic Generator"
	CopyrightStr   string = "Copyright © 2023-2025 Intel Corporation"
	BuildDateStr   string = ""
)

// Version string
func Version() string {
	return VersionStr
}

// Build Date string
func BuildDate() string {
	return BuildDateStr
}

// Command string
func Application() string {
	return ApplicationStr
}

// Copyright string
func Copyright() string {
	return CopyrightStr
}

// CommandInfo returning the basic information string
func CommandInfo(color bool) string {
	if !color {
		return fmt.Sprintf("%s, Version: %s Pid: %d %s", Application(), Version(), os.Getpid(), Copyright())
	} else {
		return fmt.Sprintf("%s, Version: %s Pid: %d %s",
			cz.Yellow(Application()), cz.Green(Version()), os.Getpid(), cz.SkyBlue(Copyright()))
	}
}

// NumCPUs is the number of CPUs in the system (logical cores)
func NumCPUs() int {
	var once sync.Once

	once.Do(func() {
		num, err := cpu.Counts(true)
		if err != nil {
			fmt.Printf("Unable to get number of CPUs: %v", err)
			os.Exit(1)
		}
		numCPUs = num
	})

	return numCPUs
}

// Format the bytes into human readable format
func Format(units []string, v uint64, w ...interface{}) string {
	var index int

	bytes := float64(v)
	for index = 0; index < len(units); index++ {
		if bytes < 1024.0 {
			break
		}
		bytes = bytes / 1024.0
	}

	precision := uint64(0)
	for _, v := range w {
		precision = v.(uint64)
	}

	return fmt.Sprintf("%.*f %s", precision, bytes, units[index])
}

// FormatBytes into KB, MB, GB, ...
func FormatBytes(v uint64, w ...interface{}) string {

	return Format([]string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}, v, w...)
}

// FormatUnits into KB, MB, GB, ...
func FormatUnits(v uint64, w ...interface{}) string {

	return Format([]string{" ", "K", "M", "G", "T", "P", "E", "Z", "Y"}, v, w...)
}

// BitRate - return the network bit rate
func BitRate(ioPkts, ioBytes uint64) float64 {
	return float64(((ioPkts * constants.PktOverheadSize) + ioBytes) * 8)
}

func AcceptNumber(textToCheck string, lastChar rune) bool {

	return lastChar >= '0' && lastChar <= '9'
}

func AcceptIPv4(textToCheck string, lastChar rune) bool {

	return AcceptNumber(textToCheck, lastChar) || lastChar == '.'
}

func AcceptIPv4CiDR(textToCheck string, lastChar rune) bool {

	return AcceptNumber(textToCheck, lastChar) || lastChar == '.' || lastChar == '/'
}

func AcceptFloat(textToCheck string, lastChar rune) bool {

	return AcceptNumber(textToCheck, lastChar) || lastChar == '.'
}

func AcceptHex(textToCheck string, lastChar rune) bool {

	return AcceptNumber(textToCheck, lastChar) ||
		(lastChar >= 'a' && lastChar <= 'f') ||
		(lastChar >= 'A' && lastChar <= 'F')
}

func AcceptMac(textToCheck string, lastChar rune) bool {

	return AcceptHex(textToCheck, lastChar) || lastChar == ':'
}

func ParseNumberUint64(text string, val *uint64) error {

	if len(text) == 0 {
		return nil
	}
	if v, err := strconv.ParseUint(text, 10, 64); err != nil {
		return err
	} else {
		*val = v
		return nil
	}
}

func ParseNumberFloat64(text string, val *float64) error {

	if len(text) == 0 {
		return nil
	}
	if v, err := strconv.ParseFloat(text, 64); err != nil {
		return err
	} else {
		*val = v
		return nil
	}
}

func ParseNumberUint16(text string, val *uint16) error {

	if len(text) == 0 {
		return nil
	}
	if v, err := strconv.ParseUint(text, 10, 16); err != nil {
		return err
	} else {
		*val = uint16(v)
		return nil
	}
}

func SetupSignals(signals ...os.Signal) {

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, signals...)
	go func() {
		sig := <-sigs

		fmt.Printf("Signal: %v\n", sig)
		pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)

		os.Exit(1)
	}()
}
