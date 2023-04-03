// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/pktgen/go-pktgen/internal/cfg"
	"github.com/pktgen/go-pktgen/internal/cpudata"
	"github.com/pktgen/go-pktgen/internal/dbg"
	"github.com/pktgen/go-pktgen/internal/devbind"
	"github.com/pktgen/go-pktgen/pkgs/kview"

	flags "github.com/jessevdk/go-flags"
	cz "github.com/pktgen/go-pktgen/internal/colorize"
	et "github.com/pktgen/go-pktgen/internal/etimers"
	"github.com/pktgen/go-pktgen/internal/tlog"
	vp "github.com/pktgen/go-pktgen/internal/vpanels"
)

const (
	// pktgenVersion string
	goPktgenVersion = "24.09.0"
)

const (
	SinglePanelIndex = iota
	SysInfoPanelIndex
	CPUPanelIndex
)

// Panels is a function which returns the feature's main primitive and its title.
// It receives a "nextFeature" function which can be called to advance the
// presentation to the next slide.
type Panels func(pages *kview.Panels, nextPanel func()) (title string, content kview.Primitive)

type ModalPage struct {
	title string
	modal interface{}
}

// Pktgen for monitoring and system performance data
type Pktgen struct {
	version    string             // Version of Pktgen
	dbg        *dbg.DbgInfo       // Debugging information
	app        *kview.Application // Application or top level application
	panels     *kview.Panels      // Panels for presentation
	timers     *et.EventTimers    // Event Timers
	cpuData    *cpudata.CPUData   // CPU data
	portCnt    int                // Maximum number of ports
	cfg        *cfg.System        // Configuration system
	ModalPages []*ModalPage       // Modal pages
	db         *devbind.DevBind
}

// Options command line options
type Options struct {
	ConfigData  string `short:"c" long:"config-data" description:"JSON configuration file or string"`
	Ptty		int    `short:"p" long:"ptty" description:"Enable pseudo-TTY mode (for debugging)" default:"0"`
	ShowVersion bool   `short:"V" long:"version" description:"Print out version and exit"`
	Verbose     bool   `short:"v" long:"verbose" description:"Output verbose messages"`
}

// Global to the main package for the tool
var (
	pktgen  Pktgen
	options Options
	parser  = flags.NewParser(&options, flags.Default)
)

const (
	mainLog = "MainLogID"
)

