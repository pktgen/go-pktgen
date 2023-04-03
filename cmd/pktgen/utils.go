// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package main

import (
	"fmt"

	vp "github.com/pktgen/go-pktgen/internal/vpanels"
)

func buildPanelString(idx int) string {
	// Build the panel selection string at the bottom of the xterm and
	// highlight the selected tab/panel item.
	s := ""
	for index, p := range vp.GetPanelNames() {
		if index == idx {
			s += fmt.Sprintf("F%d:[orange::r]%s[white::-]", index+1, p)
		} else {
			s += fmt.Sprintf("F%d:[orange::-]%s[white::-]", index+1, p)
		}
		if (index + 1) < vp.Count() {
			s += " "
		}
	}
	return s
}

// Version number string
func Version() string {
	return pktgen.version
}

func AddModalPage(title string, modal interface{}) {
	pktgen.ModalPages = append(pktgen.ModalPages, &ModalPage{title: title, modal: modal})
}
