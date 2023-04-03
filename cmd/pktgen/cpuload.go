// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package main

import (
	"fmt"
	"strings"

	"github.com/pktgen/go-pktgen/pkgs/kview"
	"github.com/shirou/gopsutil/cpu"

	"github.com/pktgen/go-pktgen/internal/cpudata"
	"github.com/pktgen/go-pktgen/internal/meter"
	"github.com/pktgen/go-pktgen/internal/tlog"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	tab "github.com/pktgen/go-pktgen/internal/taborder"
	vp "github.com/pktgen/go-pktgen/internal/vpanels"
)

// PanelCPULoad - Data for main page information
type PanelCPULoad struct {
	flex0     *kview.Flex
	cpuInfo   *kview.TextView
	cpuLayout *kview.Table
	cpuInfo1  *kview.TextView
	cpuInfo2  *kview.TextView
	cpuInfo3  *kview.TextView
	percent   []float64
	meter     *meter.Meter
}

const (
	cpuloadPanelName        string = "Cpuload"
	cpuloadLogID            string = "CpuloadLogID"
	cpuloadHelpID           string = "CpuloadHelpID"
	cpuloadHelpText         string = "Cpuload Mode Text, press Esc to close."
	cpuLoadCpuTabOrderID    string = "CpuloadTabOrderID"
	cpuLoadLayoutTabOrderID string = "CpuloadLayoutTabOrder"
	cpuLoadLoad1TabOrderID  string = "CpuloadLoad1TabOrder"
	cpuLoadLoad2TabOrderID  string = "CpuloadLoad2TabOrder"
	cpuLoadLoad3TabOrderID  string = "CpuloadLoad3TabOrder"
	cpuLoadCpuTabKey        rune   = 'c'
	cpuLoadLayoutTabKey     rune   = 'l'
	cpuLoadLoad1TabKey      rune   = '1'
	cpuLoadLoad2TabKey      rune   = '2'
	cpuLoadLoad3TabKey      rune   = '3'

	labelWidth int = 14
)

func CPULoadPanel() (string, vp.VPanelFunc) {
	return cpuloadPanelName, cpuLoadPanelSetup
}

// setupCPULoad - setup and init the sysInfo page
func setupPanelCPULoad(app *kview.Application) (*PanelCPULoad, error) {

	pg := &PanelCPULoad{
		flex0: kview.NewFlex(),
	}
	pg.flex0.SetDirection(kview.FlexRow)

	tlog.Register(cpuloadLogID)

	return pg, nil
}

// cpuLoadPanelSetup setup
func cpuLoadPanelSetup(cfg vp.VPanelConfig) (*vp.VPanelInfo, error) {
	var pg *PanelCPULoad

	if p, err := setupPanelCPULoad(cfg.App); err != nil {
		return nil, err
	} else {
		pg = p
	}

	hlp.TitleBox(pg.flex0, hlp.CommandInfo(true))

	pg.cpuLoadView(kview.NewFlex())
	pg.cpuLoadUsageView(kview.NewFlex())
	pg.cpuLoadMeterSetup()
	pg.cpuLoadHelpSetup(cfg)
	pg.cpuLoadTabOrderSetup(cfg)

	return &vp.VPanelInfo{
		PanelName: cpuloadPanelName,
		HelpID:    cpuloadHelpID,
		TopFlex:   pg.flex0,
		TimerFn:   pg.cpuLoadTimer(cfg),
	}, nil
}

func (pg *PanelCPULoad) cpuLoadTimer(cfg vp.VPanelConfig) func(int, uint64) {
	return func(step int, ticks uint64) {
		if step == -1 || pg.flex0.HasFocus() {
			cfg.App.QueueUpdateDraw(func() {
				ticks++
				switch step {
				case -1: // Setup static pages
					pg.displayCPU(pg.cpuInfo)
					pg.displayLayout(pg.cpuLayout)

					percent, err := cpu.Percent(0, true)
					if err != nil {
						tlog.Printf("Percent: %v\n", err)
					}
					pg.percent = percent

				case 0:
				case 1:
				case 2:
				case 3:
					pg.updatePercent()
					pg.displayLoadData(pg.cpuInfo1, 1)
					pg.displayLoadData(pg.cpuInfo2, 2)
					pg.displayLoadData(pg.cpuInfo3, 3)
				}
			})
		}
	}
}

