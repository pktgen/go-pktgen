// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package configview

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/gdamore/tcell/v2"
	"github.com/pktgen/go-pktgen/pkgs/kview"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	"github.com/pktgen/go-pktgen/internal/constants"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	tab "github.com/pktgen/go-pktgen/internal/taborder"
)

const (
	editFormName string = "editForm"
)

type ConfigView struct {
	portCnt     uint16
	cTable      *kview.Table
	cOnce       sync.Once
	pConfig     []*constants.PacketConfig
	to          *tab.Tab
	formNames   []string
	configForms []*kview.Flex
	panels      *kview.Panels
}

func Create(panels *kview.Panels, to *tab.Tab, portCnt uint16, tabChar rune, flex *kview.Flex) *ConfigView {

	configView := &ConfigView{
		portCnt: portCnt,
		panels:  panels,
		to:      to,
	}

	configView.cTable = hlp.CreateTableView(flex,
		hlp.NewText(fmt.Sprintf("Configuration (%c) s:Toggle Start/Stop, a/A:All Start/Stop, e:Edit", tabChar),
			kview.AlignLeft), 0, 1, true)

	configView.cTable.SetSeparator(kview.Borders.Vertical)
	configView.cTable.SetSelectedStyle(tcell.ColorBlue, tcell.ColorYellow, 0)
	configView.cTable.SetSelectable(true, false)
	configView.cTable.SetFixed(1, 1)
	configView.cTable.SetSelectionChangedFunc(func(row, col int) {
		// Adjust for not selecting the first column in a row. table.go will select the next row.
		// on a mouse click
		if col > 0 {
			configView.cTable.Select(row, 0)
		}
	})

	configView.initConfigView(tabChar)

	return configView
}

func (cv *ConfigView) TableView() *kview.Table {
	return cv.cTable
}

func (cv *ConfigView) FormName(port uint16) string {
	return cv.formNames[port]
}

func (cv *ConfigView) ConfigForm(port uint16) *kview.Flex {
	return cv.configForms[port]
}

func (cv *ConfigView) PacketConfigByPort(port uint16) *constants.PacketConfig {
	return cv.pConfig[port]
}

func (cv *ConfigView) TxState(port uint16) bool {
	return cv.pConfig[port].TxState
}

func (cv *ConfigView) SetTxState(port uint16, val bool) {
	cv.pConfig[port].TxState = val
}

func (cv *ConfigView) initConfigView(tabChar rune) {

	cv.pConfig = make([]*constants.PacketConfig, cv.portCnt)
	cv.formNames = make([]string, cv.portCnt)

	for i := uint16(0); i < cv.portCnt; i++ {
		cv.pConfig[i] = &constants.PacketConfig{
			PortIndex:   i,
			TxCount:     0,
			PercentRate: 100.0,
			PktSize:     64,
			BurstCount:  128,
			TimeToLive:  64,
			SrcPort:     1245,
			DstPort:     5678,
			Proto:       "IPv4/UDP",
			VlanId:      1,
			DstIP:       net.IPNet{IP: net.IPv4(198, 18, 1, 1), Mask: net.CIDRMask(0, 32)},
			SrcIP:       net.IPNet{IP: net.IPv4(198, 18, 0, 1), Mask: net.CIDRMask(24, 32)},
			DstMAC:      []byte{0x12, 0x34, 0x45, 0x67, 0x89, 00},
			SrcMAC:      []byte{0x12, 0x34, 0x45, 0x67, 0x89, 01},
			TxState:     false,
		}
	}
	for port := uint16(0); port < cv.portCnt; port++ {
		f := cv.setupEditConfigForm(port, tabChar)
		cv.configForms = append(cv.configForms, f)
	}
}

