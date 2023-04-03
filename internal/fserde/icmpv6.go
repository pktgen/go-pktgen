/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2022-2024 Intel Corporation.
 */

package fserde

import (
	"fmt"
)

type ICMPv6Layer struct {
	hdr *LayerHdr
}

func (l *ICMPv6Layer) String() string {
	return fmt.Sprintf("%s()", l.hdr.layerName)
}

func ICMPv6New(fr *Frame) *ICMPv6Layer {
	return &ICMPv6Layer{
		hdr: LayerConstructor(fr, LayerICMPv6, LayerICMPv6Type),
	}
}

func (l *ICMPv6Layer) Name() LayerName {
	return l.hdr.layerName
}

func (el *ICMPv6Layer) Parse(opts string) error {

	return nil
}

func (l *ICMPv6Layer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	dl, ok := d.GetLayer(LayerICMPv6).(*ICMPv6Layer)
	if !ok {
		return nil
	}

	fmt.Printf("%s.ApplyDefaults: %T\n", l.Name(), dl)

	return nil
}

func (l *ICMPv6Layer) WriteLayer() error {

	l.hdr.proto.name = l.Name()
	l.hdr.proto.offset = l.hdr.fr.GetOffset(l.Name())
	l.hdr.proto.length = 0

	l.hdr.fr.AddProtocol(&l.hdr.proto)

	return nil
}
