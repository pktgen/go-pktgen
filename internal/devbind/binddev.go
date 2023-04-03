// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022-2024 Intel Corporation

package devbind

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pktgen/go-pktgen/internal/tlog"
)

const (
	// Command to retrieve hardware information
	hwInfoCmd = "lshw -json -c network"

	// Command to retrieve IP route information
	ipRouteCmd = "ip -j -o route"

	// Command to retrieve PCI lines
	pciLinesCmd = "lspci | grep Ether"

	// Default timeout for commands
	defaultTimeout = 10 * time.Second

	// Paths to file system files for driver control, bind, unbind, and driver override
	driverBind     = "/sys/bus/pci/drivers/%s/bind"
	driverUnbind   = "/sys/bus/pci/drivers/%s/unbind"
	driverOverride = "/sys/bus/pci/devices/%s/driver_override"
	modProbe       = "modprobe vfio-pci"

	// The shell command to execute or path to the shell command
	shellCmd = "bash"
)

// JSON information about network interfaces from 'ip -j -o route' command.
type IPRoute []struct {
	Dst      string `json:"dst"`
	Gateway  string `json:"gateway,omitempty"`
	Dev      string `json:"dev"`
	Protocol string `json:"protocol"`
	Prefsrc  string `json:"prefsrc"`
	Metric   int    `json:"metric,omitempty"`
	Flags    []any  `json:"flags"`
	Scope    string `json:"scope,omitempty"`
}

// JSON information about network interfaces from 'lshw -json -c network' command.
type HwInfo struct {
	ID          string        `json:"id"`
	Claimed     bool          `json:"claimed"`
	Product     string        `json:"product"`
	Vendor      string        `json:"vendor"`
	BusInfo     string        `json:"businfo"`
	LogicalName string        `json:"logicalname"`
	Config      Configuration `json:"configuration"`
}

type Configuration struct {
	Driver string `json:"driver"`
}

type bindInfo struct {
	BusInfo        string
	Driver         string
	LogicalName    string
	OriginalDriver string
}

type DevBind struct {
	hwInfo      []*HwInfo           // Hardware information
	hwDriverMap map[string]*HwInfo  // Map of device information using driver ID
	hwBusMap    map[string]*HwInfo  // Map of device information using bus ID
	ipRoute     IPRoute             // IP route information
	pciLines    []string            // PCI lines
	timeout     time.Duration       // Timeout for commands
	quit        chan bool           // Channel to signal quit
	pciAddrs    map[string]bindInfo // Map of PCI addresses to device IDs
	shellCmd    string              // Path to shell command
}

type DevBindOption func(*DevBind)

func WithTimeout(sec time.Duration) DevBindOption {

	return func(db *DevBind) {
		db.timeout = sec * time.Second
	}
}

func WithShellCmd(path string) DevBindOption {

	return func(db *DevBind) {
		db.shellCmd = path
	}
}

func New(options ...DevBindOption) *DevBind {

	db := &DevBind{
		hwDriverMap: make(map[string]*HwInfo),
		hwBusMap:    make(map[string]*HwInfo),
		pciAddrs:    make(map[string]bindInfo),
		timeout:     defaultTimeout,
		quit:        make(chan bool),
		shellCmd:    shellCmd,
	}

	// Process the option function calls
	for _, f := range options {
		f(db)
	}

	runBashCmd(modProbe) // Make sure vfio-pci module is loaded

	return db
}

func (db *DevBind) Start() {

	go func() {
		// Fetch network interface information
		go db.updatePCILines()
		go db.updateIPRoute()
		go db.updateHWInfo()
		for {
			select {
			case <-db.quit: // Stop the goroutine
				tlog.DoPrintf("DevBind: Shutting down goroutine...\n")
				return
			case <-time.After(db.timeout):
				// Fetch network interface information
				go db.updatePCILines()
				go db.updateIPRoute()
				go db.updateHWInfo()
			}
		}
	}()

	for len(db.hwInfo) == 0 {
		time.Sleep(time.Second)
	}
}

func (db *DevBind) Stop() {

	db.quit <- true

	time.Sleep(time.Second * 2) // Wait for goroutine to finish

	// Unbind all devices
	db.UnbindPorts()
}

func (db *DevBind) updateHWInfo() {

	lshw := runBashCmd(hwInfoCmd)

	if err := json.Unmarshal(lshw.Bytes(), &db.hwInfo); err != nil {
		fmt.Printf("error unmarshal HwInfo: %s\n", err)
	}

	for _, info := range db.hwInfo {
		drvName := strings.TrimSpace(info.Config.Driver)
		// Store the device information using driver ID and bus ID
		if drvName != "" {
			db.hwDriverMap[strings.TrimSpace(info.Config.Driver)] = info
		}
		pci := strings.TrimPrefix(info.BusInfo, "pci@")
		db.hwBusMap[pci] = info
	}
}

func (db *DevBind) updateIPRoute() {

	routes := runBashCmd(ipRouteCmd)

	if err := json.Unmarshal(routes.Bytes(), &db.ipRoute); err != nil {
		fmt.Printf("error unmarshal IPRoute: %s\n", err)
	}
}

func (db *DevBind) updatePCILines() {

	lspci := runBashCmd(pciLinesCmd)

	// Remove leading and trailing whitespace and split into lines.
	db.pciLines = strings.Split(strings.TrimSpace(lspci.String()), "\n")
}