func (cv *ConfigView) setupEditConfigForm(port uint16, tabChar rune) *kview.Flex {

	pg := fmt.Sprintf("%v-%v", editFormName, port)

	form := kview.NewForm()
	form.SetItemPadding(1)
	form.SetHorizontal(false)
	form.SetFieldTextColor(tcell.ColorCornsilk)
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetItemPadding(0)
	form.SetCancelFunc(func() {
		cv.panels.HidePanel(pg)
		cv.to.SetInputFocus(tabChar)
	})

	form.SetTitleAlign(kview.AlignLeft)
	form.SetRect(0, 0, 37, 20)

	sc := *cv.pConfig[port]

	form.AddInputField("TxCount    :", strconv.Itoa(int(sc.TxCount)), 15,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 15 && hlp.AcceptNumber(textToCheck, lastChar)
		}, func(text string) {
			hlp.ParseNumberUint64(text, &sc.TxCount)
		})

	form.AddInputField("Rate       :", strconv.FormatFloat(sc.PercentRate, 'f', 2, 64), 6,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 6 && hlp.AcceptFloat(textToCheck, lastChar)
		}, func(text string) {
			if err := hlp.ParseNumberFloat64(text, &sc.PercentRate); err == nil {
				if sc.PercentRate == 0 || sc.PercentRate > 100.00 {
					sc.PercentRate = 100.00
				}
			}
		})

	form.AddInputField("PktSize    :", strconv.Itoa(int(sc.PktSize)), 5,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 5 && hlp.AcceptNumber(textToCheck, lastChar)
		}, func(text string) {
			if err := hlp.ParseNumberUint16(text, &sc.PktSize); err == nil {
				if sc.PktSize < 64 {
					sc.PktSize = 64
				} else if sc.PktSize > 1522 {
					sc.PktSize = 1522
				}
			}
		})

	form.AddInputField("Burst      :", strconv.Itoa(int(sc.BurstCount)), 3,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 3 && hlp.AcceptNumber(textToCheck, lastChar)
		}, func(text string) {
			if err := hlp.ParseNumberUint16(text, &sc.BurstCount); err == nil {
				if sc.BurstCount < 32 {
					sc.BurstCount = 32
				} else if sc.BurstCount > 256 {
					sc.BurstCount = 256
				}
			}
		})

	form.AddInputField("TTL        :", strconv.Itoa(int(sc.TimeToLive)), 3,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 3 && hlp.AcceptNumber(textToCheck, lastChar)
		}, func(text string) {
			if err := hlp.ParseNumberUint16(text, &sc.TimeToLive); err == nil {
				if sc.TimeToLive > 255 {
					sc.TimeToLive = 64
				}
			}
		})

	form.AddInputField("SrcPort    :", strconv.Itoa(int(sc.SrcPort)), 5,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 5 && hlp.AcceptNumber(textToCheck, lastChar)
		}, func(text string) {
			hlp.ParseNumberUint16(text, &sc.SrcPort)
		})

	form.AddInputField("DstPort    :", strconv.Itoa(int(sc.DstPort)), 5,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 5 && hlp.AcceptNumber(textToCheck, lastChar)
		}, func(text string) {
			hlp.ParseNumberUint16(text, &sc.DstPort)
		})

	form.AddInputField("VlanID     :", strconv.Itoa(int(sc.VlanId)), 4,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 4 && hlp.AcceptNumber(textToCheck, lastChar)
		}, func(text string) {
			if err := hlp.ParseNumberUint16(text, &sc.VlanId); err == nil {
				if sc.VlanId == 0 {
					sc.VlanId = 1
				} else if sc.VlanId > 4095 {
					sc.VlanId = 4095
				}
			}
		})

	form.AddInputField("DstIP      :", sc.DstIP.String(), 15,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 15 && hlp.AcceptIPv4(textToCheck, lastChar)
		}, func(text string) {
			ip := net.ParseIP(text)
			sc.DstIP.IP = ip
			sc.DstIP.Mask = ip.DefaultMask()
		})

	form.AddInputField("SrcIP      :", sc.SrcIP.String(), 18,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 18 && hlp.AcceptIPv4CiDR(textToCheck, lastChar)
		}, func(text string) {
			ip := net.ParseIP(text)
			sc.SrcIP.IP = ip
			sc.SrcIP.Mask = ip.DefaultMask()
		})

	form.AddInputField("DstMAC     :", sc.DstMAC.String(), 18,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 18 && hlp.AcceptMac(textToCheck, lastChar)
		}, func(text string) {
			mac, err := net.ParseMAC(text)
			if err == nil {
				sc.DstMAC = mac
			}
		})

	form.AddInputField("SrcMAC     :", sc.SrcMAC.String(), 18,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 18 && hlp.AcceptMac(textToCheck, lastChar)
		}, func(text string) {
			mac, err := net.ParseMAC(text)
			if err == nil {
				sc.SrcMAC = mac
			}
		})

	form.AddDropDownSimple("Proto:", 0,
		func(optionIndex int, option *kview.DropDownOption) {
			sc.Proto = option.GetText()
		}, "IPv4/UDP", "IPv4/TCP", "IPv6/UDP", "IPv6/TCP")

	form.AddButton("Save", func() {
		cv.pConfig[port] = &sc
		cv.panels.HidePanel(pg)
		cv.to.SetInputFocus(tabChar)
	})
	form.SetButtonTextColor(tcell.ColorBlack)

	form.AddButton("Cancel", func() {
		cv.panels.HidePanel(pg)
		cv.to.SetInputFocus(tabChar)
	})
	form.SetButtonTextColor(tcell.ColorBlack)

	flex := kview.NewFlex()
	flex.SetDirection(kview.FlexRow)
	flex.AddItem(form, 0, 1, true)

	flex.SetTitleAlign(kview.AlignLeft)
	flex.SetBorder(true)
	flex.SetRect(20, 3, 35, 21)

	cv.formNames[port] = pg

	return flex
}

