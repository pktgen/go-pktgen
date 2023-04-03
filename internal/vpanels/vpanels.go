// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

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

func GetVPanels() *VPanels {
	return vPanels
}

// Initialize is a function that initializes the global VPanels instance.
// It uses a sync.Once to ensure that the initialization is performed only once,
// even if multiple goroutines call initVPanels concurrently.
func Initialize(panels *kview.Panels, app *kview.Application) *VPanels {

	vp := GetVPanels()
	vp.rootPanel = panels
	vp.rootApplication = app

	return vp
}

func (vp *VPanels) Register(panelName string, panelIndex int, panelFunc VPanelFunc) error {

	vp.mutex.Lock()
	defer vp.mutex.Unlock()

	if _, ok := vp.indexMap[panelIndex]; ok {
		return errVPanelAlreadyExists
	} else {
		vp.indexMap[panelIndex] = &vPanelData{
			Index: panelIndex,
			Name:  panelName,
			Func:  panelFunc,
		}
		vp.nameMap[panelName] = panelIndex
		vp.count++
	}

	return nil
}

func (vp *VPanels) Call() error {

	vp.mutex.Lock()
	defer vp.mutex.Unlock()

	for i := 0; i < vp.count; i++ {
		if panel, ok := vp.indexMap[i]; ok {
			cfg := VPanelConfig{
				Name:   panel.Name,
				Panels: vp.rootPanel,
				App:    vp.rootApplication,
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

func (vp *VPanels) Count() int {

	if vPanels == nil {
		return 0
	}

	vp.mutex.Lock()
	defer vp.mutex.Unlock()

	return vp.count
}

func (vp *VPanels) GetInfo(idx int) *VPanelInfo {
	if vPanels == nil {
		return nil
	}

	vp.mutex.Lock()
	defer vp.mutex.Unlock()

	if panel, ok := vp.indexMap[idx]; ok {
		return panel.Info
	} else {
		return nil
	}
}

func (vp *VPanels) GetInfoList() []*VPanelInfo {
	if vPanels == nil {
		return nil
	}

	vp.mutex.Lock()
	defer vp.mutex.Unlock()

	var infoList []*VPanelInfo
	for i := 0; i < vp.count; i++ {
		if panel, ok := vp.indexMap[i]; ok {
			infoList = append(infoList, panel.Info)
		}
	}
	return infoList
}

func (vp *VPanels) IndexByName(name string) int {

	if vPanels == nil {
		return -1
	}

	vp.mutex.Lock()
	defer vp.mutex.Unlock()

	if v, ok := vp.nameMap[name]; !ok {
		return -1
	} else {
		if panel, ok := vp.indexMap[v]; ok {
			return panel.Index
		} else {
			return -1
		}
	}
}

func (vp *VPanels) NameByIndex(idx int) string {

	if vPanels == nil {
		return ""
	}

	vp.mutex.Lock()
	defer vp.mutex.Unlock()

	if v, ok := vp.indexMap[idx]; !ok {
		return ""
	} else {
		return v.Name
	}
}

func (vp *VPanels) GetPanelNames() []string {
	if vPanels == nil {
		return nil
	}
	names := make([]string, 0)
	for i := 0; i < vp.count; i++ {
		names = append(names, vp.NameByIndex(i))
	}
	return names
}
