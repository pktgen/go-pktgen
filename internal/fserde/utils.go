/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation.
 */

package fserde

import (
	"fmt"
)

// frameDump is a helper function that returns a string representation of the
// given frame.
func (fr *Frame) FrameDump() string {

	s := fmt.Sprintf("%v(len %d: %v\n  ", fr.name, fr.frame.Len(), fr.protocols)
	for _, v := range fr.frame.Bytes() {
		s += fmt.Sprintf("%02x ", v)
	}
	s += ")\n"

	return s
}

func Roundup(v, mul uint64) uint64 {
	mul--
	return (v + mul) &^ mul
}
