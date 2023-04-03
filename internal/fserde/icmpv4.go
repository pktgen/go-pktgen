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
	ICMPv4MinLen = 8
	ICMPv6MinLen = 40
)

// ICMPHdr L4 header.
type ICMPHeader struct {
	Type       uint8  // Message type
	Code       uint8  // Message code
	Cksum      uint16 // Checksum set to zero
	Identifier uint16 // Message identifier in some messages
	SeqNum     uint16 // Message sequence number in some messages
}

type ICMPv4Layer struct {
	hdr     *LayerHdr
	icmpHdr ICMPHeader
}

func (l *ICMPv4Layer) String() string {
	return fmt.Sprintf("%s()", l.hdr.layerName)
}

func ICMPv4New(fr *Frame) *ICMPv4Layer {
	return &ICMPv4Layer{
		hdr: LayerConstructor(fr, LayerICMPv4, LayerICMPv4Type),
	}
}

func (l *ICMPv4Layer) Name() LayerName {
	return l.hdr.layerName
}

func (l *ICMPv4Layer) Parse(opts string) error {

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
		case "type":
			if v, err := strconv.ParseInt(val, 0, 0); err != nil {
				return err
			} else {
				l.icmpHdr.Type = uint8(v)
			}
		case "code":
			if v, err := strconv.ParseInt(val, 0, 0); err != nil {
				return err
			} else {
				l.icmpHdr.Code = uint8(v)
			}
		case "identifier", "ident":
			if v, err := strconv.ParseInt(val, 0, 0); err != nil {
				return err
			} else {
				l.icmpHdr.Identifier = uint16(v)
			}
		case "seq", "seqnum":
			if v, err := strconv.ParseInt(val, 0, 0); err != nil {
				return err
			} else {
				l.icmpHdr.SeqNum = uint16(v)
			}
		}
	}

	l.hdr.proto.name = l.Name()
	l.hdr.proto.offset = l.hdr.fr.GetOffset(l.Name())
	l.hdr.proto.length = 0

	l.hdr.fr.AddProtocol(&l.hdr.proto)

	return nil
}

func (l *ICMPv4Layer) ApplyDefaults() error {

	d := l.hdr.fr.defaultsFrame
	if d == nil {
		return nil
	}

	dl, ok := d.GetLayer(LayerICMPv4).(*ICMPv4Layer)
	if !ok {
		return nil
	}

	fmt.Printf("%s.ApplyDefaults: %T\n", l.Name(), dl)

	return nil
}

func (l *ICMPv4Layer) WriteLayer() error {

	fr := l.hdr.fr
	if fr == nil {
		return nil
	}

	data := fr.frame
	data.Append(l.icmpHdr.Type)
	data.Append(l.icmpHdr.Code)
	data.Append(l.icmpHdr.Cksum)
	data.Append(l.icmpHdr.Identifier)
	data.Append(l.icmpHdr.SeqNum)

	dbug.Printf("%v\n", l)
	return nil
}
