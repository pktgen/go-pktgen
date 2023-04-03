/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2022-2024 Intel Corporation.
 */

package fserde

import (
	"fmt"
)

type SCTPLayer struct {
	hdr *LayerHdr
}

func (l *SCTPLayer) String() string {
	return fmt.Sprintf("%s()", l.hdr.layerName)
}

func SCTPNew(fr *Frame) *SCTPLayer {
	return &SCTPLayer{
		hdr: LayerConstructor(fr, LayerSCTP, LayerSCTPType),
	}
}

func (l *SCTPLayer) Name() LayerName {
	return l.hdr.layerName
}

func (l *SCTPLayer) Parse(opts string) error {

	l.hdr.proto.name = l.Name()
	l.hdr.proto.offset = l.hdr.fr.GetOffset(l.Name())
	l.hdr.proto.length = 0

	l.hdr.fr.AddProtocol(&l.hdr.proto)

	return nil
}

func (l *SCTPLayer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	dl, ok := d.GetLayer(LayerSCTP).(*SCTPLayer)
	if !ok {
		return nil
	}

	fmt.Printf("%s.ApplyDefaults: %T\n", l.Name(), dl)

	return nil
}

func (l *SCTPLayer) WriteLayer() error {

	return nil
}
