// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package cpudata

import (
	"log"
	"strconv"
	"sync"

	"github.com/shirou/gopsutil/cpu"
)

// CPUData - CPU information data
type CPUData struct {
	cpuInfo         []cpu.InfoStat
	numLogical      int16
	numPhysical     int16
	numSockets      int16
	numHyperThreads int16
	cores           []uint16
	sockets         []uint16
	coreMap         map[uint16][]uint16
}

var cpuData *CPUData

// Open - Open the CPU information
func setupCPUData() error {

	cd := &CPUData{}
	cpuData = cd

	info, err := cpu.Info()
	if err != nil {
		return err
	}
	cd.cpuInfo = info

	cd.coreMap = make(map[uint16][]uint16)
	cd.cores = []uint16{}
	cd.sockets = []uint16{}

	for lcore, c := range info {
		core, _ := strconv.Atoi(c.CoreID)

		// If the core is found in the list of cores then append that core to
		// a list for cores
		if !uint16InSlice(uint16(core), cd.cores) {
			cd.cores = append(cd.cores, uint16(core))
		}

		// If the socket id is found in the list of sockets then append that socket to
		// a list for sockets
		socket, _ := strconv.Atoi(c.PhysicalID)
		if !uint16InSlice(uint16(socket), cd.sockets) {
			cd.sockets = append(cd.sockets, uint16(socket))
		}

		key := uint16((socket << 8) | core)

		// Add the core to the core map
		_, ok := cd.coreMap[key]
		if !ok {
			cd.coreMap[key] = []uint16{}
		}
		cd.coreMap[key] = append(cd.coreMap[key], uint16(lcore))
	}

	// Calculate the cores, sockets and logical cores in the system
	numLogical, _ := cpu.Counts(true)
	numPhysical, _ := cpu.Counts(false)

	cd.numSockets = int16(len(cd.sockets))
	cd.numLogical = int16(numLogical)
	cd.numPhysical = (int16(numPhysical) / cd.numSockets)
	cd.numHyperThreads = (cd.numLogical / (cd.numPhysical * cd.numSockets))

	return nil
}

func GetCPUData() *CPUData {
	var once sync.Once
	once.Do(func() {
		if err := setupCPUData(); err != nil {
			log.Fatalf("Error setting up CPU data: %v", err)
		}
	})
	return cpuData
}

// locate the cpu id or physical id in the slice
func uint16InSlice(b uint16, lst []uint16) bool {

	for _, v := range lst {
		if v == b {
			return true
		}
	}
	return false
}

func (cd *CPUData) CpuInfoList() []cpu.InfoStat {

	return cd.cpuInfo
}

func (cd *CPUData) CpuInfo(lcore uint16) cpu.InfoStat {

	return cd.cpuInfo[lcore]
}

func (cd *CPUData) NumLogicalCores() int16 {

	return cd.numLogical
}

func (cd *CPUData) NumPhysicalCores() int16 {

	return cd.numPhysical
}

func (cd *CPUData) NumSockets() int16 {

	return cd.numSockets
}

func (cd *CPUData) NumHyperThreads() int16 {

	return cd.numHyperThreads
}

func (cd *CPUData) Cores() []uint16 {

	return cd.cores
}

func (cd *CPUData) Sockets() []uint16 {

	return cd.sockets
}

func (cd *CPUData) CoreMapItem(key uint16) ([]uint16, bool) {

	v, ok := cd.coreMap[key]

	return v, ok
}
