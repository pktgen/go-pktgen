/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation.
 */

package fserde

import (
	"fmt"
)

type VxLanLayer struct {
	hdr       *LayerHdr
	vlanFlags uint8  // 8 bits VLAN flags
	vni       uint32 // 24 bits VNI value
}

func (vx *VxLanLayer) String() string {
	return fmt.Sprintf("VxLan: { flags=0x%02x, vni=0x%06x}", vx.vlanFlags, vx.vni)
}

func VxLanNew(fr *Frame) *VxLanLayer {
	return &VxLanLayer{
		hdr: LayerConstructor(fr, LayerVxLan, LayerVxLanType),
	}
}

func (l *VxLanLayer) Name() LayerName {
	return l.hdr.layerName
}

func (l *VxLanLayer) Parse(opts string) error {

	l.hdr.proto.name = l.Name()
	l.hdr.proto.offset = l.hdr.fr.GetOffset(l.Name())
	l.hdr.proto.length = 0

	l.hdr.fr.AddProtocol(&l.hdr.proto)

	return nil
}

func (l *VxLanLayer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	dl, ok := d.GetLayer(LayerVxLan).(*VxLanLayer)
	if !ok {
		return nil
	}

	fmt.Printf("%s.ApplyDefaults: %T\n", l.Name(), dl)

	return nil
}

func (l *VxLanLayer) WriteLayer() error {

	return nil
}
