// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package vpanels

import (
	"fmt"
	"sync"

	"github.com/pktgen/go-pktgen/pkgs/kview"
)

var (
	vPanels                *VPanels // A global VPanels instance or singleton instance
	errVPanelAlreadyExists = fmt.Errorf("vPanel already exists")
	errVPanelInitFailed    = fmt.Errorf("vPanel setup call failed")
)

func init() {
	vPanels = &VPanels{
		count:    0,
		indexMap: make(map[int]*vPanelData),
		nameMap:  make(map[string]int),
		mutex:    sync.Mutex{},
	}
}

// Initialize is a function that initializes the global VPanels instance.
// It uses a sync.Once to ensure that the initialization is performed only once,
// even if multiple goroutines call initVPanels concurrently.
func Initialize(panels *kview.Panels, app *kview.Application) *VPanels {

	if vPanels != nil { // Their can be only one.
		vPanels.rootPanel = panels
		vPanels.rootApplication = app
	}
	return vPanels
}

func Register(panelName string, panelIndex int, panelFunc VPanelFunc) error {

	if vPanels == nil {
		return errVPanelInitFailed
	}

	vPanels.mutex.Lock()
	defer vPanels.mutex.Unlock()

	if _, ok := vPanels.indexMap[panelIndex]; ok {
		return errVPanelAlreadyExists
	} else {
		vPanels.indexMap[panelIndex] = &vPanelData{
			Index: panelIndex,
			Name:  panelName,
			Func:  panelFunc,
		}
		vPanels.nameMap[panelName] = panelIndex
		vPanels.count++
	}

	return nil
}

func Call() error {

	if vPanels == nil {
		return errVPanelInitFailed
	}

	vPanels.mutex.Lock()
	defer vPanels.mutex.Unlock()

	for i := 0; i < vPanels.count; i++ {
		if panel, ok := vPanels.indexMap[i]; ok {
			cfg := VPanelConfig{
				Name:  panel.Name,
				Panel: vPanels.rootPanel,
				App:   vPanels.rootApplication,
			}
			if info, err := panel.Func(cfg); err != nil {
				return err
			} else {
				panel.Info = info
			}
		}
	}
	return nil
}

func Count() int {

	if vPanels == nil {
		return 0
	}

	vPanels.mutex.Lock()
	defer vPanels.mutex.Unlock()

	return vPanels.count
}

func GetInfo(idx int) *VPanelInfo {
	if vPanels == nil {
		return nil
	}

	vPanels.mutex.Lock()
	defer vPanels.mutex.Unlock()

	if panel, ok := vPanels.indexMap[idx]; ok {
		return panel.Info
	} else {
		return nil
	}
}

func IndexByName(name string) int {

	if vPanels == nil {
		return -1
	}

	vPanels.mutex.Lock()
	defer vPanels.mutex.Unlock()

	if v, ok := vPanels.nameMap[name]; !ok {
		return -1
	} else {
		if panel, ok := vPanels.indexMap[v]; ok {
			return panel.Index
		} else {
			return -1
		}
	}
}

func NameByIndex(idx int) string {

	if vPanels == nil {
		return ""
	}

	vPanels.mutex.Lock()
	defer vPanels.mutex.Unlock()

	if v, ok := vPanels.indexMap[idx]; !ok {
		return ""
	} else {
		return v.Name
	}
}

func GetPanelNames() []string {
	if vPanels == nil {
		return nil
	}
	names := make([]string, 0)
	for i := 0; i < vPanels.count; i++ {
		names = append(names, NameByIndex(i))
	}
	return names
}
