/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2022-2024 Intel Corporation.
 */

package fserde

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	Tag8021Q   = 0x8100
	DefaultVid = 0x0001
)

// The Dot1Q() protocol layer has a number of protocol-value options zero or more
// options may be specified.
//
// dst, src - are the destination and source MAC addresses.
// [proto|ethertype] - is the EtherType value.

type Dot1Q struct {
	tPid uint16 // Tag protocol ID
	tci  uint16 // Contains pcp 3bits, dei 1bit, vid 12bits
	pcp  uint16 // Part of tci
	dei  uint16 // Part of tci
	vid  uint16 // Part of tci
}

type Dot1qLayer struct {
	hdr   *LayerHdr
	dot1q Dot1Q
}

func (d *Dot1qLayer) String() string {
	return fmt.Sprintf("Dot1q(tpid=0x%04x, pcp=%v, dei=%v, vid=%v)", d.dot1q.tPid, (d.dot1q.tci >> 13),
		(d.dot1q.tci>>12)&0x1, d.dot1q.tci&0x0fff)
}

func (d *Dot1qLayer) Name() LayerName {
	return d.hdr.layerName
}

func Dot1qNew(fr *Frame) *Dot1qLayer {

	dot1q := &Dot1qLayer{
		hdr: LayerConstructor(fr, LayerDot1Q, LayerDot1QType),
	}
	return dot1q
}

func (d *Dot1qLayer) ParseDot1q(opts string) error {

	options := strings.Split(opts, ",")

	for _, opt := range options {
		opt = strings.TrimSpace(opt)

		kvp := strings.Split(opt, "=")
		if len(kvp) != 2 {
			return fmt.Errorf("invalid options string")
		}
		key := strings.ToLower(strings.TrimSpace(kvp[0]))
		val := strings.ToLower(strings.TrimSpace(kvp[1]))

		switch key {
		case "tpid":
			if v, err := strconv.ParseUint(val, 0, 16); err != nil {
				return err
			} else {
				d.dot1q.tPid = uint16(v)
			}
		case "prio", "pcp":
			if v, err := strconv.ParseUint(val, 0, 0); err != nil {
				return err
			} else {
				d.dot1q.pcp = uint16(v) & 0x07
			}
		case "cfi", "dei":
			if v, err := strconv.ParseUint(val, 0, 0); err != nil {
				if strings.EqualFold(val, "true") {
					d.dot1q.dei = 1
				} else if strings.EqualFold(val, "false") {
					d.dot1q.dei = 0
				} else {
					return fmt.Errorf("invalid boolean value: %s", val)
				}
			} else {
				d.dot1q.dei = uint16(v) & 1
			}
		case "vlan", "vid":
			if v, err := strconv.ParseUint(val, 0, 0); err != nil {
				return err
			} else {
				d.dot1q.vid = (uint16(v) & 0x0FFF)
			}
		case "tci": // when tci is set then vid, pcp and dei are ignored
			if v, err := strconv.ParseUint(val, 0, 0); err != nil {
				return err
			} else {
				d.dot1q.tci = (uint16(v) & 0xFFFF)
			}
		}
	}

	if d.dot1q.tPid == 0 {
		d.dot1q.tPid = Tag8021Q
	}
	if d.dot1q.tci == 0 {
		if d.dot1q.vid == 0 {
			d.dot1q.vid = DefaultVid
		}
		d.dot1q.tci = uint16(d.dot1q.pcp<<13 | d.dot1q.dei<<12 | d.dot1q.vid)
	}

	dbug.Printf("%v\n", d)

	return nil
}

func (d *Dot1qLayer) Parse(opts string) error {

	if err := d.ParseDot1q(opts); err != nil {
		return err
	}

	d.hdr.proto.name = d.Name()
	d.hdr.proto.offset = d.hdr.fr.GetOffset(d.Name())
	d.hdr.proto.length = 4

	d.hdr.fr.AddProtocol(&d.hdr.proto)

	dbug.Printf("%v\n", d)

	return nil
}

func (l *Dot1qLayer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	dbug.Printf("%v\n", d)

	dl, ok := d.GetLayer(LayerDot1Q).(*Dot1qLayer)
	if !ok {
		return nil
	}

	dbug.Printf("%v\n", dl)

	if l.dot1q.tPid == 0 && l.dot1q.tPid != 0 {
		l.dot1q.tPid = dl.dot1q.tPid
	}
	if l.dot1q.tci == 0 && l.dot1q.tci != 0 {
		l.dot1q.tci = dl.dot1q.tci
	}

	return nil
}

func (l *Dot1qLayer) WriteLayer() error {

	data := l.hdr.fr.frame

	dstSrcMAC := make([]byte, 12)
	copy(dstSrcMAC, data.Bytes()[:12])
	restData := make([]byte, data.Len()-12)
	copy(restData, data.Bytes()[12:])

	data.Reset()
	data.Append(dstSrcMAC)
	data.Append(l.dot1q.tPid)
	data.Append(l.dot1q.tci)
	data.Append(restData)

	dbug.Printf("%v\n", data.Bytes())

	return nil
}