func main() {

	if err := initializePktgen(); err != nil {
		log.Fatalf("Initialization failed: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Go-Pktgen: Defer tlog.Close()\n")
	defer func() {
		fmt.Printf("Go-Pktgen: Stopping tlog...\n")
        tlog.Close()
    }()

	pktgen.portCnt = pktgen.cfg.PortCount()
	if options.ShowVersion {
		fmt.Printf("Go-Pktgen Version: %s\n", pktgen.version)
		return
	}

	// Create the main panel.
	topFlex := kview.NewFlex()
	topFlex.SetDirection(kview.FlexRow)

	pktgen.app.SetRoot(topFlex, true)
	pktgen.app.EnableMouse(true)

	pktgen.db = devbind.New(devbind.WithTimeout(5))

	pktgen.db.Start()

	if err := pktgen.db.BindPorts(pktgen.cfg.PortList()); err != nil {
		fmt.Printf("Go-Pktgen: %v\n", err)
		for k, v := range pktgen.db.HwInfo() {
			fmt.Printf("hwInfo: %v : %v - %v\n", k, v.BusInfo, v.Config.Driver)
		}
		return
	}
	fmt.Printf("Go-Pktgen: Defer devbind stop\n")
	defer func() {
		fmt.Printf("Go-Pktgen: Stopping devbind...\n")
		pktgen.db.Stop()
	}()

	pktgen.timers = et.New(et.WithTimeout(2), et.WithSteps(4))
	pktgen.timers.Start()

	// The bottom row has some info on where we are.
	info := kview.NewTextView()
	info.SetDynamicColors(true)
	info.SetRegions(true)
	info.SetWrap(false)

	currentPanel := 0
	info.Highlight(strconv.Itoa(currentPanel))

	previousPanel := func() {
		pktgen.panels.HidePanel(vp.NameByIndex(currentPanel))
		currentPanel = (currentPanel - 1 + vp.Count()) % vp.Count()
		info.Highlight(vp.NameByIndex(currentPanel))
		info.ScrollToHighlight()
		pktgen.panels.ShowPanel(vp.NameByIndex(currentPanel))
		info.SetText(buildPanelString(currentPanel))
	}

	nextPanel := func() {
		pktgen.panels.HidePanel(vp.NameByIndex(currentPanel))
		currentPanel = (currentPanel + 1) % vp.Count()
		info.Highlight(vp.NameByIndex(currentPanel))
		info.ScrollToHighlight()
		pktgen.panels.ShowPanel(vp.NameByIndex(currentPanel))
		info.SetText(buildPanelString(currentPanel))
	}

	vp.Call()

	for index := 0; index < vp.Count(); index++ {
		info := vp.GetInfo(index)
		pktgen.panels.AddPanel(vp.NameByIndex(index), info.TopFlex, true, index == currentPanel)
		pktgen.timers.Add(info.PanelName, info.TimerFn)
	}

	for _, m := range pktgen.ModalPages {
		pktgen.panels.AddPanel(m.title, m.modal.(kview.Primitive), false, false)
	}

	info.SetText(buildPanelString(SinglePanelIndex))

	// Create the main panel.
	topFlex.AddItem(pktgen.panels, 0, 1, true)
	topFlex.AddItem(info, 1, 0, false)

	// Shortcuts to navigate the panels.
	pktgen.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			nextPanel()
		} else if event.Key() == tcell.KeyCtrlP {
			previousPanel()
		} else if event.Key() == tcell.KeyCtrlQ {
			pktgen.app.Stop()
		} else if event.Rune() == '?' {
			pktgen.panels.ShowPanel(vp.GetInfo(currentPanel).HelpID)
		} else {
			var idx int

			switch {
			case event.Key() >= tcell.KeyF1 && event.Key() <= tcell.KeyF19:
				idx = int(event.Key() - tcell.KeyF1)
			case event.Rune() == 'q':
				pktgen.app.Stop()
			default:
				idx = -1
			}
			if idx != -1 {
				if idx < vp.Count() {
					pktgen.panels.HidePanel(vp.NameByIndex(currentPanel))
					currentPanel = idx
					info.Highlight(strconv.Itoa(currentPanel))
					info.ScrollToHighlight()
					pktgen.panels.ShowPanel(vp.NameByIndex(currentPanel))
				}
				info.SetText(buildPanelString(idx))
			}
		}
		return event
	})

	if err := gpktApiStart(pktgen.cfg); err != nil {
		panic(err)
	}

	fmt.Printf("Go-Pktgen: Defer stopping gPkt\n")
	defer func() {
		fmt.Printf("Go-Pktgen: Stopping gpkt...\n")
		gpktApiStop()
		fmt.Printf("Go-Pktgen: Stopped gpkt...\n")
	}()

	// Start the application.
	pktgen.app.SetRoot(topFlex, true)
	pktgen.app.EnableMouse(true)
	if err := pktgen.app.Run(); err != nil {
		panic(err)
	}
	fmt.Printf("Go-Pktgen: Stopping...\n")
}

func setupSignals(signals ...os.Signal) {

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, signals...)
	go func() {
		sig := <-sigs

		fmt.Printf("Signal: %v\n", sig)
		pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)

		os.Exit(1)
	}()
}

// Setup the main Pktgen application structure.
func initializePktgen() error {
	setupSignals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)

	cz.SetDefault("ivory", "", 0, 2, "")

	pktgen = Pktgen{
		version: goPktgenVersion,
		dbg:     dbg.New(),
		app:     kview.NewApplication(),
		panels:  kview.NewPanels(),
		portCnt: pktgen.portCnt,
	}

	_, err := parser.Parse()
	if err != nil {
		fmt.Printf("*** invalid arguments %v\n", err)
		os.Exit(1)
	}

	cs := cfg.New()
	if err := cs.Open(options.ConfigData); err != nil {
		fmt.Printf("load configuration failed: %s\n", err)
		os.Exit(1)
	}
	pktgen.cfg = cs

	// Command line options override configuration file.
	if options.Ptty > 0 {
		cs.SetDebugTTY(options.Ptty)
    }

	if cs.DebugTTY() > 0 {
		err = tlog.Open(cs.DebugTTY())
		if err != nil {
			fmt.Printf("tlog open failed: %s\n", err)
			os.Exit(1)
		}
		tlog.Register(mainLog, true)
	}

	str := PktgenInfo(false)
	tlog.DoPrintf("\n===== %s =====\n", str)
	fmt.Printf("\n===== %s =====\n", str)

	cd, err := cpudata.New()
	if err != nil {
		return fmt.Errorf("new CPU data failed: %s", err)
	}
	pktgen.cpuData = cd

	pktgen.dbg.SetPrintState(true)

	vp.Initialize(pktgen.panels, pktgen.app)

	return nil
}
