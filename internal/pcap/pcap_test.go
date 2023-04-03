// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022-2024 Intel Corporation

package pcap

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/franela/goblin"
)

func mkPkt() []byte {
	/*
		Frame: 60 bytes on wire (480 bits), 60 bytes captured (480 bits)
		Ethernet II, Src: IntelCor_e4:38:44 (3c:fd:fe:e4:38:44), Dst: IntelCor_e4:34:c0 (3c:fd:fe:e4:34:c0)
		Internet Protocol Version 4, Src: 198.18.0.1, Dst: 198.18.1.1
		User Datagram Protocol, Src Port: 1234, Dst Port: 5678
		Data (18 bytes)
			Data: 6b6c6d6e6f707172737475767778797a3031
			[Length: 18]
		0000   3c fd fe e4 34 c0 3c fd fe e4 38 44 08 00 45 00   <...4.<...8D..E.
		0010   00 2e 60 ac 00 00 40 11 8c ec c6 12 00 01 c6 12   ..`...@.........
		0020   01 01 04 d2 16 2e 00 1a 93 c6 6b 6c 6d 6e 6f 70   ..........klmnop
		0030   71 72 73 74 75 76 77 78 79 7a 30 31               qrstuvwxyz01
	*/
	data, err := hex.DecodeString(
		"3cfdfee434c03cfdfee4384408004500" +
			"002e60ac000040118cecc6120001c612" +
			"010104d2162e001a93c66b6c6d6e6f70" +
			"7172737475767778797a3031")
	if err != nil {
		return []byte{}
	}
	return data
}

func TestDecodeBegin(t *testing.T) {

	g := goblin.Goblin(t)

	g.Describe("PCAP tests - ", func() {
		g.It("New", func() {
			if pcap := New(); pcap == nil {
				g.Errorf("Failed to create pcap")
			} else {
				pcap.SetLinkType(LinkTypeEthernet).SetSpanLen(MaxSpanLength)

				for i := 0; i < 16; i++ {
					time.Sleep(time.Millisecond * 10) // add some delay for time stamps
					pcap.AddPacket(mkPkt())
				}

				pcap.Write("/tmp/data.pcap")
			}
		})
	})
}
