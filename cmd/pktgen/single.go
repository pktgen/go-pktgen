// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/gdamore/tcell/v2"
	"github.com/pktgen/go-pktgen/pkgs/kview"

	"github.com/pktgen/go-pktgen/internal/configview"
	"github.com/pktgen/go-pktgen/internal/meter"
	"github.com/pktgen/go-pktgen/internal/perfview"
	"github.com/pktgen/go-pktgen/internal/portinfo"
	"github.com/pktgen/go-pktgen/internal/statsview"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	tab "github.com/pktgen/go-pktgen/internal/taborder"
	"github.com/pktgen/go-pktgen/internal/tlog"
	vp "github.com/pktgen/go-pktgen/internal/vpanels"
)

// PanelSingleMode - Data for main panel information
type PanelSingleMode struct {
	topFlex     *kview.Flex
	configView  *configview.ConfigView
	statsView   *statsview.StatsView
	perfView    *perfview.PerfView
	portInfo    *portinfo.PortInfo
	currentPort int
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
)

func init() {
	err := vp.Register(singlePanelName, SinglePanelIndex, SingleModePanelSetup)
	if err != nil {
		log.Fatalf("Error registering panel: %v\n", err)
	}
}

// SingleModePanelSetup setup
func SingleModePanelSetup(cfg vp.VPanelConfig) (*vp.VPanelInfo, error) {

	tlog.Register(singleLogID)

	ps := &PanelSingleMode{
		topFlex: kview.NewFlex(),
		to:      tab.New(singlePanelName, pktgen.app),
	}
	ps.topFlex.SetDirection(kview.FlexRow)

	// Setup and add the title box at the top of the screen
	hlp.TitleBox(ps.topFlex, PktgenInfo(true))

	// Setup views and set up tab order for each view
	ps.setupConfigView() // Config view
	ps.setupPerfView()   // Perf view
	ps.setupStatsView()  // Stats view

	ps.setupHelpModal() // Help modal view
	ps.setupMeter()     // Meter view

	ps.setupTabOrder()           // Needs to be called after all views are added
	ps.setupConfigInputCapture() // Needs to be called after setupTabOrder()

	ps.portInfo = portinfo.New(pktgen.portCnt)

	timerFn := func(step int, ticks uint64) {
		if step == -1 || ps.topFlex.HasFocus() {
			cfg.App.QueueUpdateDraw(func() {
				sv := ps.statsView
				cv := ps.configView
				pv := ps.perfView

				ticks++
				switch step {
				case -1:
					ps.pullStats()
					cv.DisplayConfigTable()
					sv.DisplayStats()
					pv.DisplayPerf(ps.meter, sv.RxPercentArray(), sv.TxPercentArray())

				case 0:
					ps.pullStats()

				case 2:
					cv.DisplayConfigTable()
					sv.DisplayStats()
					pv.DisplayPerf(ps.meter, sv.RxPercentArray(), sv.TxPercentArray())

				case 3:
				}
			})
		}
	}

	return &vp.VPanelInfo{
		PanelName: singlePanelName,
		HelpID:    singleHelpID,
		TopFlex:   ps.topFlex,
		TimerFn:   timerFn,
	}, nil
}

func (ps *PanelSingleMode) pullStats() {

	sv := ps.statsView
	cv := ps.configView
	for port := 0; port < pktgen.portCnt; port++ {
		sv.RxPercentSet(port, float64(rand.Intn(100)))
		if cv.TxState(port) {
			sv.TxPercentSet(port, float64(rand.Intn(100)))
		}
	}
}

func (ps *PanelSingleMode) setupConfigView() {

	flex := kview.NewFlex()
	flex.SetDirection(kview.FlexRow)

	ps.configView = configview.Create(pktgen.panels, ps.to, pktgen.portCnt, "c", flex)
	ps.topFlex.AddItem(flex, pktgen.portCnt+3, 0, true)

	for port := 0; port < pktgen.portCnt; port++ {
		s := ps.configView.FormName(port)
		f := ps.configView.ConfigForm(port)
		AddModalPage(s, f)
	}
}

