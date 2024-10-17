// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2023-2024 Intel Corporation

package main

import (
	//"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/pktgen/go-pktgen/pkgs/kview"

	//cz "github.com/pktgen/go-pktgen/internal/colorize"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	tab "github.com/pktgen/go-pktgen/internal/taborder"
	"github.com/pktgen/go-pktgen/internal/tlog"
	vp "github.com/pktgen/go-pktgen/internal/vpanels"
)

// PanelMapping - Data for main panel information
type PanelMapping struct {
	topFlex *kview.Flex
	mapView *kview.Table
	to      *tab.Tab
}

const (
	mappingPanelName  string = "Mapping"
	mappingLogID      string = "MappingLogID"
	mappingHelpID     string = "MappingHelpID"
	mappingHelpText   string = "Mapping Help, press Esc to close."
	mappingTabOrderID string = "MappingTabOrderID"
)

func init() {
	err := vp.Register(mappingPanelName, MappingPanelIndex, MappingPanelSetup)
	if err != nil {
		log.Fatalf("Error registering panel: %v\n", err)
	}
}

// MappingPanelSetup setup
func MappingPanelSetup(cfg vp.VPanelConfig) (*vp.VPanelInfo, error) {

	tlog.Register(mappingLogID)

	pm := &PanelMapping{
		topFlex: kview.NewFlex(),
		to:      tab.New(mappingPanelName, pktgen.app),
	}
	pm.topFlex.SetDirection(kview.FlexRow)

	// Setup and add the title box at the top of the screen
	hlp.TitleBox(pm.topFlex, PktgenInfo(true))

	// Setup views and set up tab order for each view
	pm.setupMappingView() // Mapping view

	pm.setupHelpModal() // Help modal view

	pm.setupTabOrder()           // Needs to be called after all views are added
	pm.setupMappingInputCapture() // Needs to be called after setupTabOrder()

	timerFn := func(step int, ticks uint64) {
		if step == -1 || pm.topFlex.HasFocus() {
			cfg.App.QueueUpdateDraw(func() {
				ticks++
				switch step {
				case -1:

				case 0:

				case 2:

				case 3:
				}
			})
		}
	}

	return &vp.VPanelInfo{
		PanelName: mappingPanelName,
		HelpID:    mappingHelpID,
		TopFlex:   pm.topFlex,
		TimerFn:   timerFn,
	}, nil
}

func (pm *PanelMapping) setupMappingView() {

	flex := kview.NewFlex()
	flex.SetDirection(kview.FlexRow)

	pm.mapView = hlp.CreateTableView(flex, hlp.NewText("Mapping (m)", kview.AlignLeft), 0, 1, true)
	pm.topFlex.AddItem(flex, 0, 1, false)

}

func (pm *PanelMapping) setupMappingInputCapture() {

	captureInput := func(event *tcell.EventKey) *tcell.EventKey {

		return event
	}

	pm.mapView.SetInputCapture(captureInput)
}

func (pm *PanelMapping) setupTabOrder() {

	if err := pm.to.Add(mappingTabOrderID, pm.mapView, 'm'); err != nil {
		panic(err)
	}
	// Tell the tab order we are done setting up the order.
	if err := pm.to.SetInputDone(); err != nil {
		panic(err)
	}
}

func (pm *PanelMapping) setupHelpModal() {

	modal := kview.NewModal()
	modal.SetText(mappingHelpText)
	modal.AddButtons([]string{"Got it"})

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pktgen.panels.HidePanel(mappingHelpID)
	})

	AddModalPage(mappingHelpID, modal)
}
