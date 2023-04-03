// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022-2024 Intel Corporation

package perfview

import (
	"fmt"
	"sync"

	"code.rocketnine.space/tslocum/cview"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	"github.com/pktgen/go-pktgen/internal/meter"
)

type PerfView struct {
	portCnt   int
	pTextView *cview.TextView
	pOnce     sync.Once
}

func Create(portCnt int, flex *cview.Flex, tabChar string) *PerfView {
	perfView := &PerfView{portCnt: portCnt}

	perfView.pTextView = hlp.CreateTextView(flex, "Performance ("+tabChar+")", cview.AlignLeft, 18, 0, true)

	return perfView
}

func (pv *PerfView) TextView() *cview.TextView {
	return pv.pTextView
}

// Grab the load data and display the meters
func (pv *PerfView) DisplayPerf(m *meter.Meter, rxPercent []float64, txPercent []float64) {

	view := pv.TextView()

	str := ""

	for i := 0; i < pv.portCnt; i++ {
		str += m.Draw(rxPercent[i], &meter.Info{
			Labels: []*meter.LabelInfo{ // line format 'xx:Rx xxx.xx [bar meter]'
				{Val: fmt.Sprintf("%2d", i), Fn: cz.Cyan},                   // Port number in color
				{Val: ":", Fn: nil},                                         // Colon between port number and label
				{Val: "Rx ", Fn: cz.Yellow},                                 // Receive label in color
				{Val: fmt.Sprintf("%6.2f ", rxPercent[i]), Fn: cz.DeepPink}, // Receive percent in color
			},
			Bar: &meter.LabelInfo{Val: "", Fn: cz.MediumSpringGreen}, // Meter bar in color
		})
		str += m.Draw(txPercent[i], &meter.Info{
			Labels: []*meter.LabelInfo{ // line format 'xx:Tx xxx.xx [bar meter]'
				{Val: "  ", Fn: nil},
				{Val: " ", Fn: nil},
				{Val: "Tx ", Fn: cz.Blue},
				{Val: fmt.Sprintf("%6.2f ", txPercent[i]), Fn: cz.DeepPink},
			},
			Bar: &meter.LabelInfo{Val: "", Fn: cz.Blue},
		})
	}
	str = str[:len(str)-1] // Strip the last newline character

	view.SetText(str)
	pv.pOnce.Do(func() {
		view.ScrollToBeginning()
	})
}
