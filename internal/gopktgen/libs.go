// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package gopktgen

import (
	"fmt"

	"github.com/ebitengine/purego"
	"github.com/pktgen/go-pktgen/internal/tlog"
)

func (g *GoPktgen) openLibs() error {

	// Load the libraries in the order they are listed above.
	for _, lib := range g.libList {
		if handle, err := purego.Dlopen(lib.Name, purego.RTLD_NOW|purego.RTLD_GLOBAL); err != nil {
			return fmt.Errorf("error loading %s: %v", lib.Name, err)
		} else {
			tlog.Printf("Library %s loaded successfully\n", lib.Name)
			lib.Handle = handle
		}
	}

	return nil
}

func (g *GoPktgen) closeLibs() {
	// Close all libraries in the order they were added.
	for _, lib := range g.libList {
		if lib.Handle == 0 {
			continue
		}
		if err := purego.Dlclose(lib.Handle); err != nil {
			tlog.Printf("Library %s closed successfully\n", lib.Name)
		}
	}
}

func (g *GoPktgen) loadAPIs() error {

	// Add new API functions to be registered here.
	gpktApis := []ApiInfo{
		{FuncName: "dpdk_add_argv", FuncPtr: &g.AddArgv},
		{FuncName: "dpdk_l2p_config", FuncPtr: &g.L2pConfig},
		{FuncName: "dpdk_startup", FuncPtr: &g.Startup},
		{FuncName: "dpdk_shutdown", FuncPtr: &g.Shutdown},

		{FuncName: "port_set_info", FuncPtr: &g.PortSetInfo},
		{FuncName: "port_get_info", FuncPtr: &g.PortGetInfo},
		{FuncName: "port_free_info", FuncPtr: &g.PortFreeInfo},
		{FuncName: "port_ether_stats", FuncPtr: &g.PortEtherStats},
		{FuncName: "port_packet_stats", FuncPtr: &g.PortPacketStats},
		{FuncName: "port_link_status", FuncPtr: &g.PortLinkStatus},
		{FuncName: "port_mac_address", FuncPtr: &g.PortMacAddress},
		{FuncName: "port_device_info", FuncPtr: &g.PortDeviceInfo},

		{FuncName: "mc_create", FuncPtr: &g.ChannelCreate},
		{FuncName: "mc_attach", FuncPtr: &g.ChannelAttach},
		{FuncName: "mc_destroy", FuncPtr: &g.ChannelDestroy},
		{FuncName: "mc_recv", FuncPtr: &g.ChannelRecv},
		{FuncName: "mc_send", FuncPtr: &g.ChannelSend},
		// Add more API functions here...
	}

	tlog.Printf("Registering API functions...\n")
	for _, v := range gpktApis {
		tlog.Printf("   %s\n", v.FuncName)

		// Use uintptr(0) for library handler pointer to use running program symbols.
		purego.RegisterLibFunc(v.FuncPtr, uintptr(0), v.FuncName)
	}
	return nil
}
