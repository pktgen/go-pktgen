// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022-2024 Intel Corporation

package configview

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"code.rocketnine.space/tslocum/cview"
	"github.com/gdamore/tcell/v2"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	"github.com/pktgen/go-pktgen/internal/constants"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	tab "github.com/pktgen/go-pktgen/internal/taborder"
)

const (
	configFormName string = "configForm"
)

type ConfigView struct {
	portCnt     int
	cTable      *cview.Table
	cOnce       sync.Once
	pConfig     []*constants.PacketConfig
	to          *tab.Tab
	formNames   []string
	configForms []*cview.Flex
	pages       *cview.Pages
}

func Create(pages *cview.Pages, to *tab.Tab, portCnt int, flex *cview.Flex) *ConfigView {
	configView := &ConfigView{portCnt: portCnt, pages: pages, to: to}

	configView.cTable = hlp.CreateTableView(flex, "Configuration (c) s-Toggle Start/Stop, a=StartAll, A=StopAll, e-Edit",
		cview.AlignLeft, 0, 1, true)
	configView.cTable.SetSelectable(true, false)
	configView.cTable.SetFixed(1, 1)
	configView.cTable.SetSelectionChangedFunc(func(row, col int) {
		// Adjust for not selecting the first column in a row. table.go will select the next row.
		// on a mouse click
		if col > 0 {
			configView.cTable.Select(row, 0)
		}
	})
	configView.cTable.SetSeparator(cview.Borders.Vertical)

	configView.initConfigView()

	return configView
}

func (cv *ConfigView) TableView() *cview.Table {
	return cv.cTable
}

func (cv *ConfigView) FormName(port int) string {
	return cv.formNames[port]
}

func (cv *ConfigView) ConfigForm(port int) *cview.Flex {
	return cv.configForms[port]
}

func (cv *ConfigView) PacketConfigByPort(port int) *constants.PacketConfig {
	return cv.pConfig[port]
}

func (cv *ConfigView) TxState(port int) bool {
	return cv.pConfig[port].TxState
}

func (cv *ConfigView) SetTxState(port int, val bool) {
	cv.pConfig[port].TxState = val
}

func (cv *ConfigView) initConfigView() {

	cv.pConfig = make([]*constants.PacketConfig, 8)
	cv.formNames = make([]string, 8)

	for i := 0; i < cv.portCnt; i++ {
		cv.pConfig[i] = &constants.PacketConfig{
			PortIndex:   i,
			TxCount:     0,
			PercentRate: 100.0,
			PktSize:     64,
			BurstCount:  128,
			TimeToLive:  64,
			SrcPort:     1245,
			DstPort:     5678,
			PType:       "IPv4",
			ProtoType:   "UDP",
			VlanId:      1,
			DstIP:       net.IPNet{IP: net.IPv4(198, 18, 1, 1), Mask: net.CIDRMask(0, 32)},
			SrcIP:       net.IPNet{IP: net.IPv4(198, 18, 0, 1), Mask: net.CIDRMask(24, 32)},
			DstMAC:      []byte{0x12, 0x34, 0x45, 0x67, 0x89, 00},
			SrcMAC:      []byte{0x12, 0x34, 0x45, 0x67, 0x89, 01},
			TxState:     false,
		}
	}
	for port := 0; port < cv.portCnt; port++ {
		f := cv.setupEditConfigForm(port)
		cv.configForms = append(cv.configForms, f)
	}
}