func (ps *PanelSingleMode) setupConfigInputCapture() {

	captureInput := func(event *tcell.EventKey) *tcell.EventKey {
		cv := ps.configView

		ps.currentPort, _ = ps.configView.TableView().GetSelection()

		if event.Rune() == 'e' { // Edit the packet configuration
			title := hlp.TitleColor(fmt.Sprintf("Edit Port %d", ps.currentPort-1))
			cv.ConfigForm(ps.currentPort - 1).SetTitle(title)

			pktgen.panels.ShowPanel(ps.configView.FormName(ps.currentPort - 1))
		} else if event.Rune() == 's' { // Start/Stop a single port transmitting
			sc := cv.PacketConfigByPort(ps.currentPort - 1)
			if sc.TxState {
				sc.TxState = false
				ps.statsView.TxPercentSet(ps.currentPort-1, 0) // Set rate to zero
			} else {
				sc.TxState = true
			}
		} else if event.Rune() == 'a' { // Start all ports transmitting
			for i := 0; i < pktgen.portCnt; i++ {
				cv.SetTxState(i, true)
			}
		} else if event.Rune() == 'A' { // Stop all ports transmitting
			for i := 0; i < pktgen.portCnt; i++ {
				cv.SetTxState(i, false)
				ps.statsView.TxPercentSet(i, 0) // Set rate to zero
			}
		} else if event.Key() == tcell.KeyDown { // Select the next port
			ps.currentPort++
			if ps.currentPort > pktgen.portCnt {
				ps.currentPort = 0
			}
			cv.TableView().Select(ps.currentPort, 0)
		} else if event.Key() == tcell.KeyUp { // Select the previous port
			ps.currentPort--
			if ps.currentPort < 1 {
				ps.currentPort = pktgen.portCnt
			}
			cv.TableView().Select(ps.currentPort, 0)
		} else {
			return event
		}

		return nil
	}

	ps.configView.TableView().SetInputCapture(captureInput)
}

func (ps *PanelSingleMode) setupPerfView() {

	flex := kview.NewFlex()
	flex.SetDirection(kview.FlexRow)

	ps.perfView = perfview.Create(pktgen.portCnt, flex, "p")
	ps.topFlex.AddItem(flex, 0, 2, true)
}

func (ps *PanelSingleMode) setupStatsView() {

	flex := kview.NewFlex()
	flex.SetDirection(kview.FlexRow)

	ps.statsView = statsview.Create(pktgen.portCnt, flex, "S")
	ps.topFlex.AddItem(flex, 23, 1, true)
}

func (ps *PanelSingleMode) setupTabOrder() {

	if err := ps.to.Add(singleConfigTabOrderID, ps.configView.TableView(), 'c'); err != nil {
		panic(err)
	}
	if err := ps.to.Add(singlePerfTabOrderID, ps.perfView.TextView(), 'p'); err != nil {
		panic(err)
	}
	if err := ps.to.Add(singleStatsTabOrderID, ps.statsView.TableView(), 'S'); err != nil {
		panic(err)
	}
	// Tell the tab order we are done setting up the order.
	if err := ps.to.SetInputDone(); err != nil {
		panic(err)
	}
}

func (ps *PanelSingleMode) setupHelpModal() {

	modal := kview.NewModal()
	modal.SetText(singleHelpText)
	modal.AddButtons([]string{"Got it"})

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pktgen.panels.HidePanel(singleHelpID)
	})

	AddModalPage(singleHelpID, modal)
}

func (ps *PanelSingleMode) setupMeter() {

	ps.meter = meter.New()
	ps.meter.SetWidth(func() int {
		_, _, width, _ := ps.perfView.TextView().GetInnerRect()

		return width
	})
	ps.meter.SetDraw(func(mi *meter.Info) string {
		var str string = ""

		for _, l := range mi.Labels {

			if l.Fn == nil {
				l.Fn = cz.Default
			}
			str += l.Fn(l.Val)
		}
		str += fmt.Sprintf("[%s]\n", mi.Bar.Fn(mi.Bar.Val))
		return str
	})
	ps.meter.SetRateLimits(0.0, 100.0)
}
