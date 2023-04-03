/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation.
 */

package fserde

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// The Ether() protocol layer has a number of protocol-value options zero or more
// options may be specified.
//
// dst, src - are the destination and source MAC addresses.
// [proto|ethertype] - is the EtherType value.

type EtherHdr struct {
	DstMac, SrcMac net.HardwareAddr
	EtherType      uint16
}

type EtherLayer struct {
	hdr   *LayerHdr
	ether EtherHdr
}

func ToHardwareAddr(mac string) (net.HardwareAddr, error) {
	if strings.Count(mac, ":") == 2 { // Convert xxxx:xxxx:xxxx to xxxx.xxxx.xxxx
		mac = strings.ReplaceAll(mac, ":", ".")
	}

	if mac, err := net.ParseMAC(mac); err != nil {
		return nil, err
	} else {
		return mac, nil
	}
}

func (e *EtherLayer) String() string {
	return fmt.Sprintf("Ether(dst=%v, src=%v, proto=%04x)", e.ether.DstMac, e.ether.SrcMac, e.ether.EtherType)
}

func isZeroMac(mac net.HardwareAddr) bool {
	hw, _ := ToHardwareAddr("00:00:00:00:00:00")
	return len(mac) == 0 || bytes.Equal(hw, mac)
}

func (e *EtherLayer) Name() LayerName {
	return e.hdr.layerName
}

func EtherNew(fr *Frame) *EtherLayer {

	return &EtherLayer{
		hdr: LayerConstructor(fr, LayerEther, LayerEtherType),
	}
}

// parseEther a formatted string into a byte array or frame.
func (el *EtherLayer) Parse(opts string) error {

	options := strings.Split(opts, ",")

	for _, opt := range options {

		opt = strings.TrimSpace(opt)
		if len(opt) == 0 {
			continue
		}

		kvp := strings.Split(opt, "=")
		if len(kvp) != 2 {
			return fmt.Errorf("option needs a key/value pair")
		}
		key := strings.ToLower(strings.TrimSpace(kvp[0]))
		val := strings.ToLower(strings.TrimSpace(kvp[1]))

		switch key {
		case "dst":
			if mac, err := ToHardwareAddr(val); err != nil {
				return err
			} else {
				el.ether.DstMac = mac
			}

		case "src":
			if mac, err := ToHardwareAddr(val); err != nil {
				return err
			} else {
				el.ether.SrcMac = mac
			}

		case "proto", "ethertype":
			if v, err := strconv.ParseUint(val, 0, 16); err != nil {
				return err
			} else {
				el.ether.EtherType = uint16(v)
			}

		default:
			return fmt.Errorf("unknown ether option: [%s]", opt)
		}
	}

	if isZeroMac(el.ether.DstMac) {
		if dst, err := ToHardwareAddr("00:00:00:00:00:00"); err != nil {
			return err
		} else {
			el.ether.DstMac = dst
		}
	}

	if isZeroMac(el.ether.SrcMac) {
		if src, err := ToHardwareAddr("00:00:00:00:00:00"); err != nil {
			return err
		} else {
			el.ether.SrcMac = src
		}
	}

	el.hdr.proto.name = el.Name()
	el.hdr.proto.offset = el.hdr.fr.GetOffset(el.Name())
	el.hdr.proto.length = 14

	el.hdr.fr.AddProtocol(&el.hdr.proto)

	return nil
}

func (l *EtherLayer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	dl, ok := d.GetLayer(LayerEther).(*EtherLayer)
	if !ok {
		return nil
	}

	if isZeroMac(l.ether.DstMac) && !isZeroMac(dl.ether.DstMac) {
		l.ether.DstMac = dl.ether.DstMac
	}

	if isZeroMac(l.ether.SrcMac) && !isZeroMac(dl.ether.SrcMac) {
		l.ether.SrcMac = dl.ether.SrcMac
	}

	if l.ether.EtherType == 0 && dl.ether.EtherType > 0 {
		l.ether.EtherType = dl.ether.EtherType
	}

	return nil
}

func (l *EtherLayer) WriteLayer() error {

	data := l.hdr.fr.frame

	data.Append(l.ether.DstMac)
	data.Append(l.ether.SrcMac)
	data.Append(l.ether.EtherType)

	return nil
}
