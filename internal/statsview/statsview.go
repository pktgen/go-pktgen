// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022-2024 Intel Corporation

package statsview

import (
	"fmt"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"code.rocketnine.space/tslocum/cview"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
)

type StatsView struct {
	portCnt       int
	sTable        *cview.Table
	sOnce         sync.Once
	rxPercentRate []float64
	txPercentRate []float64
}

func Create(portCnt int, flex *cview.Flex) *StatsView {
	statsView := &StatsView{portCnt: portCnt}

	statsView.sTable = hlp.CreateTableView(flex, "", cview.AlignLeft, 0, 1, true)
	statsView.sTable.SetSelectable(false, false)
	statsView.sTable.SetFixed(1, 1)
	statsView.sTable.SetSeparator(cview.Borders.Vertical)

	statsView.rxPercentRate = make([]float64, portCnt)
	statsView.txPercentRate = make([]float64, portCnt)

	return statsView
}

func (sv *StatsView) TableView() *cview.Table {
	return sv.sTable
}

func (sv *StatsView) RxPercent(idx int) float64 {
	return sv.rxPercentRate[idx]
}

func (sv *StatsView) RxPercentSet(idx int, val float64) {
	sv.rxPercentRate[idx] = val
}

func (sv *StatsView) RxPercentArray() []float64 {
	return sv.rxPercentRate
}

func (sv *StatsView) TxPercent(idx int) float64 {
	return sv.txPercentRate[idx]
}

func (sv *StatsView) TxPercentSet(idx int, val float64) {
	sv.txPercentRate[idx] = val
}

func (sv *StatsView) TxPercentArray() []float64 {
	return sv.txPercentRate
}

func (sv *StatsView) DisplayStats() {

	table := sv.sTable
	row := 0

	titles := []string{
		cz.CornSilk("Link State", 12),
		cz.Cyan("Rx pps", 8),
		cz.Cyan("Tx pps", 8),
		cz.Cyan("Rx/Tx Mbits", 12),
		cz.Cyan("Rx Max", 8),
		cz.Cyan("Tx Max", 8),
		cz.Cyan("Rx/Tx Errors", 12),
		cz.Cyan("Tot Rx Pkts", 12),
		cz.Cyan("Tot Tx Pkts", 12),
		cz.Cyan("Tot Rx Mbits", 12),
		cz.Cyan("Tot Tx Mbits", 12),
		cz.Cyan("Broadcast", 12),
		cz.Cyan("Multicast", 12),
		cz.Cyan("Sizes 64", 12),
		cz.Cyan("128-255", 12),
		cz.Cyan("256-511", 12),
		cz.Cyan("512-1023", 12),
		cz.Cyan("1024-1518", 12),
		cz.Cyan("Runts/Jumbos", 14),
		cz.Cyan("ARPs/ICMPs", 14),
	}

	t := make([]string, 0)
	t = append(t, "")
	for i := 0; i < sv.portCnt; i++ {
		if i == 0 {
			t = append(t, cz.Orange(fmt.Sprintf("Port %d", i)))
		} else {
			t = append(t, cz.Orange(fmt.Sprintf("%v", i)))
		}
	}

	hlp.TableSetHeaders(table, 0, 0, t)
	hlp.TableSetRows(table, 1, 0, titles)

	p := message.NewPrinter(language.English)

	comma := func(n interface{}) string {
		return p.Sprintf("%d", n)
	}

	for v := 0; v < sv.portCnt; v++ {

		rowData := []string{
			cz.LightYellow("UP-40000-FD"),
			cz.Cyan(comma(0)),
			cz.Cyan(comma(0)),
			cz.Wheat(comma(0) + "/" + comma(0)),
			cz.Cyan(comma(0)),
			cz.Cyan(comma(0)),
			cz.Red(comma(0) + "/" + comma(0)),
			cz.Cyan(comma(0)),
			cz.Cyan(comma(0)),
			cz.Cyan(comma(0)),
			cz.Cyan(comma(0)),
			cz.Wheat(0),
			cz.GoldenRod(0),
			cz.Cyan(0),
			cz.Cyan(0),
			cz.Cyan(0),
			cz.Cyan(0),
			cz.Cyan(0),
			cz.DeepPink("0/0"),
			cz.Wheat("0/0"),
		}

		row = 1
		for _, d := range rowData {
			hlp.TableCellSet(table, row, v+1, d)
			row++
		}
	}

	sv.sOnce.Do(func() {
		sv.sTable.ScrollToBeginning()
	})
}
