// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-23 Intel Corporation

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/pktgen/go-pktgen/pkgs/kview"
	"github.com/shirou/gopsutil/cpu"

	"github.com/pktgen/go-pktgen/internal/meter"
	"github.com/pktgen/go-pktgen/internal/tlog"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	tab "github.com/pktgen/go-pktgen/internal/taborder"
	vp "github.com/pktgen/go-pktgen/internal/vpanels"
)

// PanelCPULoad - Data for main page information
type PanelCPULoad struct {
	topFlex   *kview.Flex
	cpuInfo   *kview.TextView
	cpuLayout *kview.Table
	cpuInfo1  *kview.TextView
	cpuInfo2  *kview.TextView
	cpuInfo3  *kview.TextView
	to        *tab.Tab
	percent   []float64
	meter     *meter.Meter
}

const (
	cpuloadPanelName string = "Cpuload"
	cpuloadLogID     string = "CpuloadLogID"
	cpuloadHelpID    string = "CpuloadHelpID"
	cpuloadHelpText  string = "Cpuload Mode Text, press Esc to close."

	labelWidth int = 14
)

// Printf - send message to the tlog interface
func (pg *PanelCPULoad) Printf(format string, a ...interface{}) {
	tlog.Log(cpuloadHelpID, fmt.Sprintf("%T.", pg)+format, a...)
}

func init() {
	err := vp.Register(cpuloadPanelName, CPUPanelIndex, CPULoadPanelSetup)
	if err != nil {
		log.Fatalf("Error registering panel: %v\n", err)
	}
}

// setupCPULoad - setup and init the sysInfo page
func setupPanelCPULoad() (*PanelCPULoad, error) {

	pg := &PanelCPULoad{
		topFlex: kview.NewFlex(),
		to:      tab.New(cpuloadPanelName, pktgen.app),
	}
	pg.topFlex.SetDirection(kview.FlexRow)

	tlog.Register(cpuloadLogID)

	return pg, nil
}

// CPULoadPanelSetup setup
func CPULoadPanelSetup(cfg vp.VPanelConfig) (*vp.VPanelInfo, error) {
	var pg *PanelCPULoad

	if p, err := setupPanelCPULoad(); err != nil {
		return nil, err
	} else {
		pg = p
	}

	panel := cfg.Panel
	app := cfg.App

	flex1 := kview.NewFlex()
	flex1.SetDirection(kview.FlexColumn)
	flex2 := kview.NewFlex()
	flex2.SetDirection(kview.FlexColumn)

	hlp.TitleBox(pg.topFlex, PktgenInfo(true))

	pg.cpuInfo = hlp.CreateTextView(flex1, hlp.NewText("CPU (c)", kview.AlignLeft), 0, 2, true)
	pg.cpuLayout = hlp.CreateTableView(flex1, hlp.NewText("CPU Layout (l)", kview.AlignLeft), 0, 1, false)
	pg.topFlex.AddItem(flex1, 0, 1, true)

	pg.cpuInfo1 = hlp.CreateTextView(flex2, hlp.NewText("CPU Load (1)", kview.AlignLeft), 0, 1, true)
	pg.cpuInfo2 = hlp.CreateTextView(flex2, hlp.NewText("CPU Load (2)", kview.AlignLeft), 0, 1, false)
	pg.cpuInfo3 = hlp.CreateTextView(flex2, hlp.NewText("CPU Load (3)", kview.AlignLeft), 0, 1, false)
	pg.topFlex.AddItem(flex2, 0, 4, true)

	tabData := []tab.TabData{
		{Name: "cpuInfo", View: pg.cpuInfo, Key: 'c'},
		{Name: "cpuLayout", View: pg.cpuLayout, Key: 'l'},
		{Name: "cpuLoad1", View: pg.cpuInfo1, Key: '1'},
		{Name: "cpuload2", View: pg.cpuInfo2, Key: '2'},
		{Name: "cpuload3", View: pg.cpuInfo3, Key: '3'},
	}

	if to, err := hlp.CreateTabOrder(app, cpuloadPanelName, tabData); err != nil {
		return nil, err
	} else {
		pg.to = to
	}

	// Setup static pages
	pg.displayCPU(pg.cpuInfo)
	pg.displayLayout(pg.cpuLayout)

	percent, err := cpu.Percent(0, true)
	if err != nil {
		tlog.DoPrintf("Percent: %v\n", err)
	}
	pg.percent = percent

	modal := kview.NewModal()
	modal.SetText(cpuloadHelpText)
	modal.AddButtons([]string{"Got it"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		panel.HidePanel(cpuloadHelpID)
	})
	AddModalPage(cpuloadHelpID, modal)

	pg.meter = meter.New().
		SetWidth(func() int {
			_, _, width, _ := pg.cpuInfo1.GetInnerRect()

			return width
		}).
		SetDraw(func(mi *meter.Info) string {
			var str string = ""

			for _, l := range mi.Labels {

				if l.Fn == nil {
					l.Fn = cz.Default
				}
				str += l.Fn(l.Val)
			}
			str += fmt.Sprintf("[%s]\n", mi.Bar.Fn(mi.Bar.Val))
			return str
		}).
		SetRateLimits(0.0, 100.0)

	return &vp.VPanelInfo{
		PanelName: cpuloadPanelName,
		HelpID:    cpuloadHelpID,
		TopFlex:   pg.topFlex,
		TimerFn: func(step int, ticks uint64) {
			if step == -1 || pg.topFlex.HasFocus() {
				app.QueueUpdateDraw(func() {
					ticks++
					switch step {
					case -1:
						pg.updatePercent()
						pg.displayLoadData(pg.cpuInfo1, 1)
						pg.displayLoadData(pg.cpuInfo2, 2)
						pg.displayLoadData(pg.cpuInfo3, 3)

					case 0:
						pg.updatePercent()

					case 2:
						pg.displayLoadData(pg.cpuInfo1, 1)
						pg.displayLoadData(pg.cpuInfo2, 2)
						pg.displayLoadData(pg.cpuInfo3, 3)

					case 3:
					}
				})
			}
		},
	}, nil
}

