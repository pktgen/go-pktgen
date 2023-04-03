// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-23 Intel Corporation

package main

import (
	"fmt"
	"strings"

	"code.rocketnine.space/tslocum/cview"
	"github.com/gdamore/tcell/v2"
	"github.com/shirou/gopsutil/cpu"

	"github.com/pktgen/go-pktgen/internal/meter"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	tab "github.com/pktgen/go-pktgen/internal/taborder"
	tlog "github.com/pktgen/go-pktgen/internal/ttylog"
	vp "github.com/pktgen/go-pktgen/internal/vpanels"
)

// PageCPULoad - Data for main page information
type PageCPULoad struct {
	topFlex   *cview.Flex
	cpuInfo   *cview.TextView
	cpuLayout *cview.Table
	cpuInfo1  *cview.TextView
	cpuInfo2  *cview.TextView
	cpuInfo3  *cview.TextView
	tabOrder  *tab.Tab
	percent   []float64
	meter     *meter.Meter
}

const (
	cpuPanelName string = "CPU"
	labelWidth   int    = 14
)

func init() {
	tlog.Register("CPULoadLogID")
}

// Printf - send message to the ttylog interface
func (pg *PageCPULoad) Printf(format string, a ...interface{}) {
	tlog.Log("CPULoadLogID", fmt.Sprintf("%T.", pg)+format, a...)
}

// setupCPULoad - setup and init the sysInfo page
func setupCPULoad() *PageCPULoad {

	pg := &PageCPULoad{}

	return pg
}

// CPULoadPanelSetup setup
func CPULoadPanelSetup(pages *cview.Pages, nextPanel func()) (title string, content cview.Primitive) {

	pg := setupCPULoad()

	to := tab.New(cpuPanelName, pktgen.app)
	pg.tabOrder = to

	flex0 := cview.NewFlex()
	flex0.SetDirection(cview.FlexRow)
	flex1 := cview.NewFlex()
	flex1.SetDirection(cview.FlexColumn)
	flex2 := cview.NewFlex()
	flex2.SetDirection(cview.FlexColumn)

	hlp.TitleBox(flex0, PktgenInfo(true))

	pg.cpuInfo = hlp.CreateTextView(flex1, "CPU (c)", cview.AlignLeft, 0, 2, true)
	pg.cpuLayout = hlp.CreateTableView(flex1, "CPU Layout (l)", cview.AlignLeft, 0, 1, false)
	flex0.AddItem(flex1, 0, 1, true)

	pg.cpuInfo1 = hlp.CreateTextView(flex2, "CPU Load (1)", cview.AlignLeft, 0, 1, true)
	pg.cpuInfo2 = hlp.CreateTextView(flex2, "CPU Load (2)", cview.AlignLeft, 0, 1, false)
	pg.cpuInfo3 = hlp.CreateTextView(flex2, "CPU Load (3)", cview.AlignLeft, 0, 1, false)
	flex0.AddItem(flex2, 0, 4, true)

	if err := to.Add("cpuInfo", pg.cpuInfo, 'c'); err != nil {
		panic(err)
	}
	if err := to.Add("cpuLayout", pg.cpuLayout, 'l'); err != nil {
		panic(err)
	}
	if err := to.Add("cpuInfo1", pg.cpuInfo1, '1'); err != nil {
		panic(err)
	}
	if err := to.Add("cpuInfo2", pg.cpuInfo2, '2'); err != nil {
		panic(err)
	}
	if err := to.Add("cpuInfo3", pg.cpuInfo3, '3'); err != nil {
		panic(err)
	}

	if err := to.SetInputDone(); err != nil {
		panic(err)
	}

	pg.topFlex = flex0

	// Setup static pages
	pg.displayCPU(pg.cpuInfo)
	pg.displayLayout(pg.cpuLayout)

	percent, err := cpu.Percent(0, true)
	if err != nil {
		tlog.DoPrintf("Percent: %v\n", err)
	}
	pg.percent = percent

	pktgen.timers.Add(cpuPanelName, func(step int, ticks uint64) {
		if pg.topFlex.HasFocus() {
			pktgen.app.QueueUpdateDraw(func() {
				pg.displayCPULoad(step, ticks)
			})
		}
	})

	modal := cview.NewModal()
	modal.SetText("This is the Help Box: cpuInfoHelp Thank you for asking for help! Press Esc to close.")
	modal.AddButtons([]string{"Got it"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pages.HidePage("cpuInfoHelp")
	})
	AddModalPage("cpuInfoHelp", modal)

	flex0.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		k := event.Rune()
		switch k {
		case '?':
			tlog.DoPrintf("Question Mark! HasPage(%v)\n", pages.HasPage("cpuInfoHelp"))
			pages.ShowPage("cpuInfoHelp")
		}
		return event
	})

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

	return cpuPanelName, pg.topFlex
}

