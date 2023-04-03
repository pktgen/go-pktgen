// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package gpcommon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func (cm ModeString) Value() CoreMode {

	switch strings.ToLower(string(cm)) {
	default:
		return UnknownMode
	case "main", "display", "management":
		return MainMode
	case "rx", "rxonly", "rx:only":
		return RxMode
	case "tx", "txonly", "tx:only":
		return TxMode
	case "rx/tx", "rxtx", "rxtx:mode", "rx:tx":
		return RxTxMode
	}
}

// (CoreMode)MarshalJSON decodes JSON value into string.
func (cm CoreMode) MarshalJSON() ([]byte, error) {

	return []byte(fmt.Sprintf("%q", cm.String())), nil
}

func (cm *CoreMode) UnmarshalJSON(b []byte) error {

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*cm = ModeString(s).Value()

	return nil
}

func (lp LPortID) MarshalJSON() ([]byte, error) {

	return []byte(fmt.Sprintf("%q", lp.String())), nil
}

// Marshal indented JSON string of the given object
func MarshalIndent(v interface{}) string {

	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err.Error())
	}
	return string(b)
}

func Indent(s string) string {
	var dst bytes.Buffer

	if err := json.Indent(&dst, []byte(s), "", "  "); err != nil {
		return s
	} else {
		return dst.String()
	}
}

func (l LinkState) MaxPktsPerSec() uint64 {

	speed := uint64(l.Speed) * Million

	bitsPerFrame := ((MinFrameSize + FrameOverheadSize) * 8)

	pps := speed / bitsPerFrame
	if pps == 0	{
		pps = 1 // avoid division by zero error
	}
	return pps
}
