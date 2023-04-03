// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pktgen/go-pktgen/pkgs/kview"

	"github.com/pktgen/go-pktgen/internal/configview"
	"github.com/pktgen/go-pktgen/internal/meter"
	"github.com/pktgen/go-pktgen/internal/tlog"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	tab "github.com/pktgen/go-pktgen/internal/taborder"
	vp "github.com/pktgen/go-pktgen/internal/vpanels"
)

// PanelSingleMode - Data for main panel information
type PanelSingleMode struct {
	flex0       *kview.Flex
	configView  *configview.ConfigView
	statsView   *StatsView
	perfView    *PerfView
	currentPort uint16
	to          *tab.Tab
	meter       *meter.Meter
}

const (
	singlePanelName        string = "Single"
	singleLogID            string = "SingleLogID"
	singleHelpID           string = "SingleHelpID"
	singleHelpText         string = "Single Mode Text, press Esc to close."
	singleConfigTabOrderID string = "SingleConfigTabOrderID"
	singleStatsTabOrderID  string = "SingleStatsTabOrderID"
	singlePerfTabOrderID   string = "SinglePerfTabOrderID"
	singleConfigTabKey     rune   = 'c'
	singleStatsTabKey      rune   = 'S'
	singlePerfTabKey       rune   = 'p'
)

func SingleModePanel() (string, vp.VPanelFunc) {
	return singlePanelName, singleModePanelSetup
}

// SingleModePanelSetup setup
func singleModePanelSetup(cfg vp.VPanelConfig) (*vp.VPanelInfo, error) {

	tlog.Register(singleLogID)

	ps := &PanelSingleMode{
		flex0:       kview.NewFlex(),
		to:          tab.New(singlePanelName, cfg.App),
		currentPort: 0,
	}
	ps.flex0.SetDirection(kview.FlexRow)

	// Setup and add the title box at the top of the screen
	hlp.TitleBox(ps.flex0, hlp.CommandInfo(true))

	// Setup views and set up tab order for each view
	ps.singleConfigView(cfg)       // Config view
	ps.singlePerfView()            // Perf view
	ps.singleStatsView()           // Stats view
	ps.singleHelpSetup(cfg)        // Help view
	ps.singleTabOrderSetup(cfg)    // Tab order setup
	ps.singleConfigKeyCapture(cfg) // Config key press capture

	return &vp.VPanelInfo{
		PanelName: singlePanelName,
		HelpID:    singleHelpID,
		TopFlex:   ps.flex0,
		TimerFn:   ps.singleTimer(cfg),
	}, nil
}

func (ps *PanelSingleMode) singleTimer(cfg vp.VPanelConfig) func(int, uint64) {
	return func(step int, ticks uint64) {
		if step == -1 || ps.flex0.HasFocus() {
			cfg.App.QueueUpdateDraw(func() {
				cv := ps.configView
				pv := ps.perfView
				sv := ps.statsView

				ticks++
				switch step {
				case -1:
					pktgenApp.gPkt.UpdateStats()
					cv.DisplayConfigTable()
					pv.DisplayPerf(ps.meter, pktgenApp.gPkt.GetRxPercentSlice(), pktgenApp.gPkt.GetTxPercentSlice())
					sv.DisplayStats()

				case 0:
					pktgenApp.gPkt.UpdateStats()

				case 1:

				case 2:

				case 3:
					cv.DisplayConfigTable()
					pv.DisplayPerf(ps.meter, pktgenApp.gPkt.GetRxPercentSlice(), pktgenApp.gPkt.GetTxPercentSlice())
					sv.DisplayStats()
				}
			})
		}
	}
}

func (ps *PanelSingleMode) singleConfigView(cfg vp.VPanelConfig) {

	flex := kview.NewFlex()
	flex.SetDirection(kview.FlexRow)

	ps.configView = configview.Create(cfg.Panels, ps.to, pktgenApp.gPkt.PortCount(), singleConfigTabKey, flex)
	ps.flex0.AddItem(flex, int(pktgenApp.gPkt.PortCount()+3), 0, true)

	// Add modal pages for each port in the config view
	for port := uint16(0); port < pktgenApp.gPkt.PortCount(); port++ {
		s := ps.configView.FormName(port)
		f := ps.configView.ConfigForm(port)
		cfg.Panels.AddPanel(s, f, false, false)
	}
}

