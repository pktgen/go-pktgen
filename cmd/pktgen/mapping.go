// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package main

import (
	"fmt"

	gpc "github.com/pktgen/go-pktgen/internal/gpcommon"
	"github.com/pktgen/go-pktgen/internal/tlog"
	"github.com/pktgen/go-pktgen/pkgs/kview"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	tab "github.com/pktgen/go-pktgen/internal/taborder"
	vp "github.com/pktgen/go-pktgen/internal/vpanels"
)

// PanelMapping - Data for main panel information
type PanelMapping struct {
	flex0    *kview.Flex
	mapView  *kview.Table
	jsonView *kview.TextView
}

const (
	mappingPanelName      string = "Mapping"
	mappingLogID          string = "MappingLogID"
	mappingHelpID         string = "MappingHelpID"
	mappingHelpText       string = "Mapping Help, press Esc to close."
	mappingViewTabOrderID string = "MappingViewTabOrderID"
	mappingJsonTabOrderID string = "MappingJsonTabOrderID"
	mappingTabKey         rune   = 'm'
	mappingJSONTabKey     rune   = 'J'
)

func MappingPanel() (string, vp.VPanelFunc) {
	return mappingPanelName, mappingPanelSetup
}

// MappingPanelSetup setup
func mappingPanelSetup(cfg vp.VPanelConfig) (*vp.VPanelInfo, error) {

	tlog.Register(mappingLogID)

	pm := &PanelMapping{
		flex0: kview.NewFlex(),
	}
	pm.flex0.SetDirection(kview.FlexRow)

	// Setup and add the title box at the top of the screen
	hlp.TitleBox(pm.flex0, hlp.CommandInfo(true))

	flex := kview.NewFlex()
	flex.SetDirection(kview.FlexColumn)
	pm.flex0.AddItem(flex, 0, 1, true)

	// Setup views and set up tab order for each view
	pm.mappingView(flex)         // Mapping view
	pm.mappingJSONView(flex)     // Mapping view
	pm.mappingHelpSetup(cfg)     // Mapping help view
	pm.mappingTabOrderSetup(cfg) // Mapping tab order

	return &vp.VPanelInfo{
		PanelName: mappingPanelName,
		HelpID:    mappingHelpID,
		TopFlex:   pm.flex0,
		TimerFn:   pm.mappingTimer(cfg),
	}, nil
}

func (pm *PanelMapping) mappingTimer(cfg vp.VPanelConfig) func(int, uint64) {
	return func(step int, ticks uint64) {
		if step == -1 || pm.flex0.HasFocus() {
			cfg.App.QueueUpdateDraw(func() {
				ticks++
				switch step {
				case -1:
					pm.displayMapping(pm.mapView)
					pm.displayJSON(pm.jsonView)

				case 0:

				case 1:

				case 2:

				case 3:
				}
			})
		}
	}
}

func (pm *PanelMapping) mappingView(f1 *kview.Flex) {

	pm.mapView = hlp.CreateTableView(f1,
		hlp.NewText(fmt.Sprintf("Mapping (%c)", mappingTabKey), kview.AlignLeft), 0, 1, true)
	pm.mapView.SetFixed(1, 1)
	pm.mapView.SetBorders(true)
}

func (pm *PanelMapping) mappingJSONView(f1 *kview.Flex) {

	pm.jsonView = hlp.CreateTextView(f1,
		hlp.NewText(fmt.Sprintf("JSON (%c)", mappingJSONTabKey), kview.AlignLeft), 0, 1, false)
}

func (ps *PanelMapping) mappingHelpSetup(cfg vp.VPanelConfig) {

	modal := kview.NewModal()
	modal.SetText(mappingHelpText)
	modal.AddButtons([]string{"Got it"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		cfg.Panels.HidePanel(mappingHelpID)
	})
	cfg.Panels.AddPanel(mappingHelpID, modal, false, false)
}

func (pm *PanelMapping) mappingTabOrderSetup(cfg vp.VPanelConfig) error {

	tabData := []tab.TabData{
		{Name: mappingViewTabOrderID, View: pm.mapView, Key: mappingTabKey},
		{Name: mappingJsonTabOrderID, View: pm.jsonView, Key: mappingJSONTabKey},
	}

	if _, err := hlp.CreateTabOrder(cfg.App, mappingPanelName, tabData); err != nil {
		return err
	}
	return nil
}

func (pm *PanelMapping) displayMapping(view *kview.Table) {

	coreList := pktgenApp.gPkt.CoreList()

	pm.mapView.Clear()
	titles := []hlp.TextInfo{}
	titles = append(titles, hlp.NewText(cz.Yellow("Core"), kview.AlignCenter))
	for i := uint16(0); i < pktgenApp.gPkt.PortCount(); i++ {
		s := fmt.Sprintf("Port:%2d", i)
		titles = append(titles, hlp.NewText(cz.Yellow(s), kview.AlignCenter))
	}
	row := 0
	col := 0
	row = hlp.TableSetHeaders(view, row, col, titles)

	setCell := func(row, col int, value string, align int) int {
		tableCell := kview.NewTableCell(value)
		tableCell.SetAlign(align)
		tableCell.SetSelectable(false)
		pm.mapView.SetCell(row, col, tableCell)

		return 0
	}

	row = 1
	col = 0
	for _, cl := range coreList {
		if cl.Mode == gpc.UnknownMode {
			continue
		}
		if cl.Mode == gpc.MainMode {
			setCell(row, col, cz.Green(fmt.Sprintf("%d-Main", cl.Core)), kview.AlignCenter)
			row++
			continue
		} else {
			setCell(row, 0, cz.Green(cl.Core), kview.AlignCenter)
		}
		if cl.LPort.Port.Pid < gpc.MaxEtherPorts {
			col = int(cl.LPort.Port.Pid) + 1
		} else {
			col = 1
		}
		m := cl.Mode
		switch m {
		case gpc.UnknownMode:
			setCell(row, col, cz.Red(m.String()), kview.AlignCenter)
		case gpc.MainMode:
			setCell(row, col, cz.Green(m.String()), kview.AlignCenter)
		case gpc.RxMode:
			setCell(row, col, cz.Cyan(m.String()), kview.AlignCenter)
		case gpc.TxMode:
			setCell(row, col, cz.GoldenRod(m.String()), kview.AlignCenter)
		case gpc.RxTxMode:
			setCell(row, col, cz.Blue(m.String()), kview.AlignCenter)
		}

		row++
	}
}

func (pm *PanelMapping) displayJSON(view *kview.TextView) {

	view.SetText(pktgenApp.gPkt.Marshal())
}