// Callback timer routine to display the cpuinfo panel
func (pg *PageCPULoad) displayCPULoad(step int, ticks uint64) {

	switch step {
	case 0:
		percent, err := cpu.Percent(0, true)
		if err != nil {
			tlog.DoPrintf("Percent: %v\n", err)
		}
		pg.percent = percent

	case 2:
		pg.displayLoadData(pg.cpuInfo1, 1)
		pg.displayLoadData(pg.cpuInfo2, 2)
		pg.displayLoadData(pg.cpuInfo3, 3)
	}
}

// Display the CPU information
func (pg *PageCPULoad) displayCPU(view *cview.TextView) {
	str := ""

	cd := pktgen.cpuData
	str += fmt.Sprintf("CPU   Vendor   : %s\n", cz.GoldenRod(cd.CpuInfo(0).VendorID, -14))
	str += fmt.Sprintf("      Model    : %s\n\n", cz.MediumSpringGreen(cd.CpuInfo(0).ModelName))
	str += fmt.Sprintf("Cores Logical  : %s\n", cz.Yellow(cd.NumLogicalCores(), -6))
	str += fmt.Sprintf("      Physical : %s\n", cz.Yellow(cd.NumPhysicalCores(), -6))
	str += fmt.Sprintf("      Threads  : %s\n", cz.Yellow(cd.NumHyperThreads(), -6))
	str += fmt.Sprintf("      Sockets  : %s\n", cz.Yellow(cd.NumSockets()))

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
func (pg *PageCPULoad) displayLayout(view *cview.Table) {

	cd := pktgen.cpuData

	str := cz.LightBlue(" Core", -5)
	tableCell := cview.NewTableCell(cz.YellowGreen(str))
	tableCell.SetAlign(cview.AlignLeft)
	tableCell.SetSelectable(false)
	view.SetCell(0, 0, tableCell)

	for k, s := range cd.Sockets() {
		str = cz.LightBlue(fmt.Sprintf("Socket %d", s))
		tableCell := cview.NewTableCell(cz.YellowGreen(str))
		tableCell.SetAlign(cview.AlignCenter)
		tableCell.SetSelectable(false)
		view.SetCell(0, k+1, tableCell)
	}

	row := int16(1)

	pg.Printf("numPhysical %d, numSockets %d\n", cd.NumPhysicalCores(), cd.NumSockets())
	pg.Printf("cd.Cores = %v\n", cd.Cores())
	for _, cid := range cd.Cores() {
		col := int16(0)

		tableCell := cview.NewTableCell(cz.Red(cid, 4))
		tableCell.SetAlign(cview.AlignLeft)
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
			tableCell := cview.NewTableCell(cz.YellowGreen(str))
			tableCell.SetAlign(cview.AlignLeft)
			tableCell.SetSelectable(false)
			view.SetCell(int(row), int(col+1), tableCell)
			col++
		}
		row++
	}
	view.ScrollToBeginning()
}

// Grab the percent load data and display the meters
func (pg *PageCPULoad) displayLoadData(view *cview.TextView, flg int) {

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
func (pg *PageCPULoad) displayLoad(percent []float64, start, end int16, view *cview.TextView) {

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
