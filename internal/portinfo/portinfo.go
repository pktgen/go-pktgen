// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022-2024 Intel Corporation

package portinfo

type PortStats struct {
	InPackets  uint64 // Number of input packets
	InBytes    uint64 // Number of input bytes
	InErrors   uint64 // Number of input errors
	InMissed   uint64 // Number of missed input packets
	RxInvalid  uint64 // Number of invalid Rx packets
	OutPackets uint64 // Number of output packets
	OutBytes   uint64 // Number of output bytes
	OutErrors  uint64 // Number of output errors
	OutDropped uint64 // Number of dropped output packets
	TxInvalid  uint64 // Number of invalid Tx packets

	RxPacketRate uint64 // Receive Packets per second
	TxPacketRate uint64 // Transmit Packets per second
	RxMbits      uint64 // Receive Mega-bits per second
	TxMbits      uint64 // Transmit Mega-bits per second
	RxMaxPPS     uint64 // Maximum Rx packets per second
	TxMaxPPS     uint64 // Maximum Tx packets per second

	Broadcast      uint64 // Number of broadcast packets received
	Multicast      uint64 // Number of multicast packets received
	Size64         uint64 // Number of 64-bytes packets received
	Size65To127    uint64 // Number of 65-127-byte packets received
	Size128To255   uint64 // Number of 128-255-byte packets received
	Size256To511   uint64 // Number of 256-511-byte packets received
	Size512To1023  uint64 // Number of 512-1023-byte packets received
	Size1024To1518 uint64 // Number of 1024-1518-byte packets received
	RuntPackets    uint64 // Number of Runt packets received
	JumboPackets   uint64 // Number of jumbo packets received
}

type PortData struct {
	PortID       int        // Port ID
	LinkState    string     // Link state i.e., FD-10000-UP
	Stats        *PortStats // Port statistics
	RxPkts       uint64
	TxPkts       uint64
	TotalRxPkts  uint64
	TotalTxPkts  uint64
	TotalRxMbits uint64
	TotalTxMbits uint64
}

type PortInfo struct {
	dataArr []*PortData
}

func New(portCnt int) *PortInfo {
	ps := &PortInfo{
		dataArr: make([]*PortData, portCnt),
	}

	for i := 0; i < portCnt; i++ {
		ps.dataArr[i] = &PortData{
			PortID:    i,
			LinkState: "Down",
			Stats:     &PortStats{},
		}
	}

	return ps
}

func (p *PortInfo) GetPortStats(portID int) *PortStats {
	return p.dataArr[portID].Stats
}

func (p *PortInfo) GetPortDataArray() []*PortData {
	return p.dataArr
}

func (p *PortInfo) ClearPortStats(portID int) {
	p.dataArr[portID].Stats = &PortStats{}
}

func (p *PortInfo) ClearAllPortStats() {
	for i := 0; i < len(p.dataArr); i++ {
		p.ClearPortStats(i)
	}
}

func (p *PortInfo) CollectStats() {

	for i := 0; i < len(p.dataArr); i++ {
		s := p.GetPortStats(i)
		s.InPackets = 0
	}
}
