// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package main

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/pktgen/go-pktgen/pkgs/kview"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/pktgen/go-pktgen/internal/constants"
	"github.com/pktgen/go-pktgen/internal/iobind"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	tab "github.com/pktgen/go-pktgen/internal/taborder"
	"github.com/pktgen/go-pktgen/internal/tlog"
	vp "github.com/pktgen/go-pktgen/internal/vpanels"
)

// Display system device configuration

// PanelSystem - Data for main page information
type PanelSystem struct {
	flex0         *kview.Flex
	host          *kview.TextView
	mem           *kview.TextView
	netInfo       *kview.Table
	netStats      *kview.Table
	netPCI        *kview.Table
	rowNetCount   int // Number of current rows in network view
	rowStatsCount int // Number of current rows in stats view
}

const (
	sysinfoPanelName       string = "System"
	sysinfoPanelLogID      string = "SysInfoLogID"
	sysinfoHelpID          string = "SysInfoHelpID"
	sysinfoHelpText        string = "SysInfo Mode Text, press Esc to close."
	sysinfoHostTabOrderID  string = "SysinfoHostTabOrderID"
	sysinfoMemTabOrderID   string = "SysinfoMemTabOrderID"
	sysinfoNetTabOrderID   string = "SysinfoNetTabOrderID"
	sysinfoStatsTabOrderID string = "SysinfoStatsTabOrderID"
	sysinfoPCITabOrderID   string = "SysinfoPCITabOrderID"
	sysinfoHostTabKey      rune   = 'h'
	sysinfoMemTabKey       rune   = 'm'
	sysinfoNetTabKey       rune   = 'n'
	sysinfoStatsTabKey     rune   = 's'
	sysinfoPCITabKey       rune   = 'd'
)

func SysInfoPanel() (string, vp.VPanelFunc) {
	return sysinfoPanelName, sysInfoPanelSetup
}

// SysInfoPanelSetup setup the main cpu page
func sysInfoPanelSetup(cfg vp.VPanelConfig) (*vp.VPanelInfo, error) {

	tlog.Register(sysinfoPanelLogID)

	ps := &PanelSystem{
		flex0: kview.NewFlex(),
	}
	ps.flex0.SetDirection(kview.FlexRow)

	hlp.TitleBox(ps.flex0, hlp.CommandInfo(true))

	flex1 := kview.NewFlex()
	flex1.SetDirection(kview.FlexColumn)
	ps.flex0.AddItem(flex1, 10, 0, true)

	ps.sysInfoHostMemView(flex1)

	ps.sysInfoHostNetView()
	ps.sysInfoNetStatsView()
	ps.sysInfoNetDevView()
	ps.sysInfoHelpSetup(cfg)
	ps.sysInfoTabOrderSetup(cfg)

	return &vp.VPanelInfo{
		PanelName: sysinfoPanelName,
		HelpID:    sysinfoHelpID,
		TopFlex:   ps.flex0,
		TimerFn:   ps.sysInfoTimer(cfg),
	}, nil
}

func (ps *PanelSystem) sysInfoTimer(cfg vp.VPanelConfig) func(int, uint64) {
	return func(step int, ticks uint64) {
		if step == -1 || ps.flex0.HasFocus() {
			cfg.App.QueueUpdateDraw(func() {
				ticks++
				switch step {
				case 0:
					iobind.Update()

				case 1:

				case 2:

				case -1, 3:
					ps.displayHost(ps.host)
					ps.displayMem(ps.mem)
					ps.displayNetInfo(ps.netInfo)
					ps.displayNetStats(ps.netStats)
					ps.displayNetPCI(ps.netPCI)
				}
			})
		}
	}
}

func (ps *PanelSystem) sysInfoHostMemView(f1 *kview.Flex) {

	ps.host = hlp.CreateTextView(f1,
		hlp.NewText(fmt.Sprintf("Host (%c)", sysinfoHostTabKey), kview.AlignLeft), 0, 1, true)
	ps.mem = hlp.CreateTextView(f1,
		hlp.NewText(fmt.Sprintf("Memory (%c)", sysinfoMemTabKey), kview.AlignLeft), 0, 1, false)
}

