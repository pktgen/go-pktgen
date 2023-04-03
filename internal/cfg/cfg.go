// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022-2024 Intel Corporation

package cfg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/tidwall/jsonc"
)

const (
	IdxCoreMask = iota
	IdxCoreList
	IdxCoreMap
	IdxServiceCoreMask
	IdxMainLCore
	IdxMBufPoolOpsName
	IdxNumChannels
	IdxMemorySize
	IdxNumRanks
	IdxBlockList
	IdxAllowList
	IdxVirtualDevices
	IdxIovaMode
	IdxDrivers
	IdxVmWareTSCMap
	IdxProcType
	IdxSysLog
	IdxLogLevel
	IdxTrace
	IdxTraceDir
	IdxTraceBufferSize
	IdxTraceMode
	IdxVersion
	IdxHelp
	IdxInMemory
	IdxBaseVirtAddr
	IdxNoTelemetry
	IdxForceMaxSimdBitwidth
	IdxHugeUnlink
	IdxNoHugePages
	IdxNoPCI
	IdxNoHPET
	IdxNoShConf
	IdxSocketMemory
	IdxSocketLimit
	IdxHugeDir
	IdxFilePrefix
	IdxCreateUIODev
	IdxVfioIntr
	IdxVfioVFToken
	IdxLegacyMemory
	IdxSingleFileSegments
	IdxMatchAllocations
	IdxHugeWorkerStack
	IdxUserArguments
)

// Make sure the order of the constants above is the same as the order of the
// of the structure below.
type configData struct {
	CoreMask             *string   `json:"core-mask"`
	CoreList             *string   `json:"core-list"`
	CoreMap              *string   `json:"core-map"`
	ServiceCoreMask      *string   `json:"service-core-mask"`
	MainLCore            int       `json:"main-lcore"`
	MBufPoolOpsName      *string   `json:"mbuf-pool-ops-name"`
	NumChannels          int       `json:"num-channels"`
	MemorySize           uint64    `json:"memory-size"`
	NumRanks             int       `json:"num-ranks"`
	BlockList            []*string `json:"block-list"`
	AllowList            []*string `json:"allow-list"`
	VirtualDevices       []*string `json:"virtual-device"`
	IovaMode             *string   `json:"iova-mode"`
	Drivers              []*string `json:"drivers"`
	VmWareTSCMap         *string   `json:"vmware-tsc-map"`
	ProcType             *string   `json:"proc-type"`
	SysLog               *string   `json:"syslog"`
	LogLevel             *string   `json:"log-level"`
	Trace                *string   `json:"trace"`
	TraceDir             *string   `json:"trace-dir"`
	TraceBufferSize      int       `json:"trace-bufsz"`
	TraceMode            *string   `json:"trace-mode"`
	Version              bool      `json:"version"`
	Help                 bool      `json:"help"`
	InMemory             bool      `json:"in-memory"`
	BaseVirtAddr         *string   `json:"base-virtaddr"`
	NoTelemetry          bool      `json:"no-telemetry"`
	ForceMaxSimdBitwidth int       `json:"force-max-simd-bitwidth"`
	HugeUnlink           *string   `json:"huge-unlink"`
	NoHugePages          bool      `json:"no-huge-pages"`
	NoPCI                bool      `json:"no-pci"`
	NoHPET               bool      `json:"no-hpet"`
	NoShConf             bool      `json:"no-shconf"`
	SocketMemory         []*string `json:"socket-mem"`
	SocketLimit          []*string `json:"socket-limit"`
	HugeDir              *string   `json:"huge-dir"`
	FilePrefix           *string   `json:"file-prefix"`
	CreateUIODev         bool      `json:"create-uio-dev"`
	VfioIntr             *string   `json:"vfio-intr"`
	VfioVFToken          *string   `json:"vfio-vf-token"`
	LegacyMemory         *string   `json:"legacy-mem"`
	SingleFileSegments   bool      `json:"single-file-segments"`
	MatchAllocations     bool      `json:"match-allocations"`
	HugeWorkerStack      int       `json:"huge-worker-stack"`
}

type System struct {
	cBytes []byte
	cd     configData
}