func (db *DevBind) BindPorts(pciList []*string) error {

	if len(pciList) == 0 {
		return fmt.Errorf("no ports specified")
	}

	for _, pciAddr := range pciList {
		pci := *pciAddr
		if !strings.HasPrefix(pci, "0000:") {
			pci = "0000:" + *pciAddr // prepend 0000: to make it a valid PCI address
		}
		if err := db.BindPort(pci); err != nil {
			return err
		}
	}

	return nil
}

/*
sudo ./dpdk-devbind.py -b vfio-pci 86:00.0
unbind_one: /sys/bus/pci/drivers/i40e/unbind = 0000:86:00.0
1 bind_one: /sys/bus/pci/devices/0000:86:00.0/driver_override = vfio-pci
2 bind_one: /sys/bus/pci/drivers/vfio-pci/bind = 0000:86:00.0
3 bind_one: /sys/bus/pci/devices/0000:86:00.0/driver_override = 0
*/
func (db *DevBind) BindPort(pciAddr string) error {

	if v, ok := db.hwBusMap[pciAddr]; ok {
		b := bindInfo{
			BusInfo:        strings.TrimPrefix(v.BusInfo, "pci@"),
			LogicalName:    v.LogicalName,
			Driver:         v.Config.Driver,
			OriginalDriver: v.Config.Driver,
		}
		db.pciAddrs[pciAddr] = b

		// Unbind the pci device if not bound to vfio-pci
		if len(v.Config.Driver) > 0 && v.Config.Driver != "vfio-pci" {
			if err := db.unbind(b.BusInfo, b.Driver); err != nil {
				return err
			}
		} else {
			fmt.Printf("PCI address %s already bound to vfio-pci\n", pciAddr)
			return nil
		}
		// Override the driver
		if err := db.override(b.BusInfo, "vfio-pci"); err != nil {
			return err
		}

		// Bind device to vfio-pci
		if err := db.bind("vfio-pci", b.BusInfo); err != nil {
			return err
		}

		// Override the driver
		if err := db.override(b.BusInfo, ""); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("%s not found in hardware information", pciAddr)
	}

	return nil
}

func (db *DevBind) UnbindPorts() error {
	if len(db.pciAddrs) == 0 {
		return fmt.Errorf("no ports specified")
	}

	for _, v := range db.pciAddrs {
		if err := db.UnbindPort("vfio-pci", v.OriginalDriver, v.BusInfo); err != nil {
			tlog.DoPrintf("error unbinding device %s from vfio-pci\n", v.BusInfo)
			return err
		}
	}

	return nil
}

/*
sudo ./dpdk-devbind.py -b i40e 86:00.0
unbind_one: /sys/bus/pci/drivers/vfio-pci/unbind = 0000:86:00.0
2 bind_one: /sys/bus/pci/drivers/i40e/bind = 0000:86:00.0
3 bind_one: /sys/bus/pci/devices/0000:86:00.0/driver_override = 0
*/
func (db *DevBind) UnbindPort(drv, oDrv, bus string) error {

	//	if err := db.unbind(bus, drv); err != nil {
	//		return err
	//	}
	//
	//	if err := db.bind(oDrv, bus); err != nil {
	//		return err
	//	}
	//
	//	if err := db.override(bus, ""); err != nil {
	//		return err
	//	}

	delete(db.pciAddrs, bus)

	return nil
}

func (db *DevBind) PCILines() []string {
	return db.pciLines
}

func (db *DevBind) HwInfo() []*HwInfo {
	return db.hwInfo
}

func (db *DevBind) HwDriverMap() map[string]*HwInfo {
	return db.hwDriverMap
}

func (db *DevBind) HwBusMap() map[string]*HwInfo {
	return db.hwBusMap
}

func (db *DevBind) IPRoute() IPRoute {
	return db.ipRoute
}

func (db *DevBind) override(bus, drv string) error {

	override := fmt.Sprintf(driverOverride, bus)

	// Override the driver
	if drv == "" {
		zero := make([]byte, 1)
		zero[0] = 0
		tlog.DoPrintf("Override: %s = 0(%d)\n", override, len(zero))
		if err := os.WriteFile(override, zero, 0644); err != nil {
			return err
		}
	} else {
		tlog.DoPrintf("Override: %s = %s(%d)\n", override, drv, len(drv))

		if err := os.WriteFile(override, []byte(drv), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (db *DevBind) bind(drv, bus string) error {

	bind := fmt.Sprintf(driverBind, drv)

	tlog.DoPrintf("    Bind: %s = %s\n", bind, bus)

	// Bind the driver
	if err := os.WriteFile(bind, []byte(bus), 0644); err != nil {
		return err
	}
	return nil
}

func (db *DevBind) unbind(bus, drv string) error {

	bind := fmt.Sprintf(driverUnbind, drv)

	tlog.DoPrintf("  Unbind: %s = %s\n", bind, bus)

	// Unbind the driver
	if err := os.WriteFile(bind, []byte(bus), 0644); err != nil {
		return err
	}
	return nil
}

func runBashCmd(args ...string) bytes.Buffer {

	// Add -c flag to run bash command in a sub-shell
	cmds := make([]string, 0, len(args)+1)
	cmds = append(cmds, "-c")
	cmds = append(cmds, args...)

	fmt.Printf("Running command: bash %s\n", strings.Join(cmds, " "))

	// Create a new command
	cmd := exec.Command(shellCmd, cmds...)

	// Buffers to capture stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Start the command
	if err := cmd.Start(); err != nil {
		return stderrBuf
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return stderrBuf
	}
	if stderrBuf.String() != "" {
		return stderrBuf
	}

	// Return stdout buffer
	return stdoutBuf
}