func (pg *PanelCPULoad) updatePercent() {
	percent, err := cpu.Percent(0, true)
	if err != nil {
		tlog.DoPrintf("Percent: %v\n", err)
	}
	pg.percent = percent
}

// Display the CPU information
func (pg *PanelCPULoad) displayCPU(view *kview.TextView) {
	str := ""

	cd := pktgen.cpuData
	str += fmt.Sprintf("CPU   Vendor   : %s\n", cz.GoldenRod(cd.CpuInfo(0).VendorID, -14))
	str += fmt.Sprintf("      Model    : %s\n\n", cz.MediumSpringGreen(cd.CpuInfo(0).ModelName))
	str += fmt.Sprintf("Cores Logical  : %s per socket\n", cz.Yellow(cd.NumLogicalCores()/cd.NumSockets(), -6))
	str += fmt.Sprintf("      Physical : %s per socket\n", cz.Yellow(cd.NumPhysicalCores(), -6))
	str += fmt.Sprintf("      Threads  : %s per physical core\n", cz.Yellow(cd.NumHyperThreads(), -6))
	str += fmt.Sprintf("      Sockets  : %s\n", cz.Yellow(cd.NumSockets()))
	str += fmt.Sprintf(" Total logical : %s", cz.Yellow(cd.NumLogicalCores(), -6))

	view.SetText(str)
	view.ScrollToBeginning()
}

// Build up a string for displaying the CPU layout window
func buildStr(a []uint16, width int) string {

	str := "{"

	for k, v := range a {
		str += cz.Green(v, width)
		if k < (len(a) - 1) {
			str += " /"
		}
	}

	return str + " }"
}

// Display the CPU layout data
func (pg *PanelCPULoad) displayLayout(view *kview.Table) {

	cd := pktgen.cpuData

	str := cz.LightBlue(" Core", -5)
	tableCell := kview.NewTableCell(cz.YellowGreen(str))
	tableCell.SetAlign(kview.AlignLeft)
	tableCell.SetSelectable(false)
	view.SetCell(0, 0, tableCell)

	for k, s := range cd.Sockets() {
		str = cz.LightBlue(fmt.Sprintf("Socket %d", s))
		tableCell := kview.NewTableCell(cz.YellowGreen(str))
		tableCell.SetAlign(kview.AlignCenter)
		tableCell.SetSelectable(false)
		view.SetCell(0, k+1, tableCell)
	}

	row := int16(1)

	pg.Printf("numPhysical %d, numSockets %d\n", cd.NumPhysicalCores(), cd.NumSockets())
	pg.Printf("cd.Cores = %v\n", cd.Cores())
	for _, cid := range cd.Cores() {
		col := int16(0)

		tableCell := kview.NewTableCell(cz.Red(cid, 4))
		tableCell.SetAlign(kview.AlignLeft)
		tableCell.SetSelectable(false)
		view.SetCell(int(row), int(col), tableCell)

		pg.Printf("cid %d\n", cid)
		for sid := int16(0); sid < cd.NumSockets(); sid++ {
			pg.Printf("  sid %d\n", sid)
			key := uint16(sid<<uint16(8)) | cid
			v, ok := cd.CoreMapItem(key)
			if ok {
				str = fmt.Sprintf(" %s", buildStr(v, 3))
			} else {
				str = fmt.Sprintf(" %s", strings.Repeat(".", 10))
			}
			tableCell := kview.NewTableCell(cz.YellowGreen(str))
			tableCell.SetAlign(kview.AlignLeft)
			tableCell.SetSelectable(false)
			view.SetCell(int(row), int(col+1), tableCell)
			col++
		}
		row++
	}
	view.ScrollToBeginning()
}

// Grab the percent load data and display the meters
func (pg *PanelCPULoad) displayLoadData(view *kview.TextView, flg int) {

	cd := pktgen.cpuData
	num := int16(cd.NumLogicalCores()/3) + 1

	switch flg {
	case 1:
		pg.displayLoad(pg.percent, 0, num, view)
	case 2:
		pg.displayLoad(pg.percent, num, num*int16(2), view)
	case 3:
		pg.displayLoad(pg.percent, num*int16(2), cd.NumLogicalCores(), view)
	}
}

// Display the load meters
func (pg *PanelCPULoad) displayLoad(percent []float64, start, end int16, view *kview.TextView) {

	_, _, width, _ := view.GetInnerRect()

	width -= labelWidth
	if width <= 0 {
		return
	}
	str := ""

	str += fmt.Sprintf("%s\n", cz.Orange("Core Percent          Load Meter"))

	for i := start; i < end; i++ {
		str += pg.meter.Draw(percent[i], &meter.Info{
			Labels: []*meter.LabelInfo{
				{Val: fmt.Sprintf("%3d", i), Fn: nil},
				{Val: ":", Fn: nil},
				{Val: fmt.Sprintf("%5.1f", percent[i]), Fn: cz.Red},
				{Val: "%", Fn: nil},
			},
			Bar: &meter.LabelInfo{Val: "", Fn: cz.MediumSpringGreen},
		})
	}
	str = str[:len(str)-1] // Strip the last newline character

	view.SetText(str)
	view.ScrollToBeginning()
}
