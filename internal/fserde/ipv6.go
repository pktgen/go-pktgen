/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation.
 */

package fserde

import (
	"fmt"

	"golang.org/x/net/ipv6"
)

type IPv6Layer struct {
	hdr    *LayerHdr
	ip6Hdr ipv6.Header
}

func (l *IPv6Layer) String() string {
	return fmt.Sprintf("%s()", l.hdr.layerName)
}

func IPv6New(fr *Frame) *IPv6Layer {
	return &IPv6Layer{
		hdr: LayerConstructor(fr, LayerIPv6, LayerIPv6Type),
	}
}

func (l *IPv6Layer) Name() LayerName {
	return l.hdr.layerName
}

func (l *IPv6Layer) Parse(opts string) error {

	return nil
}

func (l *IPv6Layer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	dl, ok := d.GetLayer(LayerIPv6).(*IPv6Layer)
	if !ok {
		return nil
	}

	fmt.Printf("%s.ApplyDefaults: %T\n", l.Name(), dl)

	return nil
}

func (l *IPv6Layer) WriteLayer() error {

	l.hdr.proto.name = l.Name()
	l.hdr.proto.offset = l.hdr.fr.GetOffset(l.Name())
	l.hdr.proto.length = 0

	l.hdr.fr.AddProtocol(&l.hdr.proto)

	return nil
}