func (pg *PanelCPULoad) cpuLoadView(f1 *kview.Flex) {

	f1.SetDirection(kview.FlexColumn)

	pg.cpuInfo = hlp.CreateTextView(f1,
		hlp.NewText(fmt.Sprintf("CPU (%c)", cpuLoadCpuTabKey), kview.AlignLeft), 0, 2, true)
	pg.cpuLayout = hlp.CreateTableView(f1,
		hlp.NewText(fmt.Sprintf("CPU Layout (%c)", cpuLoadLayoutTabKey), kview.AlignLeft), 0, 1, false)
	pg.flex0.AddItem(f1, 0, 1, true)
}

func (pg *PanelCPULoad) cpuLoadUsageView(f1 *kview.Flex) {

	f1.SetDirection(kview.FlexColumn)

	pg.cpuInfo1 = hlp.CreateTextView(f1, hlp.NewText(fmt.Sprintf("CPU Load (%c)", cpuLoadLoad1TabKey), kview.AlignLeft), 0, 1, true)
	pg.cpuInfo2 = hlp.CreateTextView(f1, hlp.NewText(fmt.Sprintf("CPU Load (%c)", cpuLoadLoad2TabKey), kview.AlignLeft), 0, 1, false)
	pg.cpuInfo3 = hlp.CreateTextView(f1, hlp.NewText(fmt.Sprintf("CPU Load (%c)", cpuLoadLoad3TabKey), kview.AlignLeft), 0, 1, false)
	pg.flex0.AddItem(f1, 0, 4, true)
}

func (pg *PanelCPULoad) cpuLoadMeterSetup() {
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
}

func (pg *PanelCPULoad) cpuLoadHelpSetup(cfg vp.VPanelConfig) {

	modal := kview.NewModal()
	modal.SetText(cpuloadHelpText)
	modal.AddButtons([]string{"Got it"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		cfg.Panels.HidePanel(cpuloadHelpID)
	})
	cfg.Panels.AddPanel(cpuloadHelpID, modal, false, false)

}

func (pg *PanelCPULoad) cpuLoadTabOrderSetup(cfg vp.VPanelConfig) error {

	tabData := []tab.TabData{
		{Name: cpuLoadCpuTabOrderID, View: pg.cpuInfo, Key: cpuLoadCpuTabKey},
		{Name: cpuLoadLayoutTabOrderID, View: pg.cpuLayout, Key: cpuLoadLayoutTabKey},
		{Name: cpuLoadLoad1TabOrderID, View: pg.cpuInfo1, Key: cpuLoadLoad1TabKey},
		{Name: cpuLoadLoad2TabOrderID, View: pg.cpuInfo2, Key: cpuLoadLoad2TabKey},
		{Name: cpuLoadLoad3TabOrderID, View: pg.cpuInfo3, Key: cpuLoadLoad3TabKey},
	}

	if _, err := hlp.CreateTabOrder(cfg.App, cpuloadPanelName, tabData); err != nil {
		return err
	}

	return nil
}

func (pg *PanelCPULoad) updatePercent() {
	percent, err := cpu.Percent(0, true)
	if err != nil {
		tlog.Printf("Percent: %v\n", err)
	}
	pg.percent = percent
}

// Display the CPU information
func (pg *PanelCPULoad) displayCPU(view *kview.TextView) {
	str := ""

	cd := cpudata.GetCPUData()
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

	cd := cpudata.GetCPUData()

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

	for _, cid := range cd.Cores() {
		col := int16(0)

		tableCell := kview.NewTableCell(cz.Red(cid, 4))
		tableCell.SetAlign(kview.AlignLeft)
		tableCell.SetSelectable(false)
		view.SetCell(int(row), int(col), tableCell)

		for sid := int16(0); sid < cd.NumSockets(); sid++ {
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

	cd := cpudata.GetCPUData()
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