func (ps *PanelSystem) sysInfoHostNetView() {

	ps.netInfo = hlp.CreateTableView(ps.flex0,
		hlp.NewText(fmt.Sprintf("Host Network Information (%c)", sysinfoNetTabKey), kview.AlignLeft), 0, 1, false)
	ps.netInfo.SetSelectable(false, false)
	ps.netInfo.SetFixed(1, 1)
	ps.netInfo.SetSeparator(kview.Borders.Vertical)
}

func (ps *PanelSystem) sysInfoNetStatsView() {

	ps.netStats = hlp.CreateTableView(ps.flex0,
		hlp.NewText(fmt.Sprintf("Network Statistics (%c)", sysinfoStatsTabKey), kview.AlignLeft), 0, 1, false)
	ps.netStats.SetFixed(1, 1)
	ps.netStats.SetSeparator(kview.Borders.Vertical)
	ps.netStats.SetScrollBarVisibility(kview.ScrollBarAuto)
	ps.netStats.SetScrollBarColor(tcell.ColorCornflowerBlue)
}

func (ps *PanelSystem) sysInfoNetDevView() {

	ps.netPCI = hlp.CreateTableView(ps.flex0,
		hlp.NewText(fmt.Sprintf("Network Devices (%c)", sysinfoPCITabKey), kview.AlignLeft), 0, 1, false)
	ps.netPCI.SetFixed(1, 0)
	ps.netPCI.SetSeparator(kview.Borders.Vertical)
	ps.netPCI.SetScrollBarVisibility(kview.ScrollBarAuto)
	ps.netPCI.SetScrollBarColor(tcell.ColorCornflowerBlue)
}

func (ps *PanelSystem) sysInfoHelpSetup(cfg vp.VPanelConfig) {

	modal := kview.NewModal()
	modal.SetText(sysinfoHelpText)
	modal.AddButtons([]string{"Got it"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		cfg.Panels.HidePanel(sysinfoHelpID)
	})
	cfg.Panels.AddPanel(sysinfoHelpID, modal, false, false)
}

func (ps *PanelSystem) sysInfoTabOrderSetup(cfg vp.VPanelConfig) error {

	tabData := []tab.TabData{
		{Name: sysinfoHostTabOrderID, View: ps.host, Key: sysinfoHostTabKey},
		{Name: sysinfoMemTabOrderID, View: ps.mem, Key: sysinfoMemTabKey},
		{Name: sysinfoNetTabOrderID, View: ps.netInfo, Key: sysinfoNetTabKey},
		{Name: sysinfoStatsTabOrderID, View: ps.netStats, Key: sysinfoStatsTabKey},
		{Name: sysinfoPCITabOrderID, View: ps.netPCI, Key: sysinfoPCITabKey},
	}

	if _, err := hlp.CreateTabOrder(cfg.App, sysinfoPanelName, tabData); err != nil {
		return err
	}

	return nil
}

// Display the Host information
func (ps *PanelSystem) displayHost(view *kview.TextView) {

	str := ""
	info, _ := host.Info()
	str += fmt.Sprintf("Hostname: %s\n", cz.Yellow(info.Hostname))
	str += fmt.Sprintf("Host ID : %s\n", cz.Green(info.HostID))

	c := cases.Title(language.AmericanEnglish)
	str += fmt.Sprintf("OS      : %s-%s\n",
		cz.GoldenRod(c.String(info.OS)), cz.Orange(c.String(info.KernelVersion)))
	str += fmt.Sprintf("Platform: %s %s\nFamily  : %s\n",
		cz.MediumSpringGreen(c.String(info.Platform)),
		cz.LightSkyBlue(c.String(info.PlatformVersion)),
		cz.Green(c.String(info.PlatformFamily)))

	days := info.Uptime / (60 * 60 * 24)
	hours := (info.Uptime - (days * 60 * 60 * 24)) / (60 * 60)
	minutes := ((info.Uptime - (days * 60 * 60 * 24)) - (hours * 60 * 60)) / 60
	s := fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
	str += fmt.Sprintf("Uptime  : %s\n", cz.DeepPink(s))

	role := info.VirtualizationRole
	if len(role) == 0 {
		role = "unknown"
	}
	vsys := info.VirtualizationSystem
	if len(vsys) == 0 {
		vsys = "unknown"
	}
	str += fmt.Sprintf("Virtual Role: %s, System: %s\n", cz.Yellow(role), cz.Yellow(vsys))
	str += fmt.Sprintf("Go-Pktgen Build Date: %s", cz.MediumSpringGreen(hlp.BuildDate()))

	view.SetText(str)
}

