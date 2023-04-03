// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022-2024 Intel Corporation

package fserde

import (
	"bytes"
	"fmt"

	"github.com/pktgen/go-pktgen/internal/pcap"
)

func (fg *FrameSerde) WritePCAP(path string, frameType FrameType) error {

	if len(path) == 0 {
		return fmt.Errorf("path is empty")
	}
	if pc := pcap.New(); pc == nil {
		return fmt.Errorf("failed to create pcap file %s", path)
	} else {
		for _, name := range fg.frameNames {
			key := FrameKey{name: name, ftype: frameType}
			if fr, ok := fg.frames[key]; ok {
				b := fr.frame.Bytes()

				// pad out the packet length to the minimum packet length (60).
				if fr.frame.Len() < MinPacketLen {
					b = append(b, bytes.Repeat([]byte("\x00"), MinPacketLen-fr.frame.Len())...)
				}
				cl := fr.layersMap[LayerCount].(*CountLayer)
				for i := 0; i < int(cl.count); i++ {
					pc.AddPacket(b)
				}
			}
		}
		pc.Write(path)
	}

	return nil
}
