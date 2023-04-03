// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package main

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/pktgen/go-pktgen/pkgs/kview"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	et "github.com/pktgen/go-pktgen/internal/etimers"
	vp "github.com/pktgen/go-pktgen/internal/vpanels"
)

func registerVPanels() error {

	panelSetups := []func() (string, vp.VPanelFunc){
		SingleModePanel,
		SysInfoPanel,
		CPULoadPanel,
		MappingPanel,
	}

	for idx, fn := range panelSetups {
		name, setup := fn()
		err := vp.GetVPanels().Register(name, idx, setup)
		if err != nil {
			return fmt.Errorf("registering panel: %v\n", err)
		}
	}

	return nil
}

// Setup the main Pktgen application structure.
func (pg *PktgenApp) initPanels() error {

	cz.SetDefault("ivory", "", 0, 2, "")

	// Create the main application view.
	if app := kview.NewApplication(); app != nil {
		pg.appView = app
	} else {
		return fmt.Errorf("failed to create kview application")
	}

	if panels := kview.NewPanels(); panels == nil {
		return fmt.Errorf("failed to create panels")
	} else {
		pg.kPanels = panels
	}
	vp.Initialize(pg.kPanels, pg.appView)

	if err := registerVPanels(); err != nil {
		return err
	}

	vp := vp.GetVPanels()
	if err := vp.Call(); err != nil { // Call the panel functions to create the panels.
		return err
	}

	// Setup the application windows and structures.
	if err := pg.setupInputCapture(); err != nil {
		return err
	}

	return nil
}

func (pg *PktgenApp) setupInputCapture() error {

	// Create the main panel.
	topFlex := kview.NewFlex()
	topFlex.SetDirection(kview.FlexRow)

	timers := et.New(et.WithTimeout(1), et.WithSteps(4))
	timers.Start()

	// The bottom row has some info on where we are.
	info := kview.NewTextView()
	info.SetDynamicColors(true)
	info.SetRegions(true)
	info.SetWrap(false)

	currentPanel := 0
	info.Highlight(strconv.Itoa(currentPanel))

	vp := vp.GetVPanels()
	previousPanel := func() {
		pg.kPanels.HidePanel(vp.NameByIndex(currentPanel))
		currentPanel = (currentPanel - 1 + vp.Count()) % vp.Count()
		info.Highlight(vp.NameByIndex(currentPanel))
		info.ScrollToHighlight()
		pg.kPanels.ShowPanel(vp.NameByIndex(currentPanel))
		info.SetText(buildPanelFooter(currentPanel))
	}

	nextPanel := func() {
		pg.kPanels.HidePanel(vp.NameByIndex(currentPanel))
		currentPanel = (currentPanel + 1) % vp.Count()
		info.Highlight(vp.NameByIndex(currentPanel))
		info.ScrollToHighlight()
		pg.kPanels.ShowPanel(vp.NameByIndex(currentPanel))
		info.SetText(buildPanelFooter(currentPanel))
	}

	for index := 0; index < vp.Count(); index++ {
		info := vp.GetInfo(index)
		pg.kPanels.AddPanel(vp.NameByIndex(index), info.TopFlex, true, index == currentPanel)
		timers.Add(info.PanelName, info.TimerFn)
	}

	info.SetText(buildPanelFooter(0)) // Display the initial panel info.

	// Create the main panel.
	topFlex.AddItem(pg.kPanels, 0, 1, true)
	topFlex.AddItem(info, 1, 0, false)

	// Shortcuts to navigate the panels.
	pg.appView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			nextPanel()
		} else if event.Key() == tcell.KeyCtrlP {
			previousPanel()
		} else if event.Key() == tcell.KeyCtrlQ {
			pg.appView.Stop()
		} else if event.Rune() == '?' {
			pg.kPanels.ShowPanel(vp.GetInfo(currentPanel).HelpID)
			pg.kPanels.SendToFront(vp.GetInfo(currentPanel).HelpID)
		} else {
			var idx int

			switch {
			case event.Key() >= tcell.KeyF1 && event.Key() <= tcell.KeyF19:
				idx = int(event.Key() - tcell.KeyF1)
			case event.Rune() == 'q':
				pg.appView.Stop()
			default:
				idx = -1
			}
			if idx != -1 {
				if idx < vp.Count() {
					pg.kPanels.HidePanel(vp.NameByIndex(currentPanel))
					currentPanel = idx
					info.Highlight(strconv.Itoa(currentPanel))
					info.ScrollToHighlight()
					pg.kPanels.ShowPanel(vp.NameByIndex(currentPanel))
				}
				info.SetText(buildPanelFooter(idx))
			}
		}
		return event
	})

	// Start the application.
	pg.appView.SetRoot(topFlex, true)
	pg.appView.EnableMouse(true)

	return nil
}

func buildPanelFooter(idx int) string {

	vp := vp.GetVPanels()
	// Build the panel selection string at the bottom of the xterm and
	// highlight the selected tab/panel item.
	s := ""
	for index, p := range vp.GetPanelNames() {
		if index == idx {
			s += fmt.Sprintf("F%d:[orange::r]%s[white::-]", index+1, p)
		} else {
			s += fmt.Sprintf("F%d:[orange::-]%s[white::-]", index+1, p)
		}
		if (index + 1) < vp.Count() {
			s += " "
		}
	}
	return s
}
