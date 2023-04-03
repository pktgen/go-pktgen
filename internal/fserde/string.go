/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2022-2024 Intel Corporation.
 */

package fserde

import (
	//"fmt"
	"sort"
)

// String converts a EncodedFrame map to a string.
func (f FrameMap) String() string {
	s := ""

	if len(f) == 0 {
		return s
	}
	keys := make([]*Frame, 0, len(f))

	for _, frame := range f {
		keys = append(keys, frame)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i].name < keys[j].name })

	for _, v := range keys {
		s += v.String() + "\n"
	}
	return s[0 : len(s)-1]
}

// String converts a protocol map into a string.
/*
func (pm ProtoMap) String() string {
	s := ""

	if len(pm) == 0 {
		return s
	}
	keys := make([]*Proto, 0, len(pm))

	for _, pm := range pm {
		keys = append(keys, pm)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i].index < keys[j].index })

	for _, v := range keys {
		s += fmt.Sprintf("%v{off: %d, len: %d} ",
			LayerNames[v.index].hdr.layerName, v.offset, v.length)
	}
	return s[0 : len(s)-1]
}
*/
