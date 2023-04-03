/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation.
 */

package fserde

import (
	"fmt"
)

type Dot1adLayer struct {
	hdr *LayerHdr
}

func Dot1adNew(fr *Frame) *Dot1adLayer {
	return &Dot1adLayer{
		hdr: LayerConstructor(fr, LayerDot1AD, LayerDot1ADType),
	}
}

func (l *Dot1adLayer) String() string {
	return fmt.Sprintf("%s()", l.hdr.layerName)
}

func (d *Dot1adLayer) Name() LayerName {
	return d.hdr.layerName
}

func (l *Dot1adLayer) Parse(opts string) error {

	l.hdr.proto.name = l.Name()
	l.hdr.proto.offset = l.hdr.fr.GetOffset(l.Name())
	l.hdr.proto.length = 0

	l.hdr.fr.AddProtocol(&l.hdr.proto)

	return nil
}

func (l *Dot1adLayer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	dl, ok := d.GetLayer(LayerDot1AD).(*Dot1adLayer)
	if !ok {
		return nil
	}

	fmt.Printf("%s.ApplyDefaults: %T\n", l.Name(), dl)

	return nil
}

func (l *Dot1adLayer) WriteLayer() error {

	return nil
}
