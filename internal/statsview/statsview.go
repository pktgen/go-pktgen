// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022-2024 Intel Corporation

package statsview

import (
	"fmt"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/pktgen/go-pktgen/pkgs/kview"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
)

type StatsView struct {
	portCnt       int
	sTable        *kview.Table
	sOnce         sync.Once
	rxPercentRate []float64
	txPercentRate []float64
}

func Create(portCnt int, flex *kview.Flex, tabChar string) *StatsView {
	statsView := &StatsView{portCnt: portCnt}

	statsView.sTable = hlp.CreateTableView(flex, hlp.NewText("Statistics ("+tabChar+")", kview.AlignLeft), 0, 1, true)
	statsView.sTable.SetSelectable(false, false)
	statsView.sTable.SetFixed(2, 1)
	statsView.sTable.SetSeparator(kview.Borders.Vertical)

	statsView.rxPercentRate = make([]float64, portCnt)
	statsView.txPercentRate = make([]float64, portCnt)

	return statsView
}

func (sv *StatsView) TableView() *kview.Table {
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
	table.Clear()

	row := 0
	width := -14
	titles := []hlp.TextInfo{
		hlp.NewText(cz.CornSilk("Link State", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Rx pps", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Tx pps", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Rx/Tx Mbits", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Rx Max", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Tx Max", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Rx/Tx Errors", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Tot Rx Pkts", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Tot Tx Pkts", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Tot Rx Mbits", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Tot Tx Mbits", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Broadcast", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Multicast", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Sizes 64", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("128-255", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("256-511", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("512-1023", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("1024-1518", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("Runts/Jumbos", width), kview.AlignLeft),
		hlp.NewText(cz.Cyan("ARPs/ICMPs", width), kview.AlignLeft),
	}

	t := make([]hlp.TextInfo, 0)
	t = append(t, hlp.NewText("", kview.AlignLeft))
	for i := 0; i < sv.portCnt; i++ {
			t = append(t, hlp.NewText(cz.Orange(fmt.Sprintf("Port %2d", i), 14), kview.AlignRight))
	}

	hlp.TableSetHeaders(table, 0, 0, t)
	hlp.TableSetRows(table, 1, 0, titles)

	p := message.NewPrinter(language.English)

	comma := func(n interface{}) string {
		return p.Sprintf("%d", n)
	}

	for v := 0; v < sv.portCnt; v++ {

		rowData := []hlp.TextInfo{
			hlp.NewText(cz.LightYellow("UP-100000-FD"), kview.AlignRight),
			hlp.NewText(cz.Cyan(comma(0)), kview.AlignRight),
			hlp.NewText(cz.Cyan(comma(0)), kview.AlignRight),
			hlp.NewText(cz.Wheat(comma(0) + "/" + comma(0)), kview.AlignRight),
			hlp.NewText(cz.Cyan(comma(0)), kview.AlignRight),
			hlp.NewText(cz.Cyan(comma(0)), kview.AlignRight),
			hlp.NewText(cz.Red(comma(0) + "/" + comma(0)), kview.AlignRight),
			hlp.NewText(cz.Cyan(comma(0)), kview.AlignRight),
			hlp.NewText(cz.Cyan(comma(0)), kview.AlignRight),
			hlp.NewText(cz.Cyan(comma(0)), kview.AlignRight),
			hlp.NewText(cz.Cyan(comma(0)), kview.AlignRight),
			hlp.NewText(cz.Wheat(0), kview.AlignRight),
			hlp.NewText(cz.GoldenRod(0), kview.AlignRight),
			hlp.NewText(cz.Cyan(0), kview.AlignRight),
			hlp.NewText(cz.Cyan(0), kview.AlignRight),
			hlp.NewText(cz.Cyan(0), kview.AlignRight),
			hlp.NewText(cz.Cyan(0), kview.AlignRight),
			hlp.NewText(cz.Cyan(0), kview.AlignRight),
			hlp.NewText(cz.DeepPink("0/0"), kview.AlignRight),
			hlp.NewText(cz.Wheat("0/0"), kview.AlignRight),
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
