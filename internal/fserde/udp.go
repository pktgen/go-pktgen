/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation.
 */

package fserde

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	UDPHeaderLen      = 8
	UDPDefaultLen     = 8
	UDPChecksumOffset = 6
)

type UDPHdr struct {
	SrcPort  uint16
	DstPort  uint16
	Length   uint16
	Checksum bool
}

type UDPLayer struct {
	hdr    *LayerHdr
	udpHdr UDPHdr
}

func (u *UDPLayer) String() string {
	hdr := u.udpHdr

	return fmt.Sprintf("UDP(sport=%d, dport=%d, length=%d, checksum=%v)",
		hdr.SrcPort, hdr.DstPort, hdr.Length, hdr.Checksum)
}

func UDPNew(fr *Frame) *UDPLayer {
	return &UDPLayer{
		hdr: LayerConstructor(fr, LayerUDP, LayerUDPType),
	}
}

func (l *UDPLayer) Name() LayerName {
	return l.hdr.layerName
}

func (u *UDPLayer) Parse(opts string) error {

	options := strings.Split(opts, ",")

	for _, opt := range options {
		opt = strings.TrimSpace(opt)

		kvp := strings.Split(opt, "=")
		if len(kvp) != 2 {
			return fmt.Errorf("invalid option: %s", opt)
		}

		key := strings.ToLower(strings.TrimSpace(kvp[0]))
		val := strings.ToLower(strings.TrimSpace(kvp[1]))

		switch key {
		case "sport", "srcport", "src":
			if v, err := strconv.ParseInt(val, 0, 0); err != nil {
				return err
			} else {
				u.udpHdr.SrcPort = uint16(v)
			}
		case "dport", "dstport", "dst":
			if v, err := strconv.ParseInt(val, 0, 0); err != nil {
				return err
			} else {
				u.udpHdr.DstPort = uint16(v)
			}
		case "checksum":
			switch val {
			case "on", "yes", "true", "enable", "enabled", "1":
				u.udpHdr.Checksum = true
			case "off", "no", "false", "disable", "disabled", "0":
				u.udpHdr.Checksum = false
			default:
				return fmt.Errorf("checksum invalid value: %s", val)
			}
		}
	}
	u.udpHdr.Length = UDPDefaultLen

	u.hdr.proto.name = u.Name()
	u.hdr.proto.offset = u.hdr.fr.GetOffset(u.Name())
	u.hdr.proto.length = UDPDefaultLen

	u.hdr.fr.AddProtocol(&u.hdr.proto)

	return nil
}

func (l *UDPLayer) ApplyDefaults() error {

	df := l.hdr.fr.defaultsFrame
	if df == nil {
		return nil
	}

	dl, ok := df.GetLayer(LayerUDP).(*UDPLayer)
	if !ok {
		return nil
	}

	if l.udpHdr.SrcPort == 0 && dl.udpHdr.SrcPort != 0 {
		l.udpHdr.SrcPort = dl.udpHdr.SrcPort
	}
	if l.udpHdr.DstPort == 0 && dl.udpHdr.DstPort != 0 {
		l.udpHdr.DstPort = dl.udpHdr.DstPort
	}
	// The frame checksum is false, but default is true set checksum to true
	// can not set checksum to false from the default frame
	if !l.udpHdr.Checksum && dl.udpHdr.Checksum {
		l.udpHdr.Checksum = dl.udpHdr.Checksum
	}

	return nil
}

func (l *UDPLayer) WriteLayer() error {

	fr := l.hdr.fr
	if fr == nil {
		return nil
	}

	data := fr.frame
	data.Append(l.udpHdr.SrcPort)
	data.Append(l.udpHdr.DstPort)
	data.Append(l.udpHdr.Length)
	data.Append(uint16(0)) // force checksum to zero, update later

	return nil
}
