// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package fserde

import (
	"testing"

	"github.com/franela/goblin"
)

var (
	SerializeDefaultFrames = []string{
		"Defaults-0 := Ether(src=00:01:02:03:04:FF, proto=0x800)/" +
			"IPv4(ver=4, len=20, dst=6.5.7.8)/" +
			"UDP(sport=5678, dport=1024)/" +
			"TSC()/" +
			"Payload(size=32, fill=0xaa)",
		"Defaults-1 := Ether(src=00:05:04:03:02:01, proto=0x800)/" +
			"Dot1Q(tpid=0x8100, vlan=0x22, cfi=1, prio=7)/" +
			"IPv4(ver=4, len=20, dst=9.8.7.6)/" +
			"UDP(sport=5678, dport=3456)/" +
			"TSC()/" +
			"Payload(size=32, fill=0xaa)",
		"Defaults-3 := Ether(dst=01:22:33:44:55:66, src=00:01:02:03:04:FF, proto=0x806)/" +
			"IPv4(ver=4, len=20, dst=6.5.7.8)/" +
			"UDP(sport=5678, dport=1024)/" +
			"TSC()/" +
			"Payload(size=32, fill=0xaa)",
	}
)

func TestSerializeBegin(t *testing.T) {

	g := goblin.Goblin(t)

	g.Describe("Frame Serialize tests - ", func() {
		defs := &FrameSerdeConfig{Defaults: SerializeDefaultFrames}

		g.It("Create", func() {
			if fg, err := Create("Test 0", defs); err != nil {
				g.Errorf("create failed: %s", err)
			} else {
				defer fg.Destroy()
				g.Assert(fg != nil).IsTrue("Create failed")
			}
		})
	})
}
