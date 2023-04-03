// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package iobind

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

const (
	// Paths to file system files for driver control, bind, unbind, and driver override
	driverBind     = "/sys/bus/pci/drivers/%s/bind"
	driverUnbind   = "/sys/bus/pci/drivers/%s/unbind"
	driverOverride = "/sys/bus/pci/devices/%s/driver_override"

	shellCmd   = "bash"       // The shell command to execute or path to the shell command
	lspciCmd   = "lspci"      // Command to retrieve Network PCI data
	ioBindTool = "bin/iobind" // The IOBind program to bind/unbind devices
)

type BindIO struct {
	netMap    map[string]*NetPCI // Map of PCI network information
	ioBindCmd string             // Path to IOBind command
}

/*
 * Convert lspci -Dvmmnnk output to a list and map for network interfaces.
 * Parse following format:
 * Slot:	    0000:86:00.0
 * Class:	    Ethernet controller [0200]
 * Vendor:	    Intel Corporation [8086]
 * Device:	    Ethernet Controller XL710 for 40GbE QSFP+ [1583]
 * SVendor:	    Intel Corporation [8086]
 * SDevice:	    Ethernet Converged Network Adapter XL710-Q2 [0001]
 * Rev:	        02
 * ProgIf:	    00
 * Driver:	    vfio-pci
 * Module:	    i40e
 * NUMANode:	1
 * IOMMUGroup:	8
 */
type NetPCI struct {
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

type IOBindOption func(*BindIO)

var db *BindIO

func init() {
	db = &BindIO{}
}

func WithIOBindCmd(path string) IOBindOption {

	return func(db *BindIO) {
		db.ioBindCmd = path
	}
}

func New(options ...IOBindOption) *BindIO {

	db.netMap = make(map[string]*NetPCI)
	db.ioBindCmd = ioBindTool

	// Process the option function calls
	for _, f := range options {
		f(db)
	}

	if err := db.parsePciInfo(); err != nil {
		return nil
	}

	return db
}

func (db *BindIO) parsePciInfo() error {
	var lines []string

	if hwInfoStr, err := db.runBashCmd(lspciCmd, "-Dvmmnnk"); err != nil {
		return fmt.Errorf("netInfo: error running %s: %v", lspciCmd, err)
	} else {
		lines = strings.Split(hwInfoStr, "\n")
	}

	db.netMap = make(map[string]*NetPCI)

	slot := ""
	hw := &NetPCI{}
	for _, line := range lines {
		line = strings.TrimSpace(line)

		s := strings.Split(line, ":")[0]
		switch s {
		case "Slot":
			slot = strings.TrimSpace(line[6:])
		case "Class":
			if strings.Contains(line, "Ethernet controller") {
				hw = &NetPCI{Slot: slot, Class: strings.TrimSpace(line[7:])}
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
				hw = &NetPCI{}
			}
		default:
		}
	}
	return nil
}

func Update() error {

	return db.parsePciInfo()
}

func IOBindPorts(devices []string) error {

	if len(devices) == 0 {
		return fmt.Errorf("no ports specified")
	}

	cwd, _ := os.Getwd()
	cmd := fmt.Sprintf("%s/%s", cwd, db.ioBindCmd)
	cmds := []string{"-b"}
	for _, bus := range devices {
		cmds = append(cmds, bus)
	}
	if _, err := db.runBashCmd(cmd, cmds...); err != nil {
		return fmt.Errorf("error executing %v command: %v\n", cmd, err)
	}
	return nil
}

func IOUnbindPorts(devices []string) error {

	if len(devices) == 0 {
		return fmt.Errorf("no ports specified")
	}

	cwd, _ := os.Getwd()
	cmd := fmt.Sprintf("%s/%s", cwd, db.ioBindCmd)
	cmds := []string{"-u"}
	for _, bus := range devices {
		cmds = append(cmds, bus)
	}
	fmt.Printf("Unbinding devices: %s, %v\n", cmd, cmds)
	if _, err := db.runBashCmd(cmd, cmds...); err != nil {
		return fmt.Errorf("error executing %s: %v\n", cmd, err)
	}
	return nil
}

func BindPorts(devices []string) error {

	if len(devices) == 0 {
		return fmt.Errorf("no ports specified")
	}

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
func (db *BindIO) bindPort(v *NetPCI) error {

	// Unbind the pci device if not bound to vfio-pci
	if v.Driver != "" && v.Driver != "vfio-pci" {
		if err := db.unbind(v.Driver, v.Slot); err != nil {
			return err
		}
	} else {
		fmt.Printf("PCI address %s already bound to vfio-pci\n", v.Slot)
		return nil
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

func UnbindPorts(devices []string) error {

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
func (db *BindIO) unbindPort(v *NetPCI) error {

	if v.Driver == v.Module {
		return fmt.Errorf("driver and module are the same %q, not overriding\n", v.Driver)
	}

	if v.Driver != "" {
		if err := db.unbind(v.Driver, v.Slot); err != nil {
			return err
		}
	}

	time.Sleep(time.Second) // Wait for device to unbind
	if err := db.bind(v.Module, v.Slot); err != nil {
		return err
	}

	if err := db.override(v.Slot, ""); err != nil {
		return err
	}

	delete(db.netMap, v.Slot)

	return nil
}

func PciNetList() []NetPCI {

	list := make([]NetPCI, 0, len(db.netMap))
	for _, v := range db.netMap {
		list = append(list, *v)
	}
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Slot < list[j].Slot
	})
	return list
}

func (db *BindIO) override(bus, drv string) error {

	override := fmt.Sprintf(driverOverride, bus)

	val := []byte("\000") // Initialize with null byte

	// Override the driver if provided
	if drv != "" {
		val = []byte(drv)
	}

	if err := writeOnlyFile(override, val); err != nil {
		return err
	}

	return nil
}

func (db *BindIO) bind(drv, bus string) error {

	bind := fmt.Sprintf(driverBind, drv)

	// Bind the driver
	if err := writeOnlyFile(bind, []byte(bus)); err != nil {
		return err
	}
	return nil
}

func (db *BindIO) unbind(drv, bus string) error {

	unbind := fmt.Sprintf(driverUnbind, drv)

	// Unbind the driver
	if err := writeOnlyFile(unbind, []byte(bus)); err != nil {
		return err
	}
	return nil
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

func (db *BindIO) runBashCmd(cmd string, args ...string) (string, error) {

	if path, err := exec.LookPath(cmd); err != nil {
		return "", fmt.Errorf("command not found: %v", err)
	} else {
		if buff, err := exec.Command(path, args...).CombinedOutput(); err != nil {
			return "", err
		} else {
			return string(buff), nil
		}
	}
}
