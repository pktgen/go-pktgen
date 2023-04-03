// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"code.rocketnine.space/tslocum/cview"
	"github.com/gdamore/tcell/v2"
	"github.com/pktgen/go-pktgen/internal/cfg"
	"github.com/pktgen/go-pktgen/internal/cpudata"
	"github.com/pktgen/go-pktgen/internal/dbg"
	"github.com/pktgen/go-pktgen/internal/etimers"

	flags "github.com/jessevdk/go-flags"
	cz "github.com/pktgen/go-pktgen/internal/colorize"
	tlog "github.com/pktgen/go-pktgen/internal/ttylog"
	vp "github.com/pktgen/go-pktgen/internal/vpanels"
)

const (
	// pktgenVersion string
	goPktgenVersion = "24.09.0"
)

// PanelInfo for title and primitive
type PanelInfo struct {
	title     string
	primitive cview.Primitive
}

// Panels is a function which returns the feature's main primitive and its title.
// It receives a "nextFeature" function which can be called to advance the
// presentation to the next slide.
type Panels func(pages *cview.Panels, nextPanel func()) (title string, content cview.Primitive)

type ModalPage struct {
	title string
	modal interface{}
}

// Pktgen for monitoring and system performance data
type Pktgen struct {
	version    string             // Version of Pktgen
	dbg        *dbg.DbgInfo       // Debugging information
	app        *cview.Application // Application or top level application
	timers     *etimers.EventTimers
	cpuData    *cpudata.CPUData
	panels     []PanelInfo
	portCnt    int
	ModalPages []*ModalPage
}

// Options command line options
type Options struct {
	ConfigData  string `short:"c" long:"config-data" description:"JSON configuration file or string"`
	Ptty        string `short:"p" long:"ptty" description:"path to ptty /dev/pts/X"`
	PortCnt     uint   `short:"P" long:"port-count" description:"Max number of ports to use" default:"8"`
	ShowVersion bool   `short:"V" long:"version" description:"Print out version and exit"`
	Verbose     bool   `short:"v" long:"verbose" description:"Output verbose messages"`
}

// Global to the main package for the tool
var pktgen Pktgen
var options Options
var parser = flags.NewParser(&options, flags.Default)

const (
	mainLog = "MainLogID"
)

func buildPanelString(idx int) string {
	// Build the panel selection string at the bottom of the xterm and
	// highlight the selected tab/panel item.
	s := ""
	for index, p := range pktgen.panels {
		if index == idx {
			s += fmt.Sprintf("F%d:[orange::r]%s[white::-]", index+1, p.title)
		} else {
			s += fmt.Sprintf("F%d:[orange::-]%s[white::-]", index+1, p.title)
		}
		if (index + 1) < len(pktgen.panels) {
			s += " "
		}
	}
	return s
}

// Setup the tool's global information and startup the process info connection
func init() {
	tlog.Register(mainLog, true)

	pktgen = Pktgen{}
	pktgen.version = goPktgenVersion
	pktgen.dbg = dbg.New()
	pktgen.dbg.SetPrintState(true)

	// Create the main cview application.
	pktgen.app = cview.NewApplication()

	cd, err := cpudata.New()
	if err != nil {
		fmt.Printf("New CPU data failed: %s\n", err)
		return
	}
	pktgen.cpuData = cd
	pktgen.portCnt = 8
}

// Version number string
func Version() string {
	return pktgen.version
}

func AddModalPage(title string, modal interface{}) {
	pktgen.ModalPages = append(pktgen.ModalPages, &ModalPage{title: title, modal: modal})
}