var options = []string{
	"-c",                        // CoreMask
	"-l",                        // CoreList
	"-lcores",                   // CoreMap
	"-s",                        // ServiceCoreMask
	"--main-lcore",              // MainLCore
	"--mbuf-pool-ops-name",      // MBufPoolOpsName
	"-n",                        // NumChannels
	"-m",                        // MemorySize
	"-r",                        // NumRanks
	"-b",                        // BlockList
	"-a",                        // AllowList
	"--vdev",                    // VirtualDevices
	"--iova-mode",               // IovaMode
	"-d",                        // Drivers
	"--vmware-tsc-map",          // VmWareTSCMap
	"--proc-type",               // ProcType
	"--syslog",                  // SysLog
	"--log-level",               // LogLevel
	"--trace",                   // Trace
	"--trace-dir",               // TraceDir
	"--trace-bufsz",             // TraceBufferSize
	"--trace-mode",              // TraceMode
	"-v",                        // Version
	"--help",                    // Help
	"--in-memory",               // InMemory
	"--base-virtaddr",           // BaseVirtAddr
	"--no-telemetry",            // NoTelemetry
	"--force-max-simd-bitwidth", // ForceMaxSimdBitwidth
	"--huge-unlink",             // HugeUnlink
	"--no-huge-pages",           // NoHugePages
	"--no-pci",                  // NoPCI
	"--no-hpet",                 // NoHPET
	"--no-shconf",               // NoShConf
	"--socket-mem",              // SocketMemory
	"--socket-limit",            // SocketLimit
	"--huge-dir",                // HugeDir
	"--file-prefix",             // FilePrefix
	"--create-uio-dev",          // CreateUIODev
	"--vfio-intr",               // VfioIntr
	"--vfio-vf-token",           // VfioVFToken
	"--legacy-mem",              // LegacyMemory
	"--single-file-segments",    // SingleFileSegments
	"--match-allocations",       // MatchAllocations
	"--huge-worker-stack",       // HugeWorkerStack
	"--",                        // User arguments
}

func New() *System {

	return &System{
		cBytes: []byte("{}"),
		cd: configData{
			CoreMask:             nil,
			CoreList:             nil,
			CoreMap:              nil,
			ServiceCoreMask:      nil,
			MainLCore:            -1,
			MBufPoolOpsName:      nil,
			NumChannels:          0,
			MemorySize:           0, // Size of memory is in MBytes.
			NumRanks:             0,
			BlockList:            nil,
			AllowList:            nil,
			VirtualDevices:       nil,
			IovaMode:             nil,
			Drivers:              nil,
			VmWareTSCMap:         nil,
			ProcType:             nil,
			SysLog:               nil,
			LogLevel:             nil,
			Trace:                nil,
			TraceDir:             nil,
			TraceBufferSize:      0,
			TraceMode:            nil,
			Version:              false,
			Help:                 false,
			InMemory:             false,
			BaseVirtAddr:         nil,
			NoTelemetry:          false,
			ForceMaxSimdBitwidth: 0,
			HugeUnlink:           nil,
			NoHugePages:          false,
			NoPCI:                false,
			NoHPET:               false,
			NoShConf:             false,
			SocketMemory:         nil,
			SocketLimit:          nil,
			HugeDir:              nil,
			FilePrefix:           nil,
			CreateUIODev:         false,
			VfioIntr:             nil,
			VfioVFToken:          nil,
			LegacyMemory:         nil,
			SingleFileSegments:   false,
			MatchAllocations:     false,
			HugeWorkerStack:      0, // Size in KBytes
		},
	}
}

func listToStr(lst []string) string {
	s := ""
	l := len(lst) - 1
	for n, b := range lst {
		s += fmt.Sprintf("%q", b)
		if n < l {
			s += ","
		}
	}

	return s
}

func (cd *configData) validateConfig() error {

	if cd.CoreList == nil {
		return fmt.Errorf("core-list is empty")
	}
	return nil
}

