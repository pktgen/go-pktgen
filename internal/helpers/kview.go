// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package helpers

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/pktgen/go-pktgen/pkgs/kview"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
	tab "github.com/pktgen/go-pktgen/internal/taborder"
)

type ModalPage struct {
	title string
	modal interface{}
}

var modalPages = make([]*ModalPage, 0)

type TextInfo struct {
	Text  string // Text can include colorized text
	Align int    // Alignment cview.AlignLeft, cview.AlignCenter, cview.AlignRight
}

func NewText(text string, align int) TextInfo {
	return TextInfo{Text: text, Align: align}
}

// TitleColor - Set the title color to the windows
func TitleColor(msg string) string {

	return fmt.Sprintf("[%s]", cz.Orange(msg))
}

// Center returns a new primitive which shows the provided primitive in its
// center, given the provided primitive's size.
func Center(width, height int, p kview.Primitive) *kview.Flex {

	f := kview.NewFlex()
	f.AddItem(kview.NewBox(), 0, 1, false)
	f1 := kview.NewFlex()
	f1.SetDirection(kview.FlexRow)
	f1.AddItem(kview.NewBox(), 0, 1, false)
	f1.AddItem(p, height, 1, true)
	f1.AddItem(kview.NewBox(), 0, 1, false)
	f.AddItem(f1, width, 1, true)
	return f
}

// TitleBox to return the top title window
func TitleBox(flex *kview.Flex, text string) *kview.TextView {

	textView := kview.NewTextView()
	textView.SetDynamicColors(true)

	textView.SetText(text)
	textView.SetTextAlign(kview.AlignCenter)

	flex.AddItem(textView, 1, 1, false)

	return textView
}

func setTableCell(table *kview.Table, row, col int, v TextInfo, sel bool) int {

	tableCell := kview.NewTableCell(v.Text)
	tableCell.SetAlign(v.Align)
	tableCell.SetSelectable(sel)
	table.SetCell(row, col, tableCell)
	col++

	return col
}

func TableCellSet(table *kview.Table, row, col int, v TextInfo) int {

	return setTableCell(table, row, col, v, false)
}

func TableCellSelect(table *kview.Table, row, col int, v TextInfo) int {

	return setTableCell(table, row, col, v, true)
}

func TableSetHeaders(table *kview.Table, row, col int, titles []TextInfo) int {

	for _, v := range titles {
		col = TableCellSet(table, row, col, v)
	}
	row++

	return row
}

func TableSetRows(table *kview.Table, row, col int, titles []TextInfo) int {

	for _, v := range titles {
		TableCellSet(table, row, col, v)
		row++
	}

	return row
}

// CreateTextView - helper routine to create a TextView
func CreateTextView(flex *kview.Flex, msg TextInfo, fixedSize, proportion int, focus bool) *kview.TextView {

	textView := kview.NewTextView()
	textView.SetDynamicColors(true)
	textView.SetWrap(true)

	if len(msg.Text) > 0 {
		textView.SetBorder(true)
		textView.SetTitle(TitleColor(msg.Text))
		textView.SetTitleAlign(msg.Align)
	}
	flex.AddItem(textView, fixedSize, proportion, focus)

	return textView
}

// CreateTableView - Helper to create a Table
func CreateTableView(flex *kview.Flex, msg TextInfo, fixedSize, proportion int, focus bool) *kview.Table {
	table := kview.NewTable()
	table.SetFixed(1, 0)

	if len(msg.Text) > 0 {
		table.SetBorder(true)
		table.SetTitle(TitleColor(msg.Text))
		table.SetTitleAlign(msg.Align)
	}
	flex.AddItem(table, fixedSize, proportion, focus)

	return table
}

// CreateForm window
func CreateForm(flex *kview.Flex, msg TextInfo, fixedSize, proportion int, focus bool) *kview.Form {

	form := kview.NewForm()

	if len(msg.Text) > 0 {
		form.SetBorder(true)
		form.SetTitleAlign(msg.Align)
		form.SetTitle(TitleColor(msg.Text))
	}

	form.SetFieldBackgroundColor(tcell.ColorReset)
	form.SetFieldTextColor(tcell.ColorReset)

	flex.AddItem(form, fixedSize, proportion, focus)

	return form
}

// CreateList window
func CreateList(flex *kview.Flex, msg TextInfo, fixedSize, proportion int, focus bool) *kview.List {

	list := kview.NewList()

	if len(msg.Text) > 0 {
		list.SetBorder(true)
		list.SetTitleAlign(msg.Align)
		list.SetTitle(TitleColor(msg.Text))
	}
	list.ShowSecondaryText(false)

	flex.AddItem(list, fixedSize, proportion, focus)

	return list
}

// SetCell content given the information
// row, col of the cell to create and fill
// msg is the string content to insert in the cell
// a is an interface{} object list
//
//	object a is int then alignment kview.AlignLeft/Right/Center
//	object a is bool then set the cell as selectable or not
func SetCell(table *kview.Table, row, col int, msg TextInfo, selectable bool) *kview.TableCell {

	tableCell := kview.NewTableCell(msg.Text)
	tableCell.SetAlign(msg.Align)
	tableCell.SetSelectable(selectable)
	table.SetCell(row, col, tableCell)

	return tableCell
}

func CreateTabOrder(app *kview.Application, panelName string, data []tab.TabData) (*tab.Tab, error) {

	to := tab.New(panelName, app)

	// Setup the tab order for each view
	for _, v := range data {
		if err := to.Add(v.Name, v.View, v.Key); err != nil {
			return nil, err
		}
	}

	if err := to.SetInputDone(); err != nil {
		return nil, err
	}

	return to, nil
}

func CreateModal(p kview.Primitive, width, height int) kview.Primitive {
	g := kview.NewGrid()
	g.SetColumns(0, width, 0)
	g.SetRows(0, height, 0)
	g.AddItem(p, 1, 1, 1, 1, 0, 0, true)

	return g
}

func AddModalPage(title string, modal interface{}) {
	modalPages = append(modalPages, &ModalPage{title: title, modal: modal})
}

func GetModalPages() []*ModalPage {
	return modalPages
}
