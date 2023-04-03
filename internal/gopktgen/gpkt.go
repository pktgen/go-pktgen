// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package gopktgen

import (
	"fmt"
	"os"
	"strings"
	"time"

	gpc "github.com/pktgen/go-pktgen/internal/gpcommon"
	"github.com/pktgen/go-pktgen/internal/tlog"
)

func createGoPktgen(pCfg *PktgenConfig) *GoPktgen {

	// Create a new GoPktgen instance with default values
	gPkt := &GoPktgen{
		GoPktgenApi: GoPktgenApi{},
		pCfg: pCfg,
		l2p: &L2p{
			Cores:       make(map[gpc.CoreID]*LogicalCore),
			LPorts:      make(map[gpc.LPortID]*LogicalPort),
			Ports:       make(map[gpc.PortID]*PhysicalPort),
			BaseMapList: []*BaseMapping{},
		},
		libList:   []*LibInfo{},
		portData:  []*gpc.PortData{},
		portStats: []*PortStats{},
		basePath:  gpc.DefaultLibraryPath,
		logPath:   "",
	}

	return gPkt
}

// Create the GoPktgen instance and load the necessary libraries
func New(pCfg *PktgenConfig, options ...GoPktgenOption) (*GoPktgen, error) {

	if pCfg == nil {
        return nil, fmt.Errorf("PktgenConfig pointer is required")
    }
	gPkt := createGoPktgen(pCfg)

	// Process the option function calls
	for _, option := range options {
		option(gPkt)
	}

	if err := gPkt.openLibs(); err != nil {
		return nil, fmt.Errorf("error loading Go-Pktgen Libraries or APIs: %s", err)
	}
	if err := gPkt.loadAPIs(); err != nil {
		return nil, fmt.Errorf("error loading Go-Pktgen APIs: %s", err)
	}

	return gPkt, nil
}

func (g *GoPktgen) Destroy() {

	g.closeLibs()
	g.ChannelDestroy(g.dpdkChan)
}

func (g *GoPktgen) attachMsgChan(name string) uintptr {

	ch := make(chan uintptr)

	go func() {
	Loop:
		for {
			// Attach to the EAL message channel
			if mc := g.ChannelAttach(name); mc == uintptr(0) {
				tlog.Printf("Failed to attach to [%s] message channel, sleeping\n", name)
				time.Sleep(time.Second / 2)
			} else {
				tlog.Printf("Attached to MsgChan [%s]\n", name)
				ch <- mc
				break Loop
			}
		}
	}()
	return <-ch
}

// Start initializes and starts the GoPktgen system.
// It performs the following steps:
//  1. Calls the scrollScreen function to clear the terminal screen.
//  2. Initializes the DPDK system using the provided log path.
//  3. Attaches to the EAL message channel with the default channel name.
//     If the attachment fails, it sleeps for half a second and tries again.
//  4. Initializes the physical port information.
//  5. Creates the threads using the attached message channel.
//  6. Returns an error if any step fails.
func (g *GoPktgen) Start() error {

	scrollScreen(128)

	tlog.Printf("Go-Pktgen started log path (%s)\n", g.logPath)

	// Initialize the DPDK subsystem
	if ret := g.Startup(g.logPath); ret < 0 {
		return fmt.Errorf("failed to initialize DPDK")
	}

	g.dpdkChan = g.attachMsgChan(gpc.ChannelDPDKName)

	// Initialize the physical port information
	for i := uint16(0); i < g.PortCount(); i++ {
		numRx, numTx := g.NumQueues(gpc.PortID(i))

		tlog.Printf("Port %d: Number of Rx/Tx Queues: %d/%d\n", i, numRx, numTx)

		cfg := gpc.NewPortConfig(gpc.PortID(i), numRx, numTx)

		if g.PortSetInfo(cfg) < 0 {
			return fmt.Errorf("failed to set port information %d", i)
		}

		g.portStats = append(g.portStats, g.NewPortStats())
	}

	// Create the threads
	if err := g.LaunchThreads(); err != nil {
		tlog.Printf("Launching threads failed: %v", err)
		return err
	}

	return nil
}

// Stop is the function to stop DPDK
func (g *GoPktgen) Stop() {
	g.Shutdown()
}

func WithBasePath(basePath string) GoPktgenOption {
	return func(g *GoPktgen) {
		g.basePath = basePath
	}
}

func WithLibNames(libNames []string) GoPktgenOption {
	return func(g *GoPktgen) {
		for _, name := range libNames {
			libPath := name
			if name[0] != '/' && name[0] != '.' { // Absolute or relative path
				if g.basePath != "" {
					libPath = g.basePath + "/" + name // Prepend the base path
				}
			}
			g.libList = append(g.libList, &LibInfo{Name: libPath, Handle: 0})
		}
	}
}

func WithLogPath(ttyID int) GoPktgenOption {
	return func(g *GoPktgen) {
		g.logPath = fmt.Sprintf("/dev/pts/%d", ttyID)
	}
}

func (g *GoPktgen) AddPorts(pci ...string) {

	for pid, id := range pci {
		g.portData = append(g.portData, &gpc.PortData{
			PCIeID:     id,
			MacAddress: gpc.MacAddress{},
			PortID:     gpc.PortID(pid),
		})
	}
}

func scrollScreen(lines int) {

	// Clear the screen by scrolling the terminal to the top left corner
	// which allows for output text to be saved from the clear screen action
	fmt.Fprintf(os.Stderr, "\r" + strings.Repeat("\n", lines))
}

func (g *GoPktgen) SetL2pConfig(cfg *L2pConfig) int {
	return g.SetL2pConfig(cfg)
}
