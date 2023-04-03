// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package gopktgen

import (
	"unsafe"

	gpc "github.com/pktgen/go-pktgen/internal/gpcommon"
)

type GoPktgenOption func(*GoPktgen)

type LibInfo struct {
	Name   string  // Name of the library.
	Handle uintptr // library handle from dlopen().
}

type ApiInfo struct {
	FuncName string // Function Name of the library.
	FuncPtr  any    // Function pointer.
}

// GoPktgen API functions
type GoPktgenApi struct {
	AddArgv        func(arg string) int                               // Add a argv value
	L2pConfig      func(cfg unsafe.Pointer) int                       // Add L2p configuration
	Startup        func(log_path string) int                          // Start function with log_path
	Shutdown       func() int                                         // Stop function
	PortSetInfo    func(portCfg *gpc.PortConfig) int                  // Set port information
	PortGetInfo    func(portID gpc.PortID) *gpc.PortConfig            // Get port information
	PortFreeInfo   func(cfg *gpc.PortConfig)                          // Free port information structure
	PortStats      func(portID gpc.PortID, stats unsafe.Pointer) int  // Get port statistics
	PortLinkStatus func(portID gpc.PortID) uint64                     // Get link status encoded as a uint64
	PortMacAddress func(portID gpc.PortID, mac unsafe.Pointer) int    // Get port MAC address
	PortDeviceInfo func(portID gpc.PortID, info unsafe.Pointer) int   // Get port device information
	ChannelCreate  func(name string, size uint32) uintptr             // MsgChan initialize
	ChannelAttach  func(name string) uintptr                          // MsgChan attach
	ChannelDestroy func(mc uintptr) int                               // MsgChan destroy
	ChannelRecv    func(mc uintptr, data unsafe.Pointer, len int) int // MsgChan receive data burst
	ChannelSend    func(mc uintptr, data unsafe.Pointer, len int) int // MsgChan send data burst
}

type GoPktgen struct {
	GoPktgenApi
	l2p       *gpc.L2p        // L2p instance
	libList   []*LibInfo      // Slice of libraries names to their handles
	portData  []*gpc.PortData // List of port information structures
	portStats []*PortStats    // List of port statistics structures
	basePath  string          // Base path for libraries
	logPath   string          // Log path for tlog_printf function
	dpdkChan  uintptr         // Message channel for GoPktgen to receive and process messages
}
