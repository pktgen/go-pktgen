// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package vpanels

import (
	"sync"

	"github.com/pktgen/go-pktgen/pkgs/kview"
)

// VPanelFunc is a function type that represents a function that creates a new VPanelData instance.
// This function is used to define the behavior of each individual panel in the VPanels system.
//
// The function takes a single parameter:
// - cfg (VPanelConfig): A struct containing configuration information for the panel.
//
// The function returns two values:
// - *VPanelData: A pointer to a VPanelData instance representing the newly created panel.
// - error: An error value that will be nil if the panel creation is successful, or an error if it fails.
type VPanelFunc func(cfg VPanelConfig) (*VPanelInfo, error)

type VPanelInfo struct {
	PanelName string                       // Name of the panel
	HelpID    string                       // Help ID for the panel
	TopFlex   *kview.Flex                  // Top Level Flex View
	TimerFn   func(step int, ticks uint64) // Timer callback function
}

// VPanelConfig is a struct that contains configuration information for a VPanel.
type VPanelConfig struct {
	Name   string             // Name of the panel
	Panels *kview.Panels      // Primary panel
	App    *kview.Application // Application or top level application
}

// vPanelData is a struct that contains the log and help text for a panel.
type vPanelData struct {
	Index int         // Index for the panel in the registerMap
	Name  string      // Name of the panel
	Func  VPanelFunc  // Function to create a new VPanelData instance
	Info  *VPanelInfo // Configuration for the panel
}

// VPanels is a primary structure that represents the VPanels system.
type VPanels struct {
	rootApplication *kview.Application  // Application or top level application
	rootPanel       *kview.Panels       // Primary panel
	count           int                 // Number of panels registered
	indexMap        map[int]*vPanelData // Map of vPanelData by index
	nameMap         map[string]int      // Map of vPanelData by name
	mutex           sync.Mutex          // Lock for concurrent access to the map
}
