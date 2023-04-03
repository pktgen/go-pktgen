// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package fserde

import (
	"fmt"
	"strings"

	"testing"

	"github.com/franela/goblin"
)

var (
	pcapFileName   = "/tmp/fserde.pcap"
	toBinaryFrames = []string{
		"PortA := Ether(dst=00:11:22:33:44:55, src=00:11:22:33:44:99, proto=0x800)/" +
			"IPv4(ver=4, dst=10.0.0.1, src=10.0.0.2, ttl=64)/" +
			"UDP(sport=1111, dport=3333, checksum=true)/" +
			"Payload(string='Port-AAA')/Count(5)",
		"Port0 := Ether(dst=00:11:22:33:44:55, proto=0x800)/" +
			"Dot1Q(tpid=0x8100, pcp=3, dei=0, vlan=1)/ " +
			"IPv4(ver=4, dst=10.0.0.3, ttl=64)/" +
			"UDP(sport=5678, dport=1234, checksum=true)/" +
			"Defaults(Defaults-0)/Count(10)",
		"Port1 := Ether( dst=00:01:02:03:04:05 )/" +
			"QinQ(Dot1q{vlan=12}, Dot1q{vlan=212})/" +
			"IPv4(dst=10.0.10.1, src=10.0.10.2)/" +
			"UDP(sport=0x1234, dport=1234)/" +
			"Payload(string='cafefood')/" +
			"Defaults(Defaults-0)",
		"Port2 := Ether(dst=00:11:22:33:44:55, src = 01:ff:ff:ff:ff:ff )/" +
			"Dot1q(vlan=0x322, cfi=1, prio=7)/" +
			"IPv4(dst=10.0.20.1)/" +
			"UDP(sport=5699)/" +
			"Payload(size=128)/" +
			"Defaults(Defaults-0)",
		"Port3:=Ether(dst=2201:2203:4405)/" +
			"Dot1ad(vlan=0x22, cfi=1, prio=7)/" +
			"IPv4(dst=10.0.30.1)/" +
			"UDP(sport=5698, checksum=true)/" +
			"Payload(size=1)/" +
			"Defaults(Defaults-1)",
		"Port4:=Ether(dst=0133:0233:0333)/" +
			"Dot1Q(vlan=0x22, cfi=1, prio=7)/" +
			"Dot1ad(vlan=0x33, cfi=1, prio=7)/" +
			"IPv4(dst=10.0.40.1)/" +
			"TCP(sport=5697, dport=3000, seq=5000, ack=5001, flags=[SYN | ACK | FIN])/" +
			"Defaults(Defaults-1)",
		"Port5:=Ether(dst=0133:0233:0333)/" +
			"Dot1Q(vlan=0x22, cfi=1, prio=7)/" +
			"Dot1ad(vlan=0x33, cfi=1, prio=7)/" +
			"IPv4(dst=192.168.0.1)/" +
			"TCP(sport=5696, dport=2000, flags=0x12)/" +
			"Defaults(Defaults-1)",
		"Port6:=Ether()/" +
			"IPv4(dst=192.168.1.1)/" +
			"TCP(sport=5685, dport=1000, seq=4000, ack=4001, window=1024, flags=[ACK | PSH])/" +
			"Defaults(Defaults-2)",
	}
	toBinaryDefaultFrames = []string{
		"Defaults-0 := Ether(src=00:01:02:03:04:FF, proto=0x800)/" +
			"IPv4(ver=4, src=10.0.0.2)/" +
			"UDP(sport=1001, dport=1024)/" +
			"Payload(size=8, fill16=0xaabb)",
		"Defaults-1 := Ether(src=00:05:04:03:02:01, proto=0x800)/" +
			"Dot1Q(tpid=0x8100, vlan=0x22, cfi=1, prio=7)/" +
			"IPv4(ver=4, src=9.8.7.6)/" +
			"UDP(sport=1002, dport=3456)/" +
			"Payload(size=24, fill=0xac)",
		"Defaults-2 := Ether(dst=01:22:33:44:55:66, src=00:01:02:03:04:FF, proto=0x800)/" +
			"IPv4(ver=4, src=6.5.7.8)/" +
			"UDP(sport=1003, dport=1024)/" +
			"Payload(size=32, fill=0xab)",
	}
	toBinaryInvalidFrames = []string{
		"Invalid0:=Ether(src=2201:2203:4405)/" +
			"Dot1QX(vlan=0x22, cfi=1, prio=7)/" +
			"QinQ(Dot1q{vlan=0x33, cfi=1, prio=7}, Dot1q{})/" +
			"IPv4(src=192.18.1.25)/" +
			"TCP(sport=5678)",
	}
	toBinaryInvalidDefaults = []string{
		"Defaults-A := Ether(src=00:01:02:03:04:FF, proto=0x800)/" +
			"IPv4(ver=4, dst=6.5.7.8)/" +
			"UDP(sport=5678, dport=1024)/" +
			"Payload(size=32, fill=0xaa)/ Defaults  ( Defaults-B)",
	}
)

