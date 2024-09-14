// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package main

import (
	"fmt"
	"math/rand"

	"code.rocketnine.space/tslocum/cview"
	"github.com/gdamore/tcell/v2"

	"github.com/pktgen/go-pktgen/internal/configview"
	"github.com/pktgen/go-pktgen/internal/meter"
	"github.com/pktgen/go-pktgen/internal/perfview"
	"github.com/pktgen/go-pktgen/internal/portinfo"
	"github.com/pktgen/go-pktgen/internal/statsview"
	vp "github.com/pktgen/go-pktgen/internal/vpanels"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	tab "github.com/pktgen/go-pktgen/internal/taborder"
	tlog "github.com/pktgen/go-pktgen/internal/ttylog"
)

// PanelSingleMode - Data for main panel information
type PanelSingleMode struct {
	topFlex     *cview.Flex
	configView  *configview.ConfigView
	statsView   *statsview.StatsView
	perfView    *perfview.PerfView
	portInfo    *portinfo.PortInfo
	currentPort int
	to          *tab.Tab
	meter       *meter.Meter
	myInfo      vp.PanelMap
}

const (
	singlePanelName string = "Single"
	singleInfoHelp  string = "singleInfoHelp"
)

func init() {
	tlog.Register("SingleModeLogID")

	vp.Register(SingleModePanelSetup)
}

// setupSingleMode - setup and init the single mode page
func setupPanelSingleMode() *PanelSingleMode {

	ps := &PanelSingleMode{}

	return ps
}

// SingleModePanelSetup setup
func SingleModePanelSetup(pi vp.VPanelConfig) (*vp.VPanelData, error) {

	ps := setupPanelSingleMode()

	ps.to = tab.New(singlePanelName, pktgen.app)

	ps.portInfo = portinfo.New(pktgen.portCnt)

	topFlex := cview.NewFlex()
	topFlex.SetDirection(cview.FlexRow)
	flex1 := cview.NewFlex()
	flex1.SetDirection(cview.FlexRow)
	flex2 := cview.NewFlex()
	flex2.SetDirection(cview.FlexRow)
	flex2.SetTitle("Stats (S)")
	flex2.SetBorder(true)
	flex2.SetTitleAlign(cview.AlignLeft)
	flex3 := cview.NewFlex()
	flex3.SetDirection(cview.FlexRow)

	hlp.TitleBox(topFlex, PktgenInfo(true))

	ps.configView = configview.Create(panels, ps.to, pktgen.portCnt, flex1)
	topFlex.AddItem(flex1, 11, 0, true)
	for port := 0; port < pktgen.portCnt; port++ {
		s := ps.configView.FormName(port)
		tlog.DoPrintf("Form Name: %v\n", s)
		f := ps.configView.ConfigForm(port)
		AddModalPage(s, f)
	}

	ps.statsView = statsview.Create(pktgen.portCnt, flex2)
	topFlex.AddItem(flex2, 23, 0, true)

	ps.perfView = perfview.Create(pktgen.portCnt, flex3, "p")
	topFlex.AddItem(flex3, 0, 1, true)

	if err := ps.to.Add("singleConfig", ps.configView.TableView(), 'c'); err != nil {
		panic(err)
	}
	if err := ps.to.Add("singleStats", ps.statsView.TableView(), 'S'); err != nil {
		panic(err)
	}
	if err := ps.to.Add("singlePerf", ps.perfView.TextView(), 'p'); err != nil {
		panic(err)
	}
	if err := ps.to.SetInputDone(); err != nil {
		panic(err)
	}

	ps.topFlex = topFlex

	pktgen.timers.Add(singlePanelName, func(step int, ticks uint64) {
		if ps.topFlex.HasFocus() {
			pktgen.app.QueueUpdateDraw(func() {
				ps.displaySingleMode(step, ticks)
			})
		}
	})

	modal := cview.NewModal()
	modal.SetText("This is the Help Box: singleInfoHelp Thank you for asking for help! Press Esc to close.")
	modal.AddButtons([]string{"Got it"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pages.HidePage(singleInfoHelp)
	})
	AddModalPage(singleInfoHelp, modal)

	tv := ps.configView.TableView()
	cv := ps.configView
	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		ps.currentPort, _ = tv.GetSelection()
		ps.currentPort--

		sc := cv.PacketConfigByPort(ps.currentPort)

		k := event.Rune()
		switch k {
		case 'e': // Edit the packet configuration
			cv.ConfigForm(ps.currentPort).SetTitle(hlp.TitleColor(fmt.Sprintf("Edit Port %d", ps.currentPort)))
			s := ps.configView.FormName(ps.currentPort)
			tlog.DoPrintf("Port %d Form Name: %v\n", ps.currentPort, s)
			pages.ShowPage(ps.configView.FormName(ps.currentPort))
		case 's': // Start/Stop a single port transmitting
			if sc.TxState {
				sc.TxState = false
				ps.statsView.TxPercentSet(ps.currentPort, 0) // Set rate to zero
			} else {
				sc.TxState = true
			}
		case 'a': // Start all ports transmitting
			for i := 0; i < pktgen.portCnt; i++ {
				cv.SetTxState(i, true)
			}
		case 'A': // Stop all ports transmitting
			for i := 0; i < pktgen.portCnt; i++ {
				cv.SetTxState(i, false)
				ps.statsView.TxPercentSet(i, 0) // Set rate to zero
			}
		default:
			ps.to.SetInputFocus(k)
		}
		return event
	})
	topFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		k := event.Rune()
		switch k {
		case '?':
			pages.ShowPage(singleInfoHelp)
		default:
		}
		return event
	})

	ps.meter = meter.New().
		SetWidth(func() int {
			_, _, width, _ := ps.perfView.TextView().GetInnerRect()

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

		return &vp.VPageData{
			PanelName: p.infoName(),
			HelpName:  p.infoHelp(),
			TopFlex:   topFlex,
			TimerFn: func(step int, ticks uint64) {
				if step == -1 || topFlex.HasFocus() {
					app.QueueUpdateDraw(func() {
						ticks++
						switch step {
						case -1: // first time initial call
							p.docker.UpdateAll()
							p.displayContainers(p.containerTable)
							p.displayProcessStatus(p.psTable)
							p.displayNetwork(p.networkTable)
							p.displayImages(p.imageTable)

						case 0:
							p.displayContainers(p.containerTable)
							p.displayProcessStatus(p.psTable)
							p.displayNetwork(p.networkTable)
							p.displayImages(p.imageTable)
						}
					})
				}
			},
		}, nil
	}

// Callback timer routine to display the panels
func (ps *PageSingleMode) displaySingleMode(step int, ticks uint64) {

	sv := ps.statsView
	cv := ps.configView
	pv := ps.perfView
	switch step {
	case 0:
		ps.pullStats()

	case 2:
		cv.DisplayConfigTable()
		sv.DisplayStats()
		pv.DisplayPerf(ps.meter, sv.RxPercentArray(), sv.TxPercentArray())
	}
}

func (ps *PageSingleMode) pullStats() {

	sv := ps.statsView
	cv := ps.configView
	for port := 0; port < pktgen.portCnt; port++ {
		sv.RxPercentSet(port, float64(rand.Intn(100)))
		if cv.TxState(port) {
			sv.TxPercentSet(port, float64(rand.Intn(100)))
		}
	}
}
