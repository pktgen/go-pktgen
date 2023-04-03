/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation.
 */

package fserde

import (
	"fmt"
	"strconv"
)

// Count() is a layer to duplicate the frame N number of times.
//    e.g., Count(10) -> duplicate the frame 10 times

// CountLayer the structure holding the information on each layer.
type CountLayer struct {
	hdr   *LayerHdr
	count uint32
}

func (l *CountLayer) String() string {
	return fmt.Sprintf("%s(%d)", l.hdr.layerName, l.count)
}

// CountNew creates a new CountLayer and is called the NewFuncs structure.
func CountNew(fr *Frame) *CountLayer {
	return &CountLayer{
		hdr: LayerConstructor(fr, LayerCount, LayerCountType),
	}
}

// Name returns the name of the layer.
func (l *CountLayer) Name() LayerName {
	return l.hdr.layerName
}

// Parse parses the layer options string.
func (l *CountLayer) Parse(opts string) error {

	if len(opts) > 0 {
		if v, err := strconv.ParseUint(opts, 0, 32); err != nil {
			return err
		} else {
			l.count = uint32(v)
		}
	}

	return nil
}

// ApplyDefaults applies the default values for the layer.
func (l *CountLayer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	dl, ok := d.GetLayer(LayerCount).(*CountLayer)
	if !ok {
		return nil
	}

	if dl.count > 0 && dl.count != l.count {
		l.count = dl.count
	}

	fmt.Printf("%s.ApplyDefaults: %T\n", l.Name(), dl)

	return nil
}

// WriteLayer writes the layer to hdr.frame []byte.
func (l *CountLayer) WriteLayer() error {

	return nil
}
