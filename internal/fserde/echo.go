/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation.
 */

package fserde

import (
	"fmt"
)

type EchoLayer struct {
	hdr *LayerHdr
}

func (l *EchoLayer) String() string {
	return fmt.Sprintf("%s()", l.hdr.layerName)
}

func EchoNew(fr *Frame) *EchoLayer {
	return &EchoLayer{
		hdr: LayerConstructor(fr, LayerEcho, LayerEchoType),
	}
}

func (l *EchoLayer) Name() LayerName {
	return l.hdr.layerName
}

func (el *EchoLayer) Parse(opts string) error {

	return nil
}

func (l *EchoLayer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	dl, ok := d.GetLayer(LayerEcho).(*EchoLayer)
	if !ok {
		return nil
	}

	fmt.Printf("%s.ApplyDefaults: %T\n", l.Name(), dl)

	return nil
}

func (l *EchoLayer) WriteLayer() error {

	l.hdr.proto.name = l.Name()
	l.hdr.proto.offset = l.hdr.fr.GetOffset(l.Name())
	l.hdr.proto.length = 0

	l.hdr.fr.AddProtocol(&l.hdr.proto)

	return nil
}