func TestDeserializeBegin(t *testing.T) {

	g := goblin.Goblin(t)

	g.Describe("Frame-Serde ToBinary tests - ", func() {
		defs := &FrameSerdeConfig{Defaults: toBinaryDefaultFrames}

		g.It("Create", func() {
			if fg, err := Create("Test 0", nil); err != nil {
				g.Errorf("create failed: %s", err)
			} else {
				defer fg.Destroy()
				g.Assert(fg != nil).IsTrue("Create failed")
			}
		})

		g.It("Create with zero Defaults", func() {
			if fg, err := Create("Test 0", &FrameSerdeConfig{}); err != nil {
				g.Errorf("create failed: %s", err)
			} else {
				defer fg.Destroy()
				g.Assert(fg != nil).IsTrue("Create failed")
			}
		})

		g.It("Create with Defaults", func() {
			if fg, err := Create("Test 0", defs); err != nil {
				g.Errorf("create failed: %s", err)
			} else {
				defer fg.Destroy()
				g.Assert(fg != nil).IsTrue("Create failed")
			}
		})

		g.It("Destroy", func() {
			if fg, err := Create("Test 0.1", defs); err != nil {
				g.Errorf("create failed: %s", err)
			} else {
				defer fg.Destroy()
				g.Assert(fg != nil).IsTrue("Create failed")
			}
		})

		g.It("CreateInvalid", func() {
			if fg, err := Create("Test 1", &FrameSerdeConfig{Defaults: toBinaryInvalidDefaults}); err != nil {
				g.Assert(fg == nil).IsTrue("Create invalid defaults failed")
			} else {
				defer fg.Destroy()
			}
		})

		g.It("ToBinary", func() {
			if fg, err := Create("Test 2", defs); err != nil {
				g.Errorf("create failed: %s", err)
			} else {
				defer fg.Destroy()

				g.Assert(fg != nil).IsTrue("Create failed")

				err := fg.StringsToBinary(toBinaryFrames)
				g.Assert(err == nil).IsTrue(fmt.Sprintf("StringsToBinary failed: %v", err))

				name := fg.FrameNames(NormalFrameType)[3]
				fr, err := fg.GetFrame(name, NormalFrameType)
				g.Assert(err == nil && strings.EqualFold(fr.name, name) == true).IsTrue(fmt.Sprintf("getting a specific frame '%v' failed: %v", "Port4", err))

				err = fg.DeleteFrame(name, NormalFrameType)
				g.Assert(err == nil).IsTrue(fmt.Sprintf("failed to delete '%v': %v", name, err))
			}
		})

		g.It("ToBinary Write to PCAP", func() {
			if fg, err := Create("Test 2", defs); err != nil {
				g.Errorf("create failed: %s", err)
			} else {
				defer fg.Destroy()

				g.Assert(fg != nil).IsTrue("Create failed")

				err := fg.StringsToBinary(toBinaryFrames)
				g.Assert(err == nil).IsTrue(fmt.Sprintf("StringsToBinary failed: %v", err))

				fmt.Printf("Write PCAP file %s\n", pcapFileName)
				if err := fg.WritePCAP(pcapFileName, NormalFrameType); err != nil {
					g.Assert(err == nil).IsTrue(fmt.Sprintf("failed to write PCAP file '%v': %v", pcapFileName, err))
				}
			}
		})

		g.It("ToBinary Frames", func() {
			if fg, err := Create("Test 3", defs); err != nil {
				g.Errorf("create failed: %s", err)
			} else {
				defer fg.Destroy()

				g.Assert(fg != nil).IsTrue("Create failed")

				err := fg.StringToBinary(toBinaryFrames[3])
				g.Assert(err == nil).IsTrue(fmt.Sprintf("StringsToBinary failed: %v", err))

				name := fg.FrameNames(NormalFrameType)[0]
				err = fg.DeleteFrame(name, NormalFrameType)
				g.Assert(err == nil).IsTrue(fmt.Sprintf("failed to delete '%v': %v", name, err))

				// Try to delete a frame that doesn't exist
				err = fg.DeleteFrame(name, NormalFrameType)
				g.Assert(err != nil).IsTrue(fmt.Sprintf("delete of '%v' should hve failed", name))
			}
		})

		g.It("ToBinary Invalid frames", func() {
			if fg, err := Create("Test 4", defs); err != nil {
				g.Errorf("create failed: %s", err)
			} else {
				defer fg.Destroy()

				g.Assert(fg != nil).IsTrue("Create failed")
				err := fg.StringToBinary(toBinaryInvalidFrames[0])
				g.Assert(err != nil).IsTrue(fmt.Sprintf("EncodeFrames failed: %v", err))

				if len(fg.FrameNames(NormalFrameType)) > 0 {
					name := fg.FrameNames(NormalFrameType)[0]
					if len(name) > 0 {
						err = fg.DeleteFrame(name, NormalFrameType)
						g.Assert(err == nil).IsTrue(fmt.Sprintf("failed to delete '%v': %v", name, err))

						// Try to delete a frame that doesn't exist
						err = fg.DeleteFrame(name, NormalFrameType)
						g.Assert(err != nil).IsTrue(fmt.Sprintf("delete of '%v' should hve failed", name))
					}
				}
			}
		})
	})
}