func (cv *ConfigView) DisplayConfigTable() {

	table := cv.TableView()
	row := 0
	col := 0

	titles := []hlp.TextInfo{
		hlp.NewText(cz.Yellow("Port", 5), kview.AlignLeft),
		hlp.NewText(cz.Yellow("TX Count", 8), kview.AlignLeft),
		hlp.NewText(cz.Yellow("% Rate", 7), kview.AlignLeft),
		hlp.NewText(cz.Yellow("Size", 4), kview.AlignLeft),
		hlp.NewText(cz.Yellow("Burst", 5), kview.AlignLeft),
		hlp.NewText(cz.Yellow("TTL", 4), kview.AlignLeft),
		hlp.NewText(cz.Yellow("sport", 5), kview.AlignLeft),
		hlp.NewText(cz.Yellow("dport", 5), kview.AlignLeft),
		hlp.NewText(cz.Yellow("Proto", 8), kview.AlignLeft),
		hlp.NewText(cz.Yellow("VLAN", 4), kview.AlignLeft),
		hlp.NewText(cz.Yellow("IP Dst"), kview.AlignLeft),
		hlp.NewText(cz.Yellow("IP Src"), kview.AlignLeft),
		hlp.NewText(cz.Yellow("MAC Dst", 14), kview.AlignLeft),
		hlp.NewText(cz.Yellow("MAC Src", 14), kview.AlignLeft),
	}
	row = hlp.TableSetHeaders(table, 0, 0, titles)

	state := func(port int, state bool) string {
		var s string = "   "

		if state {
			s = ">> "
		}
		return fmt.Sprintf("%s%s", cz.DarkMagenta(s), cz.Yellow(port, 2))
	}

	txCount := func(c uint64) string {
		if c == 0 {
			return "Forever"
		}
		p := message.NewPrinter(language.English)
		return p.Sprintf("%v", c)
	}

	for v := uint16(0); v < cv.portCnt; v++ {
		cfg := cv.pConfig[v]

		rowData := []hlp.TextInfo{
			hlp.NewText(state(int(cfg.PortIndex), cfg.TxState), kview.AlignLeft),
			hlp.NewText(cz.CornSilk(txCount(cfg.TxCount)), kview.AlignLeft),
			hlp.NewText(cz.DeepPink(strconv.FormatFloat(cfg.PercentRate, 'f', 2, 64)), kview.AlignLeft),
			hlp.NewText(cz.LightCoral(cfg.PktSize), kview.AlignLeft),
			hlp.NewText(cz.LightCoral(cfg.BurstCount), kview.AlignLeft),
			hlp.NewText(cz.LightCoral(cfg.TimeToLive), kview.AlignLeft),
			hlp.NewText(cz.LightCoral(cfg.SrcPort), kview.AlignLeft),
			hlp.NewText(cz.LightCoral(cfg.DstPort), kview.AlignLeft),
			hlp.NewText(cz.LightBlue(cfg.Proto), kview.AlignLeft),
			hlp.NewText(cz.Cyan(cfg.VlanId), kview.AlignLeft),
			hlp.NewText(cz.CornSilk(cfg.DstIP.IP.String()), kview.AlignLeft),
			hlp.NewText(cz.CornSilk(cfg.SrcIP.String()), kview.AlignLeft),
			hlp.NewText(cz.Green(cfg.DstMAC.String()), kview.AlignLeft),
			hlp.NewText(cz.Green(cfg.SrcMAC.String()), kview.AlignLeft),
		}
		for i, d := range rowData {
			if i == 0 {
				col = hlp.TableCellSelect(table, row, 0, d)
			} else {
				col = hlp.TableCellSet(table, row, col, d)
			}
		}

		row++
	}
	cv.cOnce.Do(func() {
		table.ScrollToBeginning()
	})
}