func (cs *System) openText() error {

	text := jsonc.ToJSON(bytes.TrimSpace(cs.cBytes))

	if len(text) == 0 {
		return fmt.Errorf("empty json text string")
	}

	// test for JSON string, which must start with a '{'
	if text[0] != '{' {
		return fmt.Errorf("string does not appear to be a valid JSON text missing starting '{'")
	}

	// Unmarshal json text into the Config structure
	if err := json.Unmarshal(text, &cs.cd); err != nil {
		return err
	}
	return cs.cd.validateConfig()
}

// readFile by passing in a filename or path to a JSON-C or JSON configuration
func (cs *System) readFile(s string) error {
	b, err := os.ReadFile(s)
	if err != nil {
		return err
	}
	cs.cBytes = b
	return nil
}

func (cs *System) Open(s string) error {

	if len(s) == 0 {
		s = "{}"
	}
	if err := cs.readFile(s); err != nil {
		cs.cBytes = []byte(s)
	}
	return cs.openText()
}

func (cs *System) String() string {

	if data, err := json.MarshalIndent(&cs.cd, "", "  "); err != nil {
		return fmt.Sprintf("error marshalling JSON: %v", err)
	} else {
		return string(data)
	}
}

func (cs *System) CoreMask() string {

	if cs.cd.CoreMask == nil {
		return ""
	}
	return *cs.cd.CoreMask
}

func (cs *System) CoreList() string {

	if cs.cd.CoreList == nil {
		return ""
	}
	return *cs.cd.CoreList
}

func (cs *System) CoreMap() string {

	if cs.cd.CoreMap == nil {
		return ""
	}
	return *cs.cd.CoreMap
}

func (cs *System) ServiceCoreMask() string {

	if cs.cd.ServiceCoreMask == nil {
		return ""
	}
	return *cs.cd.ServiceCoreMask
}

func (cs *System) MainLCore() int {

	return cs.cd.MainLCore
}

func (cs *System) MBufPoolOpsName() string {

	if cs.cd.MBufPoolOpsName == nil {
		return ""
	}
	return *cs.cd.MBufPoolOpsName
}

func (cs *System) NumChannels() int {

	return cs.cd.NumChannels
}

func (cs *System) MemorySize() uint64 {

	return cs.cd.MemorySize
}

func (cs *System) NumRanks() int {

	return cs.cd.NumRanks
}

func (cs *System) BlockList() []string {

	bl := []string{}
	for _, k := range cs.cd.BlockList {
		bl = append(bl, *k)
	}
	return bl
}

func (cs *System) AllowList() []string {

	al := []string{}
	for _, k := range cs.cd.AllowList {
		al = append(al, *k)
	}
	return al
}

func (cs *System) VirtualDevices() []string {

	vd := []string{}
	for _, k := range cs.cd.VirtualDevices {
		vd = append(vd, *k)
	}
	return vd
}

func (cs *System) IovaMode() string {

	if cs.cd.IovaMode == nil {
		return ""
	}
	return *cs.cd.IovaMode
}

func (cs *System) Drivers() []string {

	dr := []string{}
	for _, k := range cs.cd.Drivers {
		dr = append(dr, *k)
	}
	return dr
}

func (cs *System) VmWareTSCMap() string {
	if cs.cd.VmWareTSCMap == nil {
		return ""
	}
	return *cs.cd.VmWareTSCMap
}

func (cs *System) ProcType() string {

	if cs.cd.ProcType == nil {
		return ""
	}
	return *cs.cd.ProcType
}

func (cs *System) SysLog() string {

	if cs.cd.SysLog == nil {
		return ""
	}
	return *cs.cd.SysLog
}

func (cs *System) LogLevel() string {

	if cs.cd.LogLevel == nil {
		return ""
	}
	return *cs.cd.LogLevel
}

func (cs *System) Version() bool {

	return cs.cd.Version
}

func (cs *System) Help() bool {

	return cs.cd.Help
}

func (cs *System) InMemory() bool {

	return cs.cd.InMemory
}

func (cs *System) BaseVirtAddr() string {
	if cs.cd.BaseVirtAddr == nil {
		return ""
	}
	return *cs.cd.BaseVirtAddr
}

func (cs *System) NoTelemetry() bool {

	return cs.cd.NoTelemetry
}

func (cs *System) ForceMaxSimdBitwidth() int {

	return cs.cd.ForceMaxSimdBitwidth
}