func main() {

	setupSignals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)

	cz.SetDefault("ivory", "", 0, 2, "")

	_, err := parser.Parse()
	if err != nil {
		fmt.Printf("*** invalid arguments %v\n", err)
		os.Exit(1)
	}

	if len(options.Ptty) > 0 {
		err = tlog.Open(options.Ptty)
		if err != nil {
			fmt.Printf("ttylog open failed: %s\n", err)
			os.Exit(1)
		}
	}
	pktgen.portCnt = int(options.PortCnt)
	if options.ShowVersion {
		fmt.Printf("Go-Pktgen Version: %s\n", pktgen.version)
		return
	}

	fmt.Printf("Config: %s\n", options.ConfigData)

	cs := cfg.New()
	if err := cs.Open(options.ConfigData); err != nil {
		fmt.Printf("load configuration failed: %s\n", err)
		os.Exit(1)
	}

	str := PktgenInfo(false)
	tlog.Log(mainLog, "\n===== %s =====\n", str)
	fmt.Printf("\n===== %s =====\n", str)

	app := pktgen.app

	pktgen.timers = etimers.New(time.Second/4, 4)
	pktgen.timers.Start()

	// The bottom row has some info on where we are.
	info := cview.NewTextView()
	info.SetDynamicColors(true)
	info.SetRegions(true)
	info.SetWrap(false)

	currentPanel := 0
	info.Highlight(strconv.Itoa(currentPanel))

	panels := cview.NewPanels()
	panel := cview.NewFlex()

	previousPanel := func() {
		currentPanel = (currentPanel - 1 + len(panels)) % len(panels)
		info.Highlight(strconv.Itoa(currentPanel))
		info.ScrollToHighlight()
		pages.SwitchToPage(strconv.Itoa(currentPanel))
		info.SetText(buildPanelString(currentPanel))
	}

	nextPanel := func() {
		currentPanel = (currentPanel + 1) % len(panels)
		info.Highlight(strconv.Itoa(currentPanel))
		info.ScrollToHighlight()
		pages.SwitchToPage(strconv.Itoa(currentPanel))
		info.SetText(buildPanelString(currentPanel))
	}

	for index, f := range panels {
		title, primitive := f(panels, nextPanel)
		pages.AddPage(strconv.Itoa(index), primitive, true, index == currentPanel)
		pktgen.panels = append(pktgen.panels, PanelInfo{title: title, primitive: primitive})
	}

	for _, m := range pktgen.ModalPages {
		pages.AddPage(m.title, m.modal.(cview.Primitive), false, false)
	}

	info.SetText(buildPanelString(0))

	// Create the main panel.
	panel.SetDirection(cview.FlexRow)
	panel.AddItem(pages, 0, 1, true)
	panel.AddItem(info, 1, 1, false)

	// Shortcuts to navigate the panels.
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			nextPanel()
		} else if event.Key() == tcell.KeyCtrlP {
			previousPanel()
		} else if event.Key() == tcell.KeyCtrlQ {
			app.Stop()
		} else {
			var idx int

			switch {
			case event.Key() >= tcell.KeyF1 && event.Key() <= tcell.KeyF19:
				idx = int(event.Key() - tcell.KeyF1)
			case event.Rune() == 'q':
				app.Stop()
			default:
				idx = -1
			}
			if idx != -1 {
				if idx < len(panels) {
					currentPanel = idx
					info.Highlight(strconv.Itoa(currentPanel))
					info.ScrollToHighlight()
					pages.SwitchToPage(strconv.Itoa(currentPanel))
				}
				info.SetText(buildPanelString(idx))
			}
		}
		return event
	})

	if err := gpktInit(cs); err != nil {
		panic(err)
	}

	// Start the application.
	app.SetRoot(panel, true)
	app.EnableMouse(true)
	if err := app.Run(); err != nil {
		panic(err)
	}

	tlog.Log(mainLog, "===== Done =====\n")
}

func setupSignals(signals ...os.Signal) {
	app := pktgen.app

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, signals...)
	go func() {
		sig := <-sigs

		tlog.Log(mainLog, "Signal: %v\n", sig)
		time.Sleep(time.Second)

		app.Stop()
		os.Exit(1)
	}()
}
