/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2022-2024 Intel Corporation.
 */

package fserde

type TSCLayer struct {
	hdr *LayerHdr
}

func (l *TSCLayer) String() string {
	return "TSC()"
}

func TSCNew(fr *Frame) *TSCLayer {
	return &TSCLayer{
		hdr: LayerConstructor(fr, LayerTSC, LayerTSCType),
	}
}

func (l *TSCLayer) Name() LayerName {
	return l.hdr.layerName
}

func (l *TSCLayer) Parse(opts string) error {

	l.hdr.proto.name = l.Name()
	l.hdr.proto.offset = l.hdr.fr.GetOffset(l.Name())
	l.hdr.proto.length = 4

	l.hdr.fr.AddProtocol(&l.hdr.proto)

	return nil
}

func (l *TSCLayer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	_, ok := d.GetLayer(LayerTSC).(*TSCLayer)
	if !ok {
		return nil
	}

	dbug.Printf("%v\n", l)

	return nil
}

func (l *TSCLayer) WriteLayer() error {

	return nil
}