func (cs *System) HugeUnlink() string {

	if cs.cd.HugeUnlink == nil {
		return ""
	}
	return *cs.cd.HugeUnlink
}

func (cs *System) NoHugePages() bool {

	return cs.cd.NoHugePages
}

func (cs *System) NoPCI() bool {

	return cs.cd.NoPCI
}

func (cs *System) NoHPET() bool {

	return cs.cd.NoHPET
}

func (cs *System) NoShConf() bool {

	return cs.cd.NoShConf
}

func (cs *System) SocketMemory() []string {

	sm := []string{}
	for _, k := range cs.cd.SocketMemory {
		sm = append(sm, *k)
	}
	return sm
}

func (cs *System) SocketLimit() []string {

	sl := []string{}
	for _, k := range cs.cd.SocketMemory {
		sl = append(sl, *k)
	}
	return sl
}

func (cs *System) HugeDir() string {

	if cs.cd.HugeDir == nil {
		return ""
	}
	return *cs.cd.HugeDir
}

func (cs *System) FilePrefix() string {

	if cs.cd.FilePrefix == nil {
		return ""
	}
	return *cs.cd.FilePrefix
}

func (cs *System) CreateUIODev() bool {

	return cs.cd.CreateUIODev
}

func (cs *System) VfioIntr() string {
	if cs.cd.VfioIntr == nil {
		return ""
	}
	return *cs.cd.VfioIntr
}

func (cs *System) VfioVFToken() string {
	if cs.cd.VfioVFToken == nil {
		return ""
	}
	return *cs.cd.VfioVFToken
}

func (cs *System) LegacyMemory() string {
	if cs.cd.LegacyMemory == nil {
		return ""
	}
	return *cs.cd.LegacyMemory
}

func (cs *System) SingleFileSegments() bool {
	return cs.cd.SingleFileSegments
}

func (cs *System) MatchAllocations() bool {
	return cs.cd.MatchAllocations
}

func (cs *System) HugeWorkerStack() int {
	return cs.cd.HugeWorkerStack
}

