// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022-2024 Intel Corporation

package devbind

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	// Command to retrieve hardware information
	hwInfoCmd = "lshw -json -c network"

	// Command to retrieve IP route information
	ipRouteCmd = "ip -j -o route"

	// Command to retrieve PCI lines
	pciLinesCmd = "lspci | grep Ether"
	pciHwCmd    = "lspci -Dvmmnnk"

	// Default timeout for commands
	defaultTimeout = 10 * time.Second

	// Paths to file system files for driver control, bind, unbind, and driver override
	driverBind     = "/sys/bus/pci/drivers/%s/bind"
	driverUnbind   = "/sys/bus/pci/drivers/%s/unbind"
	driverOverride = "/sys/bus/pci/devices/%s/driver_override"

	// The shell command to execute or path to the shell command
	shellCmd = "bash"
)

// Slot:	0000:86:00.0
// Class:	Ethernet controller [0200]
// Vendor:	Intel Corporation [8086]
// Device:	Ethernet Controller XL710 for 40GbE QSFP+ [1583]
// SVendor:	Intel Corporation [8086]
// SDevice:	Ethernet Converged Network Adapter XL710-Q2 [0001]
// Rev:	02
// ProgIf:	00
// Driver:	vfio-pci
// Module:	i40e
// NUMANode:	1
// IOMMUGroup:	8
type NetInfo struct {
	Slot       string // PCI address: 0000:3b:00.0
	Class      string // Class: Ethernet controller [0200]
	Vendor     string // Vendor: Intel Corporation [8086]
	Device     string // Device: Ethernet Network Adapter E810-C-Q1 [8086:0001]
	SVendor    string // Subsystem vendor: Intel Corporation [8086]
	SDevice    string // Subsystem device: Ethernet Network Adapter E810-C-Q1 [8086:0001]
	Rev        string // Revision: 02
	ProgIf     string // Programmable interface: 00
	Driver     string // Kernel driver in use: ice
	Module     string // Kernel module: ice
	NumaNode   string // Numa node: 0
	IommuGroup string // IOMMU group: 8
}

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

