// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package gopktgen

import (
	"time"

	gpc "github.com/pktgen/go-pktgen/internal/gpcommon"
)

type PortStats struct {
	TimeStamp  time.Time          // Current time per port to calculate rates
	LinkStatus gpc.LinkState      // Link status of the port
	CurrStats  gpc.PortStatistics // Port statistics
	PrevStats  gpc.PortStatistics // Port statistics
	RateStats  gpc.PortStatistics // Port statistics

	RxPPS        uint64  // Rx packets per second
	TxPPS        uint64  // Tx packets per second
	RxMbitsPS    uint64  // Rx Mbits per second
	TxMbitsPS    uint64  // Tx Mbits per second
	MaxRxPPS     uint64  // Maximum Rx packets received per second
	MaxTxPPS     uint64  // Maximum Tx packets transmitted per second
	MaxRxMbitPS  uint64  // Maximum Rx Mbits per second per second
	MaxTxMbitPS  uint64  // Maximum Tx Mbits per second per second
	RxPacketRate float64 // Rx percent rate per second
	TxPacketRate float64 // Tx percent rate per second
}

func (g *GoPktgen) NewPortStats() *PortStats {
	return &PortStats{
		TimeStamp:  time.Now(),
		LinkStatus: gpc.LinkState{},
		CurrStats:  gpc.PortStatistics{Ether: &gpc.EtherStats{}, Packet: &gpc.PacketStats{}},
		PrevStats:  gpc.PortStatistics{Ether: &gpc.EtherStats{}, Packet: &gpc.PacketStats{}},
		RateStats:  gpc.PortStatistics{Ether: &gpc.EtherStats{}, Packet: &gpc.PacketStats{}},
	}
}

func (g *GoPktgen) UpdateStats() {

	for i, d := range g.portStats {
		g.calculateRates(gpc.PortID(i), d)
	}
}

func rateUpdate(curr, prev uint64, secs float64) uint64 {
	var rate uint64

	diff := curr - prev
	if diff > 0 {
		rate = diff
	} else {
		rate = 0
	}

	return uint64(float64(rate) / secs)
}

func (g *GoPktgen) calculateRates(pid gpc.PortID, d *PortStats) {

	if d.TimeStamp.IsZero() {
		d.TimeStamp = time.Now()
	}

	now := time.Now()
	seconds := now.Sub(d.TimeStamp).Seconds()
	if seconds < 1.0 {
		seconds = 1.0
	}
	d.TimeStamp = now

	d.LinkStatus = g.GetLinkState(pid)
	d.CurrStats.Ether = g.GetEtherStats(pid)
	d.CurrStats.Packet = g.GetPacketStats(pid)

	c := d.CurrStats
	p := d.PrevStats
	r := d.RateStats

	r.Ether.IPackets = rateUpdate(c.Ether.IPackets, p.Ether.IPackets, seconds)
	r.Ether.OPackets = rateUpdate(c.Ether.OPackets, p.Ether.OPackets, seconds)
	r.Ether.IBytes = rateUpdate(c.Ether.IBytes, p.Ether.IBytes, seconds)
	r.Ether.OBytes = rateUpdate(c.Ether.OBytes, p.Ether.OBytes, seconds)
	r.Ether.IMissed = rateUpdate(c.Ether.IMissed, p.Ether.IMissed, seconds)
	r.Ether.IErrors = rateUpdate(c.Ether.IErrors, p.Ether.IErrors, seconds)
	r.Ether.OErrors = rateUpdate(c.Ether.OErrors, p.Ether.OErrors, seconds)
	r.Ether.RxNombuf = rateUpdate(c.Ether.RxNombuf, p.Ether.RxNombuf, seconds)

	if r.Ether.IPackets > 0 {
		d.RxPacketRate = (float64(r.Ether.IPackets) / float64(d.LinkStatus.MaxPktsPerSec())) * 100.0
	} else {
		d.RxPacketRate = 0
	}
	if r.Ether.OPackets > 0 {
		d.TxPacketRate = (float64(r.Ether.OPackets) / float64(d.LinkStatus.MaxPktsPerSec())) * 100.0
	} else {
		d.TxPacketRate = 0
	}

	for q := uint16(0); q < gpc.EtherStatsQueueCntrs; q++ {

		r.Ether.QIPackets[q] = rateUpdate(c.Ether.QIPackets[q], p.Ether.QIPackets[q], seconds)
		r.Ether.QOPackets[q] = rateUpdate(c.Ether.QOPackets[q], p.Ether.QOPackets[q], seconds)
		r.Ether.QIBytes[q] = rateUpdate(c.Ether.QIBytes[q], p.Ether.QIBytes[q], seconds)
		r.Ether.QOBytes[q] = rateUpdate(c.Ether.QOBytes[q], p.Ether.QOBytes[q], seconds)
		r.Ether.QErrors[q] = rateUpdate(c.Ether.QErrors[q], p.Ether.QErrors[q], seconds)
	}

	d.RxMbitsPS = (r.Ether.IBytes * 8) / 1000000
	d.TxMbitsPS = (r.Ether.OBytes * 8) / 1000000

	if d.MaxRxPPS < r.Ether.IPackets {
		d.MaxRxPPS = r.Ether.IPackets
	}
	if d.MaxTxPPS < r.Ether.OPackets {
		d.MaxTxPPS = r.Ether.OPackets
	}

	d.PrevStats = d.CurrStats // Save previous stats
}

func (g *GoPktgen) GetPortInfo(pid gpc.PortID) *PortStats {
	if pid >= gpc.PortID(len(g.portStats)) {
		return nil
	}
	return g.portStats[pid]
}

func (g *GoPktgen) GetRxPercentSlice() []float64 {
	percentSlice := make([]float64, 0)

	for _, d := range g.portStats {
		percentSlice = append(percentSlice, d.RxPacketRate)
	}
	return percentSlice
}

func (g *GoPktgen) GetTxPercentSlice() []float64 {
	percentSlice := make([]float64, 0)

	for _, d := range g.portStats {
		percentSlice = append(percentSlice, d.TxPacketRate)
	}
	return percentSlice
}

func (g *GoPktgen) GetPortStats(pid uint16) *PortStats {
	return g.portStats[gpc.PortID(pid)]
}