// Display the information about the memory in the system
func (ps *PanelSystem) displayMem(view *kview.TextView) {

	str := ""

	v, _ := mem.VirtualMemory()

	str += fmt.Sprintf("Memory  Total: %s MiB\n", cz.Green(v.Total/constants.MegaBytes, 6))
	str += fmt.Sprintf("         Free: %s MiB\n", cz.Green(v.Free/constants.MegaBytes, 6))
	str += fmt.Sprintf("         Used: %s Percent\n\n", cz.Orange(v.UsedPercent, 6, 1))

	str += fmt.Sprintf("%s:\n", cz.MediumSpringGreen("Total Hugepage Info"))
	str += fmt.Sprintf("   Free/Total: %s/%s pages\n", cz.LightBlue(v.HugePagesFree, 6),
		cz.LightBlue(v.HugePagesTotal, 6))
	str += fmt.Sprintf("Hugepage Size: %s Kb", cz.LightBlue(v.HugePageSize/constants.KiloBytes, 6))

	view.SetText(str)
}

// Display the Host network information
func (ps *PanelSystem) displayNetInfo(view *kview.Table) {

	view.Clear()

	row := 0
	col := 0

	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("network interfaces: %s\n", err)
		return
	}
	sort.SliceStable(ifaces, func(i, j int) bool {
		return ifaces[i].Name < ifaces[j].Name
	})

	titles := []hlp.TextInfo{
		hlp.NewText(cz.Yellow("Name", -20), kview.AlignLeft),
		hlp.NewText(cz.Yellow("IP Address", 20), kview.AlignRight),
		hlp.NewText(cz.Yellow("MTU", 8), kview.AlignRight),
		hlp.NewText(cz.Yellow("Interface Flags", 32), kview.AlignRight),
		hlp.NewText(cz.Yellow("MAC Address", 20), kview.AlignRight),
	}
	row = hlp.TableSetHeaders(view, 0, 0, titles)

	setCell := func(row, col int, value string, left bool) int {
		align := kview.AlignRight
		if left {
			align = kview.AlignLeft
		}
		tableCell := kview.NewTableCell(value)
		tableCell.SetAlign(align)
		tableCell.SetSelectable(false)
		ps.netInfo.SetCell(row, col, tableCell)
		col++

		return col
	}

	for _, f := range ifaces {
		col = setCell(row, 0, cz.LightBlue(fmt.Sprintf("%-20s", f.Name)), true)
		if len(f.Addrs) > 0 {
			col = setCell(row, col, cz.Orange(f.Addrs[0].Addr), false)
		} else {
			col = setCell(row, col, " ", false)
		}
		col = setCell(row, col, cz.MediumSpringGreen(f.MTU), false)

		col = setCell(row, col, cz.LightSkyBlue(f.Flags), false)
		setCell(row, col, cz.Cyan(f.HardwareAddr), false)

		row++
	}

	for ; row < view.GetRowCount(); row++ {
		view.RemoveRow(row)
	}
	if ps.rowNetCount != view.GetRowCount() {
		ps.rowNetCount = view.GetRowCount()
		view.ScrollToBeginning()
	}
}

