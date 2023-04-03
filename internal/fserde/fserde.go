// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package fserde

import (
	"fmt"
)

const (
	NormalFrameType = iota
	DefaultFrameType
	AllFrameTypes
)

type FrameType int

// ProtoInfo is the protocol offset and length information in the packet data.
type ProtoInfo struct {
	name   LayerName // Name of the protocol layer
	offset uint16    // Offset to the beginning of the protocol
	length uint16    // Length of the protocol
}

type ProtoMap map[LayerName]*ProtoInfo  // Protocol Mapping
type LayerMap map[LayerName]interface{} // interface{} is the given layer structure

type LayerInfo struct {
	Name  LayerName   // Name of the layer i.e., "TCP" or "UDP" or "ipv4" or ...
	Type  LayerType   // Layer index or type value
	Opts  string      // Layer strings
	Layer interface{} // Layer structure pointer
}

// Frame is the representation of a packet in text and binary format.
// A frame string or binary representation of a packet string is deserialized/serialized
// into this structure.
type Frame struct {
	serde         *FrameSerde  // FrameSerde pointer
	frameType     FrameType    // frame type or index value
	name          string       // Name of the frame used to map to the frame data.
	layerInfo     []*LayerInfo // List of layers in the frame in added order.
	layersMap     LayerMap     // Map of frame layers strings by layer name.
	protocols     []*ProtoInfo // List of protocols in the frame with offsets and lengths.
	defaultsFrame *Frame       // The default frame data
	frame         *MyBuffer    // Frame binary data.
}

type FrameKey struct {
	ftype FrameType
	name  string
}
type FrameMap map[FrameKey]*Frame // Frame Mapping between frame type and frame

// Serde is the main structure of the fserde package.
// Holding the deserialized and serialized frame data.
type FrameSerde struct {
	name       string   // Name of the frame-serde instance
	frames     FrameMap // Map of Frame structures
	frameNames []string // List of frame names in same order as they were deserialized.
}

// FrameSerdeConfig is the configuration structure for the FrameSerde.Create() call.
type FrameSerdeConfig struct {
	Defaults []string // List of default layers to apply to the frames
}

// String converts a Frame structure to a string.
func (f *Frame) String() string {
	s := fmt.Sprintf("%v:=", f.name)
	for _, layer := range f.layerInfo {
		s += fmt.Sprintf("%v/", layer.Layer)
	}

	return s[:len(s)-1]
}

func (p *ProtoInfo) String() string {
	return fmt.Sprintf("offset=%v, length=%v", p.offset, p.length)
}

// GetLayer returns a pointer to the layer with the given name.
func (fr *Frame) GetLayer(name LayerName) interface{} {

	if v, ok := fr.layersMap[name]; ok {
		return v
	}
	return nil
}

func (fr *Frame) GetProtocolID() int {

	if _, ok := fr.GetLayer(LayerUDP).(*UDPLayer); ok {
		return ProtocolUDP
	} else if _, ok := fr.GetLayer(LayerTCP).(*TCPLayer); ok {
		return ProtocolTCP
	} else if _, ok := fr.GetLayer(LayerICMPv4).(*ICMPv4Layer); ok {
		return ProtocolICMPv4
	} else if _, ok := fr.GetLayer(LayerICMPv6).(*ICMPv6Layer); ok {
		return ProtocolICMPv6
	}
	return 0
}

func (fr *Frame) AddProtocol(proto *ProtoInfo) error {

	fr.protocols = append(fr.protocols, proto)

	return nil
}

// GetProtocol returns a pointer to the protocol with the given name.
func (fr *Frame) GetProtocol(name LayerName) *ProtoInfo {

	for _, proto := range fr.protocols {
		if proto.name == name {
			return proto
		}
	}
	return nil
}

// GetOffset returns the offset to the given layer name.
// The offset is the offset to the beginning of the protocol layer name
func (fr *Frame) GetOffset(name LayerName) uint16 {

	var offset uint16 = 0

	for _, proto := range fr.protocols {
		if proto.name == name { // Stop at the given layer
			break
		}
		offset += proto.length
	}
	return offset
}

// GetLength returns the length of the following layers starting with the given layer name.
func (fr *Frame) GetLength(name LayerName) uint16 {

	var length uint16 = 0

	found := false
	for _, proto := range fr.protocols {
		if !found && proto.name == name { // Start at the given layer
			found = true
			length += proto.length
			continue
		}
		if found {
			length += proto.length
		}
	}
	return length
}

// Create a FrameSerde structure from the default values.
// If the default values are present then parse them to the FrameSerde structure.
func Create(name string, cfg *FrameSerdeConfig) (*FrameSerde, error) {

	if len(name) == 0 {
		return nil, fmt.Errorf("missing frame-serde name")
	}
	fserde := &FrameSerde{
		name:   name,
		frames: make(FrameMap),
	}
	if cfg != nil && len(cfg.Defaults) > 0 {
		if err := fserde.DefaultsToBinary(cfg.Defaults); err != nil {
			return nil, err
		}
	}
	return fserde, nil
}

func (f *FrameSerde) Delete() {
	f.frames = make(FrameMap)
}

func (f *FrameSerde) Destroy() {
	f.Delete()
	f.name = ""
}

func (f *FrameSerde) FrameNames(ftype FrameType) []string {

	keys := make([]string, 0, len(f.frames))

	for _, k := range f.frameNames {
		if f, ok := f.frames[FrameKey{ftype: ftype, name: k}]; ok {
			keys = append(keys, f.name)
		}
	}
	return keys
}

func (f *FrameSerde) GetFrame(frameName string, ftype FrameType) (*Frame, error) {

	key := FrameKey{name: frameName, ftype: ftype}
	if frame, ok := f.frames[key]; ok {
		return frame, nil
	} else {
		return nil, fmt.Errorf("frame (%+v) does not exist", key)
	}
}

func (f *FrameSerde) GetFrames(ftype FrameType) []*Frame {
	frames := make([]*Frame, 0, len(f.frames))
	for _, name := range f.FrameNames(NormalFrameType) {
		if f, ok := f.frames[FrameKey{ftype: ftype, name: name}]; ok {
			frames = append(frames, f)
		}
	}
	return frames
}

func (f *FrameSerde) DeleteFrame(frameName string, ftype FrameType) error {

	key := FrameKey{name: frameName, ftype: ftype}
	if _, ok := f.frames[key]; ok {
		delete(f.frames, key)
		return nil
	} else {
		return fmt.Errorf("frame %+v not found", key)
	}
}

func (f *FrameSerde) FrameMap(ftype FrameType) FrameMap {

	frameMap := make(FrameMap)

	for k, v := range f.frames {
		if ftype == AllFrameTypes || k.ftype == ftype {
			frameMap[k] = v
		}
	}

	return frameMap
}