func (cv *ConfigView) setupEditConfigForm(port int) *cview.Flex {

	pg := fmt.Sprintf("%v-%v", configFormName, port)

	form := cview.NewForm()
	form.SetItemPadding(1)
	form.SetHorizontal(false)
	form.SetFieldTextColor(tcell.ColorBlack)
	form.SetFieldBackgroundColor(tcell.ColorBlue)
	form.SetItemPadding(0)
	form.SetCancelFunc(func() {
		cv.pages.HidePage(pg)
		cv.to.SetInputFocus('c')
	})

	form.SetTitleAlign(cview.AlignLeft)
	form.SetRect(0, 0, 35, 21)

	sc := *cv.pConfig[port]

	form.AddInputField("TxCount  :", strconv.Itoa(int(sc.TxCount)), 15,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 15 && hlp.AcceptNumber(textToCheck, lastChar)
		}, func(text string) {
			hlp.ParseNumberUint64(text, &sc.TxCount)
		})

	form.AddInputField("Rate     :", strconv.FormatFloat(sc.PercentRate, 'f', 2, 64), 6,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 6 && hlp.AcceptFloat(textToCheck, lastChar)
		}, func(text string) {
			if err := hlp.ParseNumberFloat64(text, &sc.PercentRate); err == nil {
				if sc.PercentRate == 0 || sc.PercentRate > 100.00 {
					sc.PercentRate = 100.00
				}
			}
		})

	form.AddInputField("PktSize  :", strconv.Itoa(int(sc.PktSize)), 5,
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

	form.AddInputField("Burst    :", strconv.Itoa(int(sc.BurstCount)), 3,
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

	form.AddInputField("TTL      :", strconv.Itoa(int(sc.TimeToLive)), 3,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 3 && hlp.AcceptNumber(textToCheck, lastChar)
		}, func(text string) {
			if err := hlp.ParseNumberUint16(text, &sc.TimeToLive); err == nil {
				if sc.TimeToLive > 255 {
					sc.TimeToLive = 64
				}
			}
		})

	form.AddInputField("SrcPort  :", strconv.Itoa(int(sc.SrcPort)), 5,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 5 && hlp.AcceptNumber(textToCheck, lastChar)
		}, func(text string) {
			hlp.ParseNumberUint16(text, &sc.SrcPort)
		})

	form.AddInputField("DstPort  :", strconv.Itoa(int(sc.DstPort)), 5,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 5 && hlp.AcceptNumber(textToCheck, lastChar)
		}, func(text string) {
			hlp.ParseNumberUint16(text, &sc.DstPort)
		})

	form.AddDropDownSimple("PType    :", 0,
		func(optionIndex int, option *cview.DropDownOption) {
			sc.PType = option.GetText()
		}, "IPv4", "IPv6", "ICMP")

	form.AddDropDownSimple("Protocol :", 0,
		func(optionIndex int, option *cview.DropDownOption) {
			sc.ProtoType = option.GetText()
		}, "UDP", "TCP")

	form.AddInputField("VlanID   :", strconv.Itoa(int(sc.VlanId)), 4,
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

	form.AddInputField("DstIP    :", sc.DstIP.String(), 15,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 15 && hlp.AcceptIPv4(textToCheck, lastChar)
		}, func(text string) {
			ip := net.ParseIP(text)
			sc.DstIP.IP = ip
			sc.DstIP.Mask = ip.DefaultMask()
		})

	form.AddInputField("SrcIP    :", sc.SrcIP.String(), 18,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 18 && hlp.AcceptIPv4CiDR(textToCheck, lastChar)
		}, func(text string) {
			ip := net.ParseIP(text)
			sc.SrcIP.IP = ip
			sc.SrcIP.Mask = ip.DefaultMask()
		})

	form.AddInputField("DstMAC   :", sc.DstMAC.String(), 18,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 18 && hlp.AcceptMac(textToCheck, lastChar)
		}, func(text string) {
			mac, err := net.ParseMAC(text)
			if err == nil {
				sc.DstMAC = mac
			}
		})

	form.AddInputField("SrcMAC   :", sc.SrcMAC.String(), 18,
		func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= 18 && hlp.AcceptMac(textToCheck, lastChar)
		}, func(text string) {
			mac, err := net.ParseMAC(text)
			if err == nil {
				sc.SrcMAC = mac
			}
		})

	form.AddButton("Save", func() {
		cv.pConfig[port] = &sc
		cv.pages.HidePage(pg)
		cv.to.SetInputFocus('c')
	})
	form.SetButtonTextColor(tcell.ColorBlack)

	form.AddButton("Cancel", func() {
		cv.pages.HidePage(pg)
		cv.to.SetInputFocus('c')
	})
	form.SetButtonTextColor(tcell.ColorBlack)

	flex := cview.NewFlex()
	flex.SetDirection(cview.FlexRow)
	flex.AddItem(form, 0, 1, true)

	flex.SetTitle(hlp.TitleColor("Edit Port"))
	flex.SetTitleAlign(cview.AlignLeft)
	flex.SetBorder(true)
	flex.SetRect(20, 3, 35, 21)

	cv.formNames[port] = pg

	return flex
}

func (cv *ConfigView) DisplayConfigTable() {

	table := cv.TableView()
	row := 0
	col := 0

	titles := []string{
		cz.Yellow("Port", 5),
		cz.Yellow("TX Count", 8),
		cz.Yellow("% Rate", 7),
		cz.Yellow("Size", 4),
		cz.Yellow("Burst", 5),
		cz.Yellow("TTL", 4),
		cz.Yellow("sport", 5),
		cz.Yellow("dport", 5),
		cz.Yellow("PType", 5),
		cz.Yellow("Proto", 5),
		cz.Yellow("VLAN", 4),
		cz.Yellow("IP Dst"),
		cz.Yellow("IP Src"),
		cz.Yellow("MAC Dst", 14),
		cz.Yellow("MAC Src", 14),
	}
	row = hlp.TableSetHeaders(table, 0, 0, titles)

	state := func(port int, state bool) string {
		var s string = "   "

		if state {
			s = ">> "
		}
		return fmt.Sprintf("%s%s", cz.DeepPink(s), cz.Yellow(port, 2))
	}

	txCount := func(c uint64) string {
		if c == 0 {
			return "Forever"
		}
		p := message.NewPrinter(language.English)
		return p.Sprintf("%v", c)
	}

	for v := 0; v < cv.portCnt; v++ {
		cfg := cv.pConfig[v]

		rowData := []string{
			state(cfg.PortIndex, cfg.TxState),
			cz.CornSilk(txCount(cfg.TxCount)),
			cz.DeepPink(strconv.FormatFloat(cfg.PercentRate, 'f', 2, 64)),
			cz.LightCoral(cfg.PktSize),
			cz.LightCoral(cfg.BurstCount),
			cz.LightCoral(cfg.TimeToLive),
			cz.LightCoral(cfg.SrcPort),
			cz.LightCoral(cfg.DstPort),
			cz.LightBlue(cfg.PType),
			cz.LightBlue(cfg.ProtoType),
			cz.Cyan(cfg.VlanId),
			cz.CornSilk(cfg.DstIP.IP.String()),
			cz.CornSilk(cfg.SrcIP.String()),
			cz.Green(cfg.DstMAC.String()),
			cz.Green(cfg.SrcMAC.String()),
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
