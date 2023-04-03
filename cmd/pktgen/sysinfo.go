// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pktgen/go-pktgen/pkgs/kview"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/pktgen/go-pktgen/internal/constants"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	tab "github.com/pktgen/go-pktgen/internal/taborder"
	"github.com/pktgen/go-pktgen/internal/tlog"
	vp "github.com/pktgen/go-pktgen/internal/vpanels"
)

// Display system device configuration

// PanelSystem - Data for main page information
type PanelSystem struct {
	topFlex       *kview.Flex
	host          *kview.TextView
	mem           *kview.TextView
	netInfo       *kview.Table
	netStats      *kview.Table
	netPCI        *kview.Table
	to            *tab.Tab
	rowNetCount   int // Number of current rows in network view
	rowStatsCount int // Number of current rows in stats view
}

const (
	sysinfoPanelName  string = "System"
	sysinfoPanelLogID string = "SysInfoLogID"
	sysinfoHelpID     string = "SysInfoHelpID"
	sysinfoHelpText   string = "SysInfo Mode Text, press Esc to close."
)

// Printf - send message to the tlog interface
func (ps *PanelSystem) Printf(format string, a ...interface{}) {
	tlog.Log(sysinfoHelpID, fmt.Sprintf("%T.", ps)+format, a...)
}

func init() {
	err := vp.Register(sysinfoPanelName, SysInfoPanelIndex, SysInfoPanelSetup)
	if err != nil {
		log.Fatalf("Error registering panel: %v\n", err)
	}
}

// setupSysInfo - setup and init the sysInfo page
func setupSysInfo(cfg vp.VPanelConfig) (*PanelSystem, error) {

	ps := &PanelSystem{
		topFlex: kview.NewFlex(),
		to:      tab.New(cfg.Name, pktgen.app),
	}
	ps.topFlex.SetDirection(kview.FlexRow)

	tlog.Register(sysinfoPanelLogID)

	return ps, nil
}

// SysInfoPanelSetup setup the main cpu page
func SysInfoPanelSetup(cfg vp.VPanelConfig) (*vp.VPanelInfo, error) {
	var ps *PanelSystem

	if p, err := setupSysInfo(cfg); err != nil {
		return nil, err
	} else {
		ps = p
	}

	panel := cfg.Panel
	app := cfg.App

	flex0 := ps.topFlex
	flex1 := kview.NewFlex()
	flex1.SetDirection(kview.FlexRow)
	flex2 := kview.NewFlex()
	flex2.SetDirection(kview.FlexColumn)

	hlp.TitleBox(flex0, PktgenInfo(true))

	ps.host = hlp.CreateTextView(flex2, hlp.NewText("Host (h)", kview.AlignLeft), 0, 1, true)
	ps.mem = hlp.CreateTextView(flex2, hlp.NewText("Memory (m)", kview.AlignLeft), 0, 1, false)
	flex1.AddItem(flex2, 10, 0, true)

	ps.netInfo = hlp.CreateTableView(flex1, hlp.NewText("Host Network Information (n)", kview.AlignLeft), 0, 1, false)
	ps.netInfo.SetSelectable(false, false)
	ps.netInfo.SetFixed(1, 1)
	ps.netInfo.SetSeparator(kview.Borders.Vertical)

	ps.netStats = hlp.CreateTableView(flex1, hlp.NewText("Network Statistics (s)", kview.AlignLeft), 0, 1, false)
	ps.netStats.SetFixed(1, 1)
	ps.netStats.SetSeparator(kview.Borders.Vertical)
	ps.netStats.SetScrollBarVisibility(kview.ScrollBarAuto)
	ps.netStats.SetScrollBarColor(tcell.ColorCornflowerBlue)

	ps.netPCI = hlp.CreateTableView(flex1, hlp.NewText("PCI Network Devices (p)", kview.AlignLeft), 0, 1, false)
	ps.netStats.SetFixed(1, 0)
	ps.netPCI.SetSeparator(kview.Borders.Vertical)
	ps.netPCI.SetScrollBarVisibility(kview.ScrollBarAuto)
	ps.netPCI.SetScrollBarColor(tcell.ColorCornflowerBlue)
	ps.topFlex.AddItem(flex1, 0, 3, true)

	tabData := []tab.TabData{
		{Name: "host", View: ps.host, Key: 'h'},
		{Name: "memory", View: ps.mem, Key: 'm'},
		{Name: "hostNet", View: ps.netInfo, Key: 'n'},
		{Name: "netStats", View: ps.netStats, Key: 's'},
		{Name: "netPCI", View: ps.netPCI, Key: 'p'},
	}

	if to, err := hlp.CreateTabOrder(app, sysinfoPanelName, tabData); err != nil {
		return nil, err
	} else {
		ps.to = to
	}

	modal := kview.NewModal()
	modal.SetText(sysinfoHelpText)
	modal.AddButtons([]string{"Got it"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		panel.HidePanel(sysinfoHelpID)
	})
	AddModalPage(sysinfoHelpID, modal)

	timerFn := func(step int, ticks uint64) {
		if step == -1 || ps.topFlex.HasFocus() {
			app.QueueUpdateDraw(func() {
				ticks++
				switch step {
				case -1:
					ps.displayHost(ps.host)
					ps.displayMem(ps.mem)
					ps.displayNetInfo(ps.netInfo)
					ps.netInfo.ScrollToBeginning()
					ps.displayNetStats(ps.netStats)
					ps.displayNetPCI(ps.netPCI)

				case 0:

				case 1:

				case 2:
					ps.displayHost(ps.host)
					ps.displayMem(ps.mem)
					ps.displayNetInfo(ps.netInfo)
					ps.displayNetStats(ps.netStats)
					ps.displayNetPCI(ps.netPCI)

				case 3:
				}
			})
		}
	}

	return &vp.VPanelInfo{
		PanelName: sysinfoPanelName,
		HelpID:    sysinfoHelpID,
		TopFlex:   ps.topFlex,
		TimerFn:   timerFn,
	}, nil
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
	str += fmt.Sprintf("Virtual Role: %s, System: %s", cz.Yellow(role), cz.Yellow(vsys))

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
		ps.Printf("network IO Count: %s\n", err)
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

	titles := []hlp.TextInfo{
		hlp.NewText(cz.Yellow("State", -8), kview.AlignLeft),
		hlp.NewText(cz.Yellow("PCI Address"), kview.AlignLeft),
		hlp.NewText(cz.Yellow("Device Info"), kview.AlignLeft),
	}
	row := hlp.TableSetHeaders(view, 0, 0, titles)

	lines := pktgen.db.PCILines()
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		fields := strings.Split(line, "Ethernet controller:")

		fields[0] = strings.TrimSpace(fields[0])
		fields[1] = strings.TrimSpace(fields[1])

		rowData := []hlp.TextInfo{
			hlp.NewText("", kview.AlignLeft),
			hlp.NewText(cz.CornSilk(fields[0]), kview.AlignLeft),
			hlp.NewText(cz.SkyBlue(fields[1]), kview.AlignLeft),
		}
		for _, r := range pktgen.db.HwInfo() {
			if strings.Contains(r.BusInfo, fields[0]) {
				if r.Config.Driver == "vfio-pci" {
					rowData[0].Text = cz.GoldenRod("*Usable*")
				} else {
					rowData[0].Text = "*Active*"
				}
			}
		}
		col := 0
		for _, v := range rowData {
			col = hlp.TableCellSet(view, row, col, v)
		}
		row++
	}

	view.ScrollToBeginning()
}
