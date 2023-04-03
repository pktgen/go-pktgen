// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package helpers

import (
	"fmt"

	"code.rocketnine.space/tslocum/cview"
	"github.com/gdamore/tcell/v2"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
)

// TitleColor - Set the title color to the windows
func TitleColor(msg string) string {

	return fmt.Sprintf("[%s]", cz.Orange(msg))
}

// Center returns a new primitive which shows the provided primitive in its
// center, given the provided primitive's size.
func Center(width, height int, p cview.Primitive) *cview.Flex {
	/*
		return cview.NewFlex().
			AddItem(cview.NewBox(), 0, 1, false).
			AddItem(cview.NewFlex().
				SetDirection(cview.FlexRow).
				AddItem(cview.NewBox(), 0, 1, false).
				AddItem(p, height, 1, true).
				AddItem(cview.NewBox(), 0, 1, false), width, 1, true).
			AddItem(cview.NewBox(), 0, 1, false)
	*/
	f := cview.NewFlex()
	f.AddItem(cview.NewBox(), 0, 1, false)
	f1 := cview.NewFlex()
	f1.SetDirection(cview.FlexRow)
	f1.AddItem(cview.NewBox(), 0, 1, false)
	f1.AddItem(p, height, 1, true)
	f1.AddItem(cview.NewBox(), 0, 1, false)
	f.AddItem(f1, width, 1, true)
	return f
}

// TitleBox to return the top title window
func TitleBox(flex *cview.Flex, text string) *cview.TextView {

	textView := cview.NewTextView()
	textView.SetDynamicColors(true)

	textView.SetText(text)
	textView.SetTextAlign(cview.AlignCenter)

	flex.AddItem(textView, 1, 1, false)

	return textView
}

func setTableCell(table *cview.Table, row, col int, value string, sel bool) int {

	tableCell := cview.NewTableCell(value)
	tableCell.SetAlign(cview.AlignRight)
	tableCell.SetSelectable(sel)
	table.SetCell(row, col, tableCell)
	col++

	return col
}

func TableCellSet(table *cview.Table, row, col int, value string) int {

	return setTableCell(table, row, col, value, false)
}

func TableCellSelect(table *cview.Table, row, col int, value string) int {

	return setTableCell(table, row, col, value, true)
}

func TableSetHeaders(table *cview.Table, row, col int, titles []string) int {

	for _, v := range titles {
		col = TableCellSet(table, row, col, v)
	}
	row++

	return row
}

func TableSetRows(table *cview.Table, row, col int, titles []string) int {

	for _, v := range titles {
		TableCellSet(table, row, col, v)
		row++
	}

	return row
}

// CreateTextView - helper routine to create a TextView
func CreateTextView(flex *cview.Flex, msg string, align, fixedSize, proportion int, focus bool) *cview.TextView {

	textView := cview.NewTextView()
	textView.SetDynamicColors(true)
	textView.SetWrap(true)

	if len(msg) > 0 {
		textView.SetBorder(true)
		textView.SetTitle(TitleColor(msg))
		textView.SetTitleAlign(align)
	}
	flex.AddItem(textView, fixedSize, proportion, focus)

	return textView
}

// CreateTableView - Helper to create a Table
func CreateTableView(flex *cview.Flex, msg string, align, fixedSize, proportion int, focus bool) *cview.Table {
	table := cview.NewTable()
	table.SetFixed(1, 0)

	if len(msg) > 0 {
		table.SetBorder(true)
		table.SetTitle(TitleColor(msg))
		table.SetTitleAlign(align)
	}
	flex.AddItem(table, fixedSize, proportion, focus)

	return table
}

// CreateForm window
func CreateForm(flex *cview.Flex, msg string, align, fixedSize, proportion int, focus bool) *cview.Form {

	form := cview.NewForm()

	if len(msg) > 0 {
		form.SetBorder(true)
		form.SetTitleAlign(align)
		form.SetTitle(TitleColor(msg))
	}

	form.SetFieldBackgroundColor(tcell.ColorReset)
	form.SetFieldTextColor(tcell.ColorReset)

	flex.AddItem(form, fixedSize, proportion, focus)

	return form
}

// CreateList window
func CreateList(flex *cview.Flex, msg string, align, fixedSize, proportion int, focus bool) *cview.List {

	list := cview.NewList()

	if len(msg) > 0 {
		list.SetBorder(true)
		list.SetTitleAlign(align)
		list.SetTitle(TitleColor(msg))
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
//	object a is int then alignment cview.AlignLeft/Right/Center
//	object a is bool then set the cell as selectable or not
func SetCell(table *cview.Table, row, col int, msg string, a ...interface{}) *cview.TableCell {

	align := cview.AlignRight
	selectable := false
	for _, v := range a {
		switch d := v.(type) {
		case int:
			align = d
		case bool:
			selectable = d
		}
	}
	tableCell := cview.NewTableCell(msg)
	tableCell.SetAlign(align)
	tableCell.SetSelectable(selectable)
	table.SetCell(row, col, tableCell)

	return tableCell
}

func CreateModal(p cview.Primitive, width, height int) cview.Primitive {
	g := cview.NewGrid()
	g.SetColumns(0, width, 0)
	g.SetRows(0, height, 0)
	g.AddItem(p, 1, 1, 1, 1, 0, 0, true)

	return g
}