func (cs *System) MakeArgs() ([]string, error) {

	argv := []string{"dpdk"}

	if cs.CoreList() == "" {
		if cs.CoreMask() == "" {
			return nil, fmt.Errorf("core-list or core-mask is not specified")
		} else {
			argv = append(argv, options[IdxCoreMask], cs.CoreMask())
		}
	} else {
		argv = append([]string{}, options[IdxCoreList], cs.CoreList())
	}
	if cs.CoreMap() != "" {
		argv = append(argv, options[IdxCoreMap], cs.CoreMap())
	}
	if cs.ServiceCoreMask() != "" {
		argv = append(argv, options[IdxServiceCoreMask], cs.ServiceCoreMask())
	}
	if cs.MainLCore() >= 0 {
		argv = append(argv, options[IdxMainLCore], strconv.Itoa(cs.MainLCore()))
	}
	if cs.MBufPoolOpsName() != "" {
		argv = append(argv, options[IdxMBufPoolOpsName], cs.MBufPoolOpsName())
	}
	if chnls := cs.NumChannels(); chnls > 0 {
		argv = append(argv, options[IdxNumChannels], strconv.Itoa(chnls))
	}
	if mem := cs.MemorySize(); mem > 0 {
		argv = append(argv, options[IdxMemorySize], strconv.FormatUint(mem, 10))
	}
	if ranks := cs.NumRanks(); ranks > 0 {
		argv = append(argv, options[IdxNumRanks], strconv.Itoa(ranks))
	}

	if len(cs.BlockList()) > 0 && len(cs.AllowList()) > 0 {
		return nil, fmt.Errorf("block-list and allow-list are mutually exclusive")
	}
	if bl := cs.BlockList(); len(bl) > 0 {
		for _, b := range bl {
			argv = append(argv, options[IdxBlockList], fmt.Sprintf("%q", b))
		}
	} else if al := cs.AllowList(); len(al) > 0 {
		for _, a := range al {
			argv = append(argv, options[IdxAllowList], fmt.Sprintf("%q", a))
		}
	}
	if vd := cs.VirtualDevices(); len(vd) > 0 {
		for _, v := range vd {
			argv = append(argv, options[IdxVirtualDevices], fmt.Sprintf("%q", v))
		}
	}
	if cs.IovaMode() == "pa" || cs.IovaMode() == "va" {
		argv = append(argv, options[IdxIovaMode], cs.IovaMode())
	}
	if dr := cs.Drivers(); len(dr) > 0 {
		for _, d := range dr {
			argv = append(argv, options[IdxDrivers], fmt.Sprintf("%q", d))
		}
	}
	if cs.VmWareTSCMap() != "" {
		argv = append(argv, options[IdxVmWareTSCMap], cs.VmWareTSCMap())
	}
	if cs.ProcType() == "primary" || cs.ProcType() == "secondary" || cs.ProcType() == "auto" {
		argv = append(argv, options[IdxProcType], cs.ProcType())
	}
	if cs.SysLog() != "" {
		argv = append(argv, options[IdxSysLog], cs.SysLog())
	}
	if cs.LogLevel() != "" {
		argv = append(argv, options[IdxLogLevel], fmt.Sprintf("%q", cs.LogLevel()))
	}
	if cs.Version() {
		argv = append(argv, options[IdxVersion])
	}
	if cs.Help() {
		argv = append(argv, options[IdxHelp])
	}
	if cs.InMemory() {
		argv = append(argv, options[IdxInMemory])
	}
	if cs.BaseVirtAddr() != "" {
		argv = append(argv, options[IdxBaseVirtAddr], cs.BaseVirtAddr())
	}
	if cs.NoTelemetry() {
		argv = append(argv, options[IdxNoTelemetry])
	}
	if cs.ForceMaxSimdBitwidth() > 0 {
		argv = append(argv, options[IdxForceMaxSimdBitwidth], strconv.Itoa(cs.ForceMaxSimdBitwidth()))
	}
	if cs.HugeUnlink() == "existing" || cs.HugeUnlink() == "always" || cs.HugeUnlink() == "never" {
		argv = append(argv, options[IdxHugeUnlink], cs.HugeUnlink())
	}
	if cs.NoHugePages() {
		argv = append(argv, options[IdxNoHugePages])
	}
	if cs.NoPCI() {
		argv = append(argv, options[IdxNoPCI])
	}
	if cs.NoHPET() {
		argv = append(argv, options[IdxNoHPET])
	}
	if cs.NoShConf() {
		argv = append(argv, options[IdxNoShConf])
	}
	if sm := cs.SocketMemory(); len(sm) > 0 {
		argv = append(argv, options[IdxSocketMemory], listToStr(sm))
	}
	if sl := cs.SocketLimit(); len(sl) > 0 {
		argv = append(argv, options[IdxSocketLimit], listToStr(sl))
	}
	if cs.HugeDir() != "" {
		argv = append(argv, options[IdxHugeDir], cs.HugeDir())
	}
	if cs.FilePrefix() != "" {
		argv = append(argv, options[IdxFilePrefix], cs.FilePrefix())
	}
	if cs.CreateUIODev() {
		argv = append(argv, options[IdxCreateUIODev])
	}
	if cs.VfioIntr() == "legacy" || cs.VfioIntr() == "msi" || cs.VfioIntr() == "msix" {
		argv = append(argv, options[IdxVfioIntr], cs.VfioIntr())
	}
	if cs.VfioVFToken() != "" {
		argv = append(argv, options[IdxVfioVFToken], cs.VfioVFToken())
	}
	if cs.LegacyMemory() != "" {
		argv = append(argv, options[IdxLegacyMemory], cs.LegacyMemory())
	}
	if cs.SingleFileSegments() {
		argv = append(argv, options[IdxSingleFileSegments])
	}
	if cs.MatchAllocations() {
		argv = append(argv, options[IdxMatchAllocations])
	}
	if cs.HugeWorkerStack() > 0 {
		argv = append(argv, options[cs.HugeWorkerStack()], strconv.Itoa(cs.HugeWorkerStack()))
	}

	if len(argv) == 0 {
		return nil, fmt.Errorf("no command line arguments specified")
	}

	return argv, nil
}
