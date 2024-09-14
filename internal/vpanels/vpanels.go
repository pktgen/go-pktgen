// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package vpanels

import (
	"code.rocketnine.space/tslocum/cview"
)

type VPanelMap map[string]string

type VPanelConfig struct {
	Panel   *cview.Panels
	App     *cview.Application
	OptFunc OptionFunc
}

type VPanelFunc func(pageInfo VPanelConfig) (*VPanelData, error)

type VPanelInfo struct {
	Func   VPanelFunc
	Config VPanelConfig
}

type OptionFunc func() interface{}

type VPanelData struct {
	PanelName string                       // Name of the panel
	HelpName  string                       // Name of the help panel
	TopFlex   *cview.Flex                  // Top Level Flex View
	TimerFn   func(step int, ticks uint64) // Timer callback function
}

var vpanels []VPanelInfo // A list of registered panels a SINGLETON pattern.

func Register(vpf VPanelFunc) {
	panel := VPanelInfo{
		Func: vpf,
	}
	vpanels = append(vpanels, panel)
}

func GetPanels() []VPanelInfo {
	return vpanels
}
