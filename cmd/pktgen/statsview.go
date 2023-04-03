// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package main

import (
	"fmt"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/pktgen/go-pktgen/pkgs/kview"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	gpc "github.com/pktgen/go-pktgen/internal/gpcommon"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	"github.com/pktgen/go-pktgen/internal/tlog"
)

type StatsView struct {
	portCnt uint16       // Number of ports
	sTable  *kview.Table // Statistics table
	sOnce   sync.Once    // Mutex for stats update
}

func CreateStatsView(portCnt uint16, flex *kview.Flex, tabChar rune) *StatsView {
	sv := &StatsView{portCnt: portCnt}

	s := fmt.Sprintf("Statistics (%c)", tabChar)
	sv.sTable = hlp.CreateTableView(flex, hlp.NewText(s, kview.AlignLeft), 0, 1, true)
	sv.sTable.SetSelectable(false, false)
	sv.sTable.SetFixed(2, 1)
	sv.sTable.SetSeparator(kview.Borders.Vertical)

	return sv
}

func (sv *StatsView) TableView() *kview.Table {
	return sv.sTable
}

func (sv *StatsView) DisplayStats() {

	table := sv.sTable
	table.Clear()

	row := 0
	width := -14
	titles := []hlp.TextInfo{
		hlp.NewText(cz.CornSilk("Link State", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Rx PPS", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Tx PPS", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Rx/Tx Mbits", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Rx Max PPS", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Tx Max PPS", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Rx Missed PPS", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Rx/Tx Errors", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Total Rx Pkts", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Total Tx Pkts", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Broadcast", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Multicast", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan(" 64 Bytes", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan(" 65-127", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan(" 128-255", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan(" 256-511", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan(" 512-1023", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan(" 1024-1518", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Runts/Jumbos", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("ARPs/ICMPs", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Source MAC", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("NUMA/PCIe", width), kview.AlignLeft),
	}

	t := make([]hlp.TextInfo, 0)
	t = append(t, hlp.NewText("", kview.AlignLeft))
	for i := uint16(0); i < sv.portCnt; i++ {
		t = append(t, hlp.NewText(cz.Orange(fmt.Sprintf("Port %2d", i), 14), kview.AlignRight))
	}

	hlp.TableSetHeaders(table, 0, 0, t)
	hlp.TableSetRows(table, 1, 0, titles)

	p := message.NewPrinter(language.English)

	comma := func(n interface{}) string {
		return p.Sprintf("%d", n)
	}

	for v := uint16(0); v < sv.portCnt; v++ {

		ps := pktgenApp.gPkt.GetPortStats(v)

		stats := ps.CurrStats
		rate := ps.RateStats
		link := ps.LinkStatus

		dinfo, err := pktgenApp.gPkt.GetPortDeviceInfo(gpc.PortID(v))
		if err != nil {
			tlog.Printf("Error fetching port %d MAC address: %v\n", v, err)
			dinfo = gpc.PortDeviceInfo{}
		}

		rxtxMbits := fmt.Sprintf("%s/%s", comma(ps.MaxRxMbitPS), comma(ps.MaxTxMbitPS))
		rxtxErrors := fmt.Sprintf("%s/%s", comma(stats.Ether.IErrors+stats.Ether.RxNombuf), comma(stats.Ether.OErrors))
		runtJumbos := fmt.Sprintf("%s/%s", comma(stats.Packet.RuntPackets), comma(stats.Packet.JumboPackets))
		arpIcmps := fmt.Sprintf("%s/%s", comma(stats.Packet.ArpPackets), comma(stats.Packet.IcmpPackets))
		dev := fmt.Sprintf("%d/%s", dinfo.SocketID, string(dinfo.Name[:]))

		rowData := []hlp.TextInfo{
			hlp.NewText(cz.LightYellow(link.String()), kview.AlignRight),        // Link state
			hlp.NewText(cz.Cyan(comma(rate.Ether.IPackets)), kview.AlignRight),  // Received packets
			hlp.NewText(cz.Cyan(comma(rate.Ether.OPackets)), kview.AlignRight),  // Transmitted packets
			hlp.NewText(cz.Wheat(rxtxMbits), kview.AlignRight),                  // Received/transmitted Mbits
			hlp.NewText(cz.Cyan(comma(ps.MaxRxPPS)), kview.AlignRight),          // Max Received packets
			hlp.NewText(cz.Cyan(comma(ps.MaxTxPPS)), kview.AlignRight),          // Max Transmitted packets
			hlp.NewText(cz.Red(comma(rate.Ether.IMissed)), kview.AlignRight),    // Received missed packets
			hlp.NewText(cz.Red(rxtxErrors), kview.AlignRight),                   // Received/transmitted errors
			hlp.NewText(cz.Cyan(comma(stats.Ether.IPackets)), kview.AlignRight), // Total received packets
			hlp.NewText(cz.Cyan(comma(stats.Ether.OPackets)), kview.AlignRight), // Total transmitted packets
			hlp.NewText(cz.Wheat(stats.Packet.Broadcast), kview.AlignRight),     // Broadcast packets
			hlp.NewText(cz.GoldenRod(stats.Packet.Multicast), kview.AlignRight), // Multicast packets
			hlp.NewText(cz.Cyan(stats.Packet.Size64), kview.AlignRight),         // 64-byte packets
			hlp.NewText(cz.Cyan(stats.Packet.Size65To127), kview.AlignRight),    // 65-127-byte packets
			hlp.NewText(cz.Cyan(stats.Packet.Size128To255), kview.AlignRight),   // 128-255-byte packets
			hlp.NewText(cz.Cyan(stats.Packet.Size256To511), kview.AlignRight),   // 256-511-byte packets
			hlp.NewText(cz.Cyan(stats.Packet.Size512To1023), kview.AlignRight),  // 512-1023-byte packets
			hlp.NewText(cz.Cyan(stats.Packet.Size1024To1518), kview.AlignRight), // 1024-1518-byte packets
			hlp.NewText(cz.DeepPink(runtJumbos), kview.AlignRight),              // Runts/Jumbos
			hlp.NewText(cz.Wheat(arpIcmps), kview.AlignRight),                   // ARPs/ICMPs
			hlp.NewText(cz.Wheat(dinfo.MacAddr.String()), kview.AlignRight),     // MAC Address
			hlp.NewText(cz.Wheat(dev), kview.AlignRight),                        // PCIe address
		}

		row = 1
		for _, d := range rowData {
			hlp.TableCellSet(table, row, int(v+1), d)
			row++
		}
	}

	sv.sOnce.Do(func() {
		sv.sTable.ScrollToBeginning()
	})
}