// Display the Host network statistics
func (ps *PanelSystem) displayNetStats(view *kview.Table) {

	row := 0
	col := 0

	view.Clear()
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("network interfaces: %s\n", err)
		return
	}
	sort.SliceStable(ifaces, func(i, j int) bool {
		return ifaces[i].Name < ifaces[j].Name
	})

	titles := []hlp.TextInfo{
		hlp.NewText(cz.Yellow("Name", -20), kview.AlignLeft),
		hlp.NewText(cz.Yellow("RX Pkts", 14), kview.AlignRight),
		hlp.NewText(cz.Yellow("TX Pkts", 14), kview.AlignRight),
		hlp.NewText(cz.Yellow("RX Err", 14), kview.AlignRight),
		hlp.NewText(cz.Yellow("TX Err", 14), kview.AlignRight),
		hlp.NewText(cz.Yellow("RX Drop", 14), kview.AlignRight),
		hlp.NewText(cz.Yellow("Tx Drop", 14), kview.AlignRight),
	}
	row = hlp.TableSetHeaders(view, 0, 0, titles)

	setCell := func(row, col int, value string, left bool) int {
		align := kview.AlignRight
		if left {
			align = kview.AlignLeft
		}
		tableCell := kview.NewTableCell(value)
		tableCell.SetAlign(align)
		tableCell.SetSelectable(false)
		ps.netStats.SetCell(row, col, tableCell)
		col++

		return col
	}

	ioCount, err := net.IOCounters(true)
	if err != nil {
		tlog.Printf("Error getting network IO counters: %s\n", err)
		return
	}

	for _, f := range ifaces {
		col = setCell(row, 0, cz.LightBlue(f.Name), true)

		for _, k := range ioCount {
			if k.Name != f.Name {
				continue
			}
			rowData := []hlp.TextInfo{
				hlp.NewText(cz.Wheat(k.PacketsRecv), kview.AlignRight),
				hlp.NewText(cz.Wheat(k.PacketsSent), kview.AlignRight),
				hlp.NewText(cz.Red(k.Errin), kview.AlignRight),
				hlp.NewText(cz.Red(k.Errout), kview.AlignRight),
				hlp.NewText(cz.Red(k.Dropin), kview.AlignRight),
				hlp.NewText(cz.Red(k.Dropout), kview.AlignRight),
			}
			for _, v := range rowData {
				col = hlp.TableCellSet(ps.netStats, row, col, v)
			}
			break
		}

		row++
	}

	for ; row < view.GetRowCount(); row++ {
		view.RemoveRow(row)
	}
	if ps.rowStatsCount != view.GetRowCount() {
		ps.rowStatsCount = view.GetRowCount()
		view.ScrollToBeginning()
	}
}

// Display the Host network PCI devices
func (ps *PanelSystem) displayNetPCI(view *kview.Table) {

	view.Clear()
	titles := []hlp.TextInfo{
		hlp.NewText(cz.Yellow("PCI Address", -13), kview.AlignLeft),
		hlp.NewText(cz.Yellow("Driver", -12), kview.AlignLeft),
		hlp.NewText(cz.Yellow("Module", -12), kview.AlignLeft),
		hlp.NewText(cz.Yellow("Device Info"), kview.AlignLeft),
	}
	row := hlp.TableSetHeaders(view, 0, 0, titles)

	netList := iobind.PciNetList()
	for _, net := range netList {
		rowData := []hlp.TextInfo{
			hlp.NewText(cz.CornSilk(net.Slot), kview.AlignLeft),
			hlp.NewText(net.Driver, kview.AlignLeft),
			hlp.NewText(cz.GoldenRod(net.Module), kview.AlignLeft),
			hlp.NewText(cz.SkyBlue(net.Device), kview.AlignLeft),
		}
		col := 0
		for _, v := range rowData {
			col = hlp.TableCellSet(view, row, col, v)
		}
		row++
	}

	view.ScrollToBeginning()
}
