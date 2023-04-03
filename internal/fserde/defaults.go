/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation.
 */

package fserde

import (
	"fmt"
)

// DefaultsLayer is the default implementation of the Layer interface.
type DefaultsLayer struct {
	hdr  *LayerHdr
	name string
}

// String returns a string representation of the layer.
func (l *DefaultsLayer) String() string {
	return fmt.Sprintf("Defaults(%s)", l.name)
}

func DefaultsNew(fr *Frame) *DefaultsLayer {
	return &DefaultsLayer{
		hdr: LayerConstructor(fr, LayerDefaults, LayerDefaultsType),
	}
}

func (l *DefaultsLayer) Name() LayerName {
	return l.hdr.layerName
}

func (l *DefaultsLayer) Parse(opts string) error {

	if opts == "" {
		return nil
	}

	df, err := l.hdr.fr.serde.GetFrame(opts, DefaultFrameType)
	if err != nil {
		return err
	}

	l.hdr.fr.defaultsFrame = df
	l.name = opts

	return nil
}

func (l *DefaultsLayer) ApplyDefaults() error {

	return nil
}

func (l *DefaultsLayer) WriteLayer() error {

	return nil
}