type DevBind struct {
	Inited   bool                // Flag to indicate if the DevBind object has been initialized
	hwLock   sync.Mutex          // Mutex for accessing hardware information
	timeout  time.Duration       // Timeout for commands
	quit     chan bool           // Channel to signal quit
	done     chan bool           // Channel to signal completion
	shellCmd string              // Path to shell command
	ipRoute  IPRoute             // IP route information
	netMap   map[string]*NetInfo // Map of PCI network information
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

// writeOnlyFile writes data to the named file and error out if not found.
// Since writeOnlyFile requires multiple system calls to complete, a failure mid-operation
// can leave the file in a partially written state.
func writeOnlyFile(name string, data []byte) error {
	f, err := os.OpenFile(name, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

func New(options ...DevBindOption) *DevBind {

	db := &DevBind{
		timeout:  defaultTimeout,
		quit:     make(chan bool),
		done:     make(chan bool),
		shellCmd: shellCmd,
		netMap:   make(map[string]*NetInfo),
	}

	// Process the option function calls
	for _, f := range options {
		f(db)
	}

	db.Inited = true

	return db
}

func (db *DevBind) updateInfo() {

	db.hwLock.Lock()
	defer db.hwLock.Unlock()

	if err := db.netInfo(); err != nil {
		fmt.Printf("Error fetching network information: %v\n", err)
        return
    }
	if err := db.updateIPRoute(); err != nil {
		fmt.Printf("Error fetching IP route information: %v\n", err)
        return
	}
}

func (db *DevBind) Start() {

	// Fetch network interface information
	db.updateInfo()

	go func() {
	ForLoop:
		for {
			select {
			case <-db.quit: // Stop the goroutine
				break ForLoop
			case <-time.After(db.timeout):
				// Fetch network interface information
				db.updateInfo()
			}
		}
		db.done <- true
	}()
}

func (db *DevBind) Stop() {

	db.quit <- true

	<-db.done // Wait for goroutine to finish

	db.Inited = false
}

/*
 * Convert lspci -Dvmmnnk output to a list and map for network interfaces.
 * Parse following format:
 * Slot:	0000:86:00.0
 * Class:	Ethernet controller [0200]
 * Vendor:	Intel Corporation [8086]
 * Device:	Ethernet Controller XL710 for 40GbE QSFP+ [1583]
 * SVendor:	Intel Corporation [8086]
 * SDevice:	Ethernet Converged Network Adapter XL710-Q2 [0001]
 * Rev:	02
 * ProgIf:	00
 * Driver:	vfio-pci
 * Module:	i40e
 * NUMANode:	1
 * IOMMUGroup:	8
 */
func (db *DevBind) netInfo() error {
	if !db.Inited {
		return fmt.Errorf("netInfo: devBind object not initialized")
	}

	hwInfoStr := runBashCmd(pciHwCmd)
	lines := strings.Split(hwInfoStr.String(), "\n")

	slot := ""
	hw := &NetInfo{}
	for _, line := range lines {
		line = strings.TrimSpace(line)

		s := strings.Split(line, ":")[0]
		switch s {
		case "Slot":
			slot = strings.TrimSpace(line[6:])
		case "Class":
			if strings.Contains(line, "Ethernet controller") {
				hw = &NetInfo{Slot: slot, Class: strings.TrimSpace(line[7:])}
			}
			slot = ""
		case "Vendor":
			hw.Vendor = strings.TrimSpace(line[8:])
		case "Device":
			hw.Device = strings.TrimSpace(line[8:])
		case "SVendor":
			hw.SVendor = strings.TrimSpace(line[9:])
		case "SDevice":
			hw.SDevice = strings.TrimSpace(line[9:])
		case "Rev":
			hw.Rev = strings.TrimSpace(line[4:])
		case "ProgIf":
			hw.ProgIf = strings.TrimSpace(line[8:])
		case "Driver":
			hw.Driver = strings.TrimSpace(line[8:])
		case "Module":
			hw.Module = strings.TrimSpace(line[8:])
		case "NUMANode":
			hw.NumaNode = strings.TrimSpace(line[10:])
		case "IOMMUGroup":
			hw.IommuGroup = strings.TrimSpace(line[12:])
		case "":
			if hw.Slot != "" {
				db.netMap[hw.Slot] = hw
				hw = &NetInfo{}
			}
		default:
		}
	}
	return nil
}

func (db *DevBind) updateIPRoute() error {

	if !db.Inited {
		return fmt.Errorf("updateIPRoute: devbind is not initialized")
	}
	routes := runBashCmd(ipRouteCmd)

	if err := json.Unmarshal(routes.Bytes(), &db.ipRoute); err != nil {
		return fmt.Errorf("error unmarshal IPRoute: %v\n", err)
	}
	return nil
}

func (db *DevBind) BindPorts(devices []string) error {

	if !db.Inited {
		return fmt.Errorf("devbind is nit initialized")
	}
	if len(devices) == 0 {
		return fmt.Errorf("no ports specified")
	}

	db.hwLock.Lock()
	defer db.hwLock.Unlock()

	for _, bus := range devices {
		if !strings.HasPrefix(bus, "0000:") {
			bus = "0000:" + bus // prepend 0000: to make it a valid PCI address
		}
		if v, ok := db.netMap[bus]; ok {
			if err := db.bindPort(v); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("device %s not found in hardware information\n", bus)
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
func (db *DevBind) bindPort(v *NetInfo) error {

	// Unbind the pci device if not bound to vfio-pci
	if v.Driver != "" && v.Driver != "vfio-pci" {
		if err := db.unbind(v.Driver, v.Slot); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("PCI address %s already bound to vfio-pci\n", v.Slot)
	}
	// Override the driver
	if err := db.override(v.Slot, "vfio-pci"); err != nil {
		return err
	}

	// Bind device to vfio-pci
	if err := db.bind("vfio-pci", v.Slot); err != nil {
		return err
	}

	// Override the driver
	if err := db.override(v.Slot, ""); err != nil {
		return err
	}

	return nil
}

func (db *DevBind) UnbindPorts(devices []string) error {

	if !db.Inited {
		return fmt.Errorf("devbind is not initialized")
	}

	if len(db.netMap) == 0 {
		return fmt.Errorf("no ports specified")
	}

	for _, bus := range devices {
		if !strings.HasPrefix(bus, "0000:") {
			bus = "0000:" + bus // prepend 0000: to make it a valid PCI address
        }
		if v, ok := db.netMap[bus]; ok {
			if err := db.unbindPort(v); err != nil {
				return err
			}
		} else {
            return fmt.Errorf("device %s not found", bus)
        }
	}

	return nil
}

// sudo ./dpdk-devbind.py -b i40e 86:00.0
// unbind_one: /sys/bus/pci/drivers/vfio-pci/unbind = 0000:86:00.0
// 2 bind_one: /sys/bus/pci/drivers/i40e/bind = 0000:86:00.0
// 3 bind_one: /sys/bus/pci/devices/0000:86:00.0/driver_override = 0
func (db *DevBind) unbindPort(v *NetInfo) error {

	if v.Driver == v.Module {
		return fmt.Errorf("driver and module are the same %q, not overriding\n", v.Driver)
	}

	if err := db.unbind(v.Driver, v.Slot); err != nil {
		return err
	}

	if err := db.bind(v.Module, v.Slot); err != nil {
		return err
	}

	if err := db.override(v.Slot, ""); err != nil {
		return err
	}

	delete(db.netMap, v.Slot)

	return nil
}

func (db *DevBind) NetList() []NetInfo {
	if !db.Inited {
		return nil
	}
	list := make([]NetInfo, 0, len(db.netMap))
	for _, v := range db.netMap {
        list = append(list, *v)
    }
	sort.Slice(list, func(i, j int) bool {
        return list[i].Slot < list[j].Slot
    })
	return list
}

func (db *DevBind) IPRoute() IPRoute {
	if !db.Inited {
		return IPRoute{}
	}
	return db.ipRoute
}

func (db *DevBind) override(bus, drv string) error {

	override := fmt.Sprintf(driverOverride, bus)

	val := []byte("\000")

	// Override the driver
	if drv != "" {
		val = []byte(drv)
	}

	if err := writeOnlyFile(override, val); err != nil {
		return err
	}

	return nil
}

func (db *DevBind) bind(drv, bus string) error {

	bind := fmt.Sprintf(driverBind, drv)

	// Bind the driver
	if err := writeOnlyFile(bind, []byte(bus)); err != nil {
		return err
	}
	return nil
}

func (db *DevBind) unbind(drv, bus string) error {

	unbind := fmt.Sprintf(driverUnbind, drv)

	// Unbind the driver
	if err := writeOnlyFile(unbind, []byte(bus)); err != nil {
		return err
	}
	return nil
}

func runBashCmd(args ...string) bytes.Buffer {

	// Add -c flag to run bash command in a sub-shell
	cmds := make([]string, 0, len(args)+1)
	cmds = append(cmds, "-c")
	cmds = append(cmds, args...)

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
