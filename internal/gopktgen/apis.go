// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package gopktgen

import (
	"fmt"
	"unsafe"

	gpc "github.com/pktgen/go-pktgen/internal/gpcommon"
	"github.com/pktgen/go-pktgen/internal/tlog"
)

func (g *GoPktgen) GetEtherStats(pid gpc.PortID) *gpc.EtherStats {

	stats := &gpc.EtherStats{}

	if g.PortEtherStats(pid, unsafe.Pointer(stats)) < 0 {
		return &gpc.EtherStats{}
	}

	return stats
}

func (g *GoPktgen) GetPacketStats(pid gpc.PortID) *gpc.PacketStats {

	stats := &gpc.PacketStats{}

	if g.PortPacketStats(pid, unsafe.Pointer(stats)) < 0 {
		return &gpc.PacketStats{}
	}

	return stats
}

// LinkState is the function to get link status of a port
func (g *GoPktgen) GetLinkState(pid gpc.PortID) gpc.LinkState {

	val := g.PortLinkStatus(pid)

	speed := uint32(val & 0xFFFFFFFF)
	if speed == 0 || speed == 0xFFFFFFFF {
		return gpc.LinkState{}
	}
	state := uint16((val >> 32) & 0xFFFF)
	link := gpc.LinkState{
		Speed:   speed,
		Duplex:  (state & 0x0001) != 0,
		AutoNeg: (state & 0x0002) != 0,
		Status:  (state & 0x0004) != 0,
	}

	return link
}

// GetMacAddress is the function to get MAC address of a port
func (g *GoPktgen) GetMacAddress(pid gpc.PortID) (gpc.MacAddress, error) {

	mac := gpc.MacAddress{}

	if g.PortMacAddress(pid, unsafe.Pointer(&mac)) < 0 {
		return mac, fmt.Errorf("error getting MAC address for port %d\n", pid)
	}


	return mac, nil
}

func (g *GoPktgen) GetPortDeviceInfo(pid gpc.PortID) (gpc.PortDeviceInfo, error) {
	info := gpc.PortDeviceInfo{}

	if g.PortDeviceInfo(pid, unsafe.Pointer(&info)) < 0 {
		return info, fmt.Errorf("error getting device info for port %d\n", pid)
	}

	return info, nil
}

func (g *GoPktgen) LaunchThreads() error {

	sendMsg := &gpc.ChannelMsg{
		Action: gpc.LaunchMsgType,
	}

	if ret := g.ChannelSend(g.dpdkChan, unsafe.Pointer(sendMsg), 1); ret < 0 {
		return fmt.Errorf("error launching threads")
	}
	return nil
}

func (g *GoPktgen) L2pConfigSet() error {

	for _, c := range g.l2p.Cores {
		tlog.Printf("Configuring L2p: %+v\n", c)
		cfg := L2pConfig{
			LPortID:     c.LPort.LogicalPortID,
			CoreID:      c.Core,
			CoreMode:    c.Mode,
			RxQid:       c.LPort.RxQid,
			TxQid:       c.LPort.TxQid,
			PortID:      c.LPort.Port.Pid,
			NumRxQueues: c.LPort.Port.NumRxQueues,
			NumTxQueues: c.LPort.Port.NumTxQueues,
		}
		if err := g.L2pConfig(unsafe.Pointer(&cfg)); err < 0 {
			return fmt.Errorf("failed to set L2pConfig for core %v", c.Core)
		}
	}
	return nil
}