func (ps *PanelSingleMode) singlePerfView() {

	ps.singleMeterView()           // Meter view

	flex := kview.NewFlex()
	flex.SetDirection(kview.FlexRow)

	ps.perfView = NewPerfView(pktgenApp.gPkt.PortCount())
	ps.perfView.CreatePerfView(flex, singlePerfTabKey)
	ps.flex0.AddItem(flex, 0, 1, true)
}

func (ps *PanelSingleMode) singleStatsView() {

	flex := kview.NewFlex()
	flex.SetDirection(kview.FlexRow)

	ps.statsView = CreateStatsView(pktgenApp.gPkt.PortCount(), flex, singleStatsTabKey)
	ps.flex0.AddItem(flex, 0, 3, true)
}

func (ps *PanelSingleMode) singleMeterView() {

	ps.meter = meter.New()
	ps.meter.SetWidth(func() int {
		_, _, width, _ := ps.perfView.TextView().GetInnerRect()

		return width
	})
	ps.meter.SetDraw(func(mi *meter.Info) string {
		var str strings.Builder

		for _, l := range mi.Labels {

			if l.Fn == nil {
				l.Fn = cz.Default
			}
			str.WriteString(l.Fn(l.Val))
		}
		str.WriteString(fmt.Sprintf("[%s]\n", mi.Bar.Fn(mi.Bar.Val)))
		return str.String()
	})
	ps.meter.SetRateLimits(0.0, 100.0)
}

func (ps *PanelSingleMode) singleHelpSetup(cfg vp.VPanelConfig) {
	modal := kview.NewModal()
	modal.SetText(singleHelpText)
	modal.AddButtons([]string{"Got it"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		cfg.Panels.HidePanel(singleHelpID)
	})
	cfg.Panels.AddPanel(singleHelpID, modal, false, false)
}

func (ps *PanelSingleMode) singleTabOrderSetup(cfg vp.VPanelConfig) error {
	tabData := []tab.TabData{
		{Name: singleConfigTabOrderID, View: ps.configView.TableView(), Key: singleConfigTabKey},
		{Name: singlePerfTabOrderID, View: ps.perfView.TextView(), Key: singlePerfTabKey},
		{Name: singleStatsTabOrderID, View: ps.statsView.TableView(), Key: singleStatsTabKey},
	}

	if to, err := hlp.CreateTabOrder(cfg.App, singlePanelName, tabData); err != nil {
		return err
	} else {
		ps.to = to
	}

	return nil
}

func (ps *PanelSingleMode) singleConfigKeyCapture(cfg vp.VPanelConfig) {

	captureInput := func(event *tcell.EventKey) *tcell.EventKey {
		cv := ps.configView

		pid, _ := ps.configView.TableView().GetSelection()
		ps.currentPort = uint16(pid)
		port := ps.currentPort - 1

		if event.Rune() == 'e' { // Edit the packet configuration
			title := hlp.TitleColor(fmt.Sprintf("Edit Port %d", port))
			cv.ConfigForm(ps.currentPort - 1).SetTitle(title)

			name := ps.configView.FormName(port)
			cfg.Panels.ShowPanel(name)
			cfg.Panels.SendToFront(name)
		} else if event.Rune() == 's' { // Start/Stop a single port transmitting
			sc := cv.PacketConfigByPort(port)
			if sc.TxState {
				sc.TxState = false
				//ps.statsView.TxPercentSet(port, 0) // Set rate to zero
			} else {
				sc.TxState = true
			}
		} else if event.Rune() == 'a' { // Start all ports transmitting
			for i := uint16(0); i < pktgenApp.gPkt.PortCount(); i++ {
				cv.SetTxState(i, true)
			}
		} else if event.Rune() == 'A' { // Stop all ports transmitting
			for i := uint16(0); i < pktgenApp.gPkt.PortCount(); i++ {
				cv.SetTxState(i, false)
				//ps.statsView.TxPercentSet(i, 0) // Set rate to zero
			}
		} else if event.Key() == tcell.KeyDown { // Select the next port
			ps.currentPort++
			if ps.currentPort > pktgenApp.gPkt.PortCount() {
				ps.currentPort = 0
			}
			cv.TableView().Select(int(ps.currentPort), 0)
		} else if event.Key() == tcell.KeyUp { // Select the previous port
			ps.currentPort--
			if ps.currentPort < 1 {
				ps.currentPort = pktgenApp.gPkt.PortCount()
			}
			cv.TableView().Select(int(ps.currentPort), 0)
		} else {
			return event
		}

		return nil
	}

	ps.configView.TableView().SetInputCapture(captureInput)
}
